package main

import (
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/db"
	"backend/service-platform/app/pkg/logging"
	"backend/service-platform/app/pkg/redis"
	ctxutil "backend/service-platform/app/pkg/util/context"
	httpClientUtil "backend/service-platform/app/pkg/util/httpclient"
	server "backend/service-platform/app/worker"
	"context"
	"time"

	"go.uber.org/zap"
)

func main() {
	env := ctxutil.GetAppModeFromEnv()
	ctx := ctxutil.SetAppMode(context.Background(), env)

	// Configure log
	logConfig := logging.NewLogConfig("[service-platform]", env)
	logger, err := logConfig.NewLogging()
	if err != nil {
		panic(err)
	}
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)
	zap.ReplaceGlobals(logger)

	// Load ENV variables
	cfg, err := config.ReadApplicationConfig(env, logger)
	if err != nil {
		logger.Error("Failed to load APP configuration", zap.Error(err))
	}

	// Configure database
	database, err := db.NewDB(cfg, logger)
	if err != nil {
		panic(err)
	}
	defer func(database *db.DB) {
		err = database.Close()
		if err != nil {
			logger.Error("Failed to close Database connection", zap.Error(err))
		}
		logger.Info("Failed to closed Database connection")
	}(database)

	// Configure Redis
	rds, err := redis.NewRedisClusterClient(cfg.RedisConfig, logger)
	if err != nil {
		panic(err)
	}
	defer func(rds redis.Redis) {
		if err := rds.Close(); err != nil {
			logger.Error("Failed to close Redis connection", zap.Error(err))
		} else {
			logger.Info("Failed to closed Redis connection")
		}
	}(rds)

	// Configure HttpClient
	httpClient := httpClientUtil.NewRestyClient(30*time.Second, logger)

	workerServer := server.Server{
		Config:     cfg,
		Logger:     logger,
		DB:         database,
		Redis:      rds,
		HttpClient: httpClient,
		Clients:    runtime.Clients{},
	}
	workerServer.Start(ctx)
}
