package integration

import (
	service "backend/service-platform/app/service"
	"context"
	"database/sql"
	"time"

	"backend/service-platform/app/api/controller"
	"backend/service-platform/app/api/middleware"
	"backend/service-platform/app/internal/validator"
	"backend/service-platform/app/manager"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"backend/service-platform/app/api/router"
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/db"
	"backend/service-platform/app/pkg/logging"
	"backend/service-platform/app/pkg/redis"
	ctxutil "backend/service-platform/app/pkg/util/context"
)

type RouterSuite struct {
	suite.Suite
	resource     runtime.Resource
	r            *require.Assertions
	a            *assert.Assertions
	e            *echo.Echo
	ctx          context.Context
	repositories *repository.Repositories
	services     *service.Services
	managers     *manager.Managers
	suiteSetupAt time.Time
	testSetupAt  time.Time
}

func (s *RouterSuite) SetupSuite() {
	s.r = s.Suite.Require()
	s.a = s.Suite.Assert()
	env := ctxutil.AppMode("test")
	s.ctx = ctxutil.SetAppMode(context.Background(), env)

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

	// Connect to redis for tests
	rds, err := redis.NewRedisClusterClient(cfg.RedisConfig, logger)
	if err != nil {
		// In test environment, if cluster fails, we might be using single-node Redis
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
	s.resource = res

	repositories := repository.NewRepositories(res)
	s.repositories = repositories

	// Use worker config from resource
	workerConfig := res.Config.WorkerConfig

	// Create managers first to get the JobManager
	tempManagers := manager.NewManagers(res, nil, repositories)

	// Try to create services with job manager to enable SQS listener
	// Fall back to basic services if SQS is not available (e.g., LocalStack not running)
	var services *service.Services
	func() {
		defer func() {
			if r := recover(); r != nil {
				res.Logger.Warn("SQS service creation failed, falling back to basic services", zap.Any("error", r))
				services = service.NewServices(res, workerConfig)
			}
		}()
		services = service.NewServicesWithJobManager(res, workerConfig, tempManagers.JobManager)
	}()
	s.services = services

	// Create final managers with services
	managers := manager.NewManagers(res, services, repositories)
	s.managers = managers

	controllers := controller.NewControllers(managers, res)
	validators := validator.NewValidators(res)

	middlewares := middleware.NewMiddleware(res)
	s.e = router.NewRouter(res, validators, middlewares, controllers, repositories).Echo
	s.suiteSetupAt = s.startSuiteTimestamp()
}

func (s *RouterSuite) TearDownSuite() {
	s.resource.Logger.Info("Starting integration test cleanup")

	// Clean up all test data from database
	if err := s.cleanAllTestData(); err != nil {
		s.resource.Logger.Error("Failed to clean database in test teardown", zap.Error(err))
	}

	// Clean up Redis
	if s.resource.Redis != nil {
		s.cleanRedis()
		if err := s.resource.Redis.Close(); err != nil {
			s.resource.Logger.Error("Failed to close Redis connection in test teardown", zap.Error(err))
		}
	}

	// Close database connections
	if s.resource.DB != nil {
		if err := s.resource.DB.Close(); err != nil {
			s.resource.Logger.Error("Failed to close database connection in test teardown", zap.Error(err))
		}
	}

	s.resource.Logger.Info("Integration test cleanup completed")
}

func (s *RouterSuite) SetupTest() {
	s.testSetupAt = s.startSuiteTimestamp()
	if err := s.cleanDBAt(s.testSetupAt); err != nil {
		s.T().Fatal(err)
	}
}

func (s *RouterSuite) TearDownTest() {
	if err := s.cleanDBAt(s.testSetupAt); err != nil {
		s.T().Fatal(err)
	}
}

func (s *RouterSuite) startSuiteTimestamp() time.Time {
	var timestamp time.Time
	err := s.resource.DB.PrimaryDb.QueryRow("SELECT NOW()").Scan(&timestamp)
	s.r.NoError(err)
	now := time.Now()
	if timestamp.After(time.Now()) {
		timestamp = now
	}
	return timestamp
}

func (s *RouterSuite) cleanDBAt(timestamp time.Time) error {
	s.resource.Logger.Debug("clean test db")
	rows, err := s.resource.DB.PrimaryDb.QueryContext(s.ctx, `
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
			s.resource.Logger.Error(err.Error())
		}
	}(rows)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return err
		}
		_, err = s.resource.DB.PrimaryConn().NewDelete().Table(table).Where("created_at > ?", timestamp).Exec(s.ctx)
		if err != nil {
			s.resource.Logger.Error("failed to clean db", zap.String("table", table), zap.Error(err))
		} else {
			s.resource.Logger.Debug("cleaned db", zap.String("table", table), zap.Time("created_at > ", timestamp))
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

// cleanAllTestData removes all test data from the database
func (s *RouterSuite) cleanAllTestData() error {
	s.resource.Logger.Info("Cleaning all test data from database")

	// Get all tables with created_at column
	rows, err := s.resource.DB.PrimaryDb.QueryContext(s.ctx, `
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

	// Clean each table
	for _, table := range tables {
		_, err := s.resource.DB.PrimaryConn().NewDelete().Table(table).Where("created_at >= ?", s.suiteSetupAt).Exec(s.ctx)
		if err != nil {
			s.resource.Logger.Error("Failed to clean table", zap.String("table", table), zap.Error(err))
		} else {
			s.resource.Logger.Debug("Cleaned table", zap.String("table", table))
		}
	}

	return nil
}

// cleanRedis flushes all Redis data
func (s *RouterSuite) cleanRedis() {
	s.resource.Logger.Info("Flushing all Redis data")

	ctx := context.Background()

	// Flush all data from Redis
	if err := s.resource.Redis.Reset(ctx); err != nil {
		s.resource.Logger.Error("Failed to flush Redis", zap.Error(err))
	} else {
		s.resource.Logger.Info("Successfully flushed all Redis data")
	}
}
