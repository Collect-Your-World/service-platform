package main

import (
	"backend/service-platform/app/pkg/aws"
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	server "backend/service-platform/app/api"
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/db"
	"backend/service-platform/app/pkg/logging"
	"backend/service-platform/app/pkg/redis"
	ctxutil "backend/service-platform/app/pkg/util/context"
	httpClientUtil "backend/service-platform/app/pkg/util/httpclient"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	env := ctxutil.GetAppModeFromEnv()
	ctx := ctxutil.SetAppMode(context.Background(), env)

	logger := setupLogging(env)
	defer func() {
		_ = logger.Sync()
	}()

	cfg := loadConfiguration(env, logger)
	database := setupDatabase(cfg, logger)
	defer closeDatabase(database, logger)

	redisClient := setupRedis(cfg, logger)
	defer closeRedis(redisClient, logger)

	externalClients := setupExternalClients(ctx, cfg, logger)

	httpServer := createServer(cfg, logger, database, redisClient, externalClients)
	httpServer.Start(ctx)
}

func setupLogging(env ctxutil.AppMode) *zap.Logger {
	logConfig := logging.NewLogConfig("[service-platform]", env)
	logger, err := logConfig.NewLogging()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	return logger
}

func loadConfiguration(env ctxutil.AppMode, logger *zap.Logger) config.ApplicationConfig {
	cfg, err := config.ReadApplicationConfig(env, logger)
	if err != nil {
		panic(err)
	}
	return cfg
}

func setupDatabase(cfg config.ApplicationConfig, logger *zap.Logger) *db.DB {
	database, err := db.NewDB(cfg, logger)
	if err != nil {
		panic(err)
	}
	return database
}

func closeDatabase(database *db.DB, logger *zap.Logger) {
	if err := database.Close(); err != nil {
		logger.Error("error closing database", zap.Error(err))
	} else {
		logger.Info("closed database connection")
	}
}

func setupRedis(cfg config.ApplicationConfig, logger *zap.Logger) redis.Redis {
	redisClient, err := redis.NewRedisClusterClient(cfg.RedisConfig, logger)
	if err != nil {
		panic(err)
	}
	return redisClient
}

func closeRedis(redisClient redis.Redis, logger *zap.Logger) {
	if err := redisClient.Close(); err != nil {
		logger.Error("error closing redis connection", zap.Error(err))
	} else {
		logger.Info("closed redis connection")
	}
}

type ExternalClients struct {
	HttpClient *resty.Client
	SqsClient  *sqs.Client
}

func setupExternalClients(ctx context.Context, cfg config.ApplicationConfig, logger *zap.Logger) ExternalClients {
	httpClient := httpClientUtil.NewRestyClient(30*time.Second, logger)

	sqsClient, err := aws.NewSQSClient(ctx, cfg)
	if err != nil {
		logger.Error("Failed to create SQS client", zap.Error(err))
	}

	return ExternalClients{
		HttpClient: httpClient,
		SqsClient:  sqsClient,
	}
}

func createServer(cfg config.ApplicationConfig, logger *zap.Logger, database *db.DB, redisClient redis.Redis, clients ExternalClients) server.Server {
	return server.Server{
		Config:     cfg,
		Logger:     logger,
		DB:         database,
		Redis:      redisClient,
		HttpClient: clients.HttpClient,
		SqsClient:  clients.SqsClient,
		Clients:    runtime.Clients{},
	}
}
