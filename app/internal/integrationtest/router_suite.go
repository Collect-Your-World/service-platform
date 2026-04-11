package integrationtest

import (
	"context"
	"database/sql"
	"time"

	appmanagers "backend/service-platform/app/internal/managers"
	platctrl "backend/service-platform/app/internal/platform/controllers"
	echomw "backend/service-platform/app/internal/platform/middleware"
	"backend/service-platform/app/internal/platform/router"
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/internal/platform/validator"
	"backend/service-platform/app/internal/repository"
	"backend/service-platform/app/internal/testutil"
	worker "backend/service-platform/app/internal/worker"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/pkg/db"
	"backend/service-platform/app/pkg/logging"
	"backend/service-platform/app/pkg/redis"
	ctxutil "backend/service-platform/app/pkg/util/context"
)

type RouterSuite struct {
	suite.Suite
	Resource     runtime.Resource
	R            *require.Assertions
	A            *assert.Assertions
	Echo         *echo.Echo
	Ctx          context.Context
	Repositories *repository.Repositories
	Services     *worker.Services
	Managers     *appmanagers.Managers
	SuiteSetupAt time.Time
	TestSetupAt  time.Time
}

func (s *RouterSuite) SetupSuite() {
	s.R = s.Suite.Require()
	s.A = s.Suite.Assert()
	env := ctxutil.AppMode("test")
	s.Ctx = ctxutil.SetAppMode(context.Background(), env)

	logConfig := logging.NewLogConfig("[service-platform]", env)
	logger, err := logConfig.NewLogging()
	if err != nil {
		panic(err)
	}
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)
	zap.ReplaceGlobals(logger)

	cfg, err := config.ReadApplicationConfig(env, logger)
	if err != nil {
		panic(err)
	}

	database, err := db.NewDB(cfg, logger)
	if err != nil {
		panic(err)
	}

	rds, err := redis.NewRedisClusterClient(cfg.RedisConfig, logger)
	if err != nil {
		logger.Warn("Redis cluster connection failed, this might be expected in test environment", zap.Error(err))
		panic(err)
	}

	res := runtime.Resource{
		Config:  cfg,
		DB:      database,
		Redis:   rds,
		Logger:  logger,
		Clients: runtime.Clients{},
	}
	s.Resource = res

	repositories := repository.NewRepositories(res)
	s.Repositories = repositories

	workerConfig := res.Config.WorkerConfig

	tempManagers := appmanagers.NewManagers(res, nil, repositories)

	var services *worker.Services
	func() {
		defer func() {
			if r := recover(); r != nil {
				res.Logger.Warn("SQS service creation failed, falling back to basic services", zap.Any("error", r))
				services = worker.NewServices(res, workerConfig)
			}
		}()
		services = worker.NewServicesWithJobManager(res, workerConfig, tempManagers.JobManager)
	}()
	s.Services = services

	managers := appmanagers.NewManagers(res, nil, repositories)
	s.Managers = managers

	controllers := platctrl.NewControllers(managers, res)
	validators := validator.NewValidators(res)

	middlewares := echomw.NewMiddleware(res)
	s.Echo = router.NewRouter(res, validators, middlewares, controllers, repositories).Echo
	s.SuiteSetupAt = s.startSuiteTimestamp()
}

func (s *RouterSuite) TearDownSuite() {
	s.Resource.Logger.Info("Starting integration test cleanup")

	if err := s.cleanAllTestData(); err != nil {
		s.Resource.Logger.Error("Failed to clean database in test teardown", zap.Error(err))
	}

	if s.Resource.Redis != nil {
		s.cleanRedis()
		if err := s.Resource.Redis.Close(); err != nil {
			s.Resource.Logger.Error("Failed to close Redis connection in test teardown", zap.Error(err))
		}
	}

	if s.Resource.DB != nil {
		if err := s.Resource.DB.Close(); err != nil {
			s.Resource.Logger.Error("Failed to close database connection in test teardown", zap.Error(err))
		}
	}

	s.Resource.Logger.Info("Integration test cleanup completed")
}

func (s *RouterSuite) SetupTest() {
	s.TestSetupAt = s.startSuiteTimestamp()
	if err := s.cleanDBAt(s.TestSetupAt); err != nil {
		s.T().Fatal(err)
	}
	testutil.ClearCookies()
}

func (s *RouterSuite) TearDownTest() {
	if err := s.cleanDBAt(s.TestSetupAt); err != nil {
		s.T().Fatal(err)
	}
}

func (s *RouterSuite) startSuiteTimestamp() time.Time {
	var timestamp time.Time
	err := s.Resource.DB.PrimaryDb.QueryRow("SELECT NOW()").Scan(&timestamp)
	s.R.NoError(err)
	now := time.Now()
	if timestamp.After(time.Now()) {
		timestamp = now
	}
	return timestamp
}

func (s *RouterSuite) cleanDBAt(timestamp time.Time) error {
	s.Resource.Logger.Debug("clean test db")
	rows, err := s.Resource.DB.PrimaryDb.QueryContext(s.Ctx, `
	SELECT t.table_name
	FROM information_schema.tables t
	JOIN information_schema.columns c
		ON t.table_name = c.table_name
		AND t.table_schema = c.table_schema
	WHERE t.table_schema = 'public'
		AND c.column_name = 'created_at'
	`)
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			s.Resource.Logger.Error(err.Error())
		}
	}(rows)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return err
		}
		_, err = s.Resource.DB.PrimaryConn().NewDelete().Table(table).Where("created_at > ?", timestamp).Exec(s.Ctx)
		if err != nil {
			s.Resource.Logger.Error("failed to clean db", zap.String("table", table), zap.Error(err))
		} else {
			s.Resource.Logger.Debug("cleaned db", zap.String("table", table), zap.Time("created_at > ", timestamp))
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func (s *RouterSuite) cleanAllTestData() error {
	s.Resource.Logger.Info("Cleaning all test data from database")

	rows, err := s.Resource.DB.PrimaryDb.QueryContext(s.Ctx, `
	SELECT t.table_name
	FROM information_schema.tables t
	JOIN information_schema.columns c
		ON t.table_name = c.table_name
		AND t.table_schema = c.table_schema
	WHERE t.table_schema = 'public'
		AND c.column_name = 'created_at'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return err
		}
		tables = append(tables, table)
	}

	for _, table := range tables {
		_, err := s.Resource.DB.PrimaryConn().NewDelete().Table(table).Where("created_at >= ?", s.SuiteSetupAt).Exec(s.Ctx)
		if err != nil {
			s.Resource.Logger.Error("Failed to clean table", zap.String("table", table), zap.Error(err))
		} else {
			s.Resource.Logger.Debug("Cleaned table", zap.String("table", table))
		}
	}

	return nil
}

func (s *RouterSuite) cleanRedis() {
	s.Resource.Logger.Info("Flushing all Redis data")

	ctx := context.Background()

	if err := s.Resource.Redis.Reset(ctx); err != nil {
		s.Resource.Logger.Error("Failed to flush Redis", zap.Error(err))
	} else {
		s.Resource.Logger.Info("Successfully flushed all Redis data")
	}
}
