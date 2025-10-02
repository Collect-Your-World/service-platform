package service

import (
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/queue"
	"backend/service-platform/app/pkg/worker"
	"backend/service-platform/app/pkg/worker/handlers"
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WorkerService manages worker pools and background processes
type WorkerService struct {
	workerPool   worker.Pool
	jobRepo      repository.JobRepository
	queue        queue.Queue
	logger       *zap.Logger
	workerConfig config.WorkerConfig
}

// NewWorkerService creates a new worker service with all necessary components
func NewWorkerService(res runtime.Resource, workerConfig config.WorkerConfig) *WorkerService {
	logger := res.Logger.With(zap.String("component", "worker_service"))

	// Create a job repository
	jobRepo := repository.NewJobRepository(res)

	// Create Redis queue
	redisQueue := queue.NewRedisQueue(res.Redis.GetUniversalClient(), logger)

	// Create handler registry and register handlers
	handlerRegistry := worker.NewJobHandlerRegistry(logger)
	handlerRegistry.Register(handlers.NewInitClaimHandler(logger))
	handlerRegistry.Register(handlers.NewCompleteClaimHandler(logger))
	handlerRegistry.Register(handlers.NewKYCVerificationHandler(logger))

	// Create a worker pool
	workerPool := worker.NewWorkerPool(
		workerConfig.PoolSize,
		redisQueue,
		jobRepo,
		handlerRegistry,
		logger,
	)

	return &WorkerService{
		workerPool:   workerPool,
		jobRepo:      jobRepo,
		queue:        redisQueue,
		logger:       logger,
		workerConfig: workerConfig,
	}
}

// Start starts all worker processes
func (ws *WorkerService) Start(ctx context.Context) error {
	ws.logger.Info("Starting worker service", zap.Int("pool_size", ws.workerConfig.PoolSize))

	var wg sync.WaitGroup

	// Start a worker pool
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ws.workerPool.Start(ctx); err != nil {
			ws.logger.Error("Failed to start worker pool", zap.Error(err))
		}
	}()

	// Start retry scheduler
	wg.Add(1)
	go func() {
		defer wg.Done()
		ws.runRetryScheduler(ctx)
	}()

	// Start health monitor
	wg.Add(1)
	go func() {
		defer wg.Done()
		ws.runHealthMonitor(ctx)
	}()

	ws.logger.Info("Worker service started successfully")

	// Wait for context cancellation
	<-ctx.Done()
	ws.logger.Info("Shutting down worker service")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ws.workerPool.Stop(shutdownCtx); err != nil {
		ws.logger.Error("Failed to stop worker pool gracefully", zap.Error(err))
		return err
	}

	wg.Wait()
	ws.logger.Info("Worker service stopped")
	return nil
}

// Stop gracefully stops all worker processes
func (ws *WorkerService) Stop(ctx context.Context) error {
	return ws.workerPool.Stop(ctx)
}

// GetStats returns worker pool statistics
func (ws *WorkerService) GetStats() worker.PoolStats {
	return ws.workerPool.GetStats()
}

// GetWorkerPool returns the underlying worker pool
func (ws *WorkerService) GetWorkerPool() worker.Pool {
	return ws.workerPool
}

func (ws *WorkerService) runRetryScheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	retryLogger := ws.logger.With(zap.String("component", "retry_scheduler"))
	retryLogger.Info("Starting retry scheduler")

	for {
		select {
		case <-ctx.Done():
			retryLogger.Info("Retry scheduler stopping")
			return
		case <-ticker.C:
			ws.processRetryableJobs(ctx, retryLogger)
		}
	}
}

func (ws *WorkerService) processRetryableJobs(ctx context.Context, logger *zap.Logger) {
	retryTime := time.Now().Add(-5 * time.Minute)
	jobs, err := ws.jobRepo.GetRetryableJobs(ctx, retryTime, 100)
	if err != nil {
		logger.Error("Failed to get retryable jobs", zap.Error(err))
		return
	}

	if len(jobs) == 0 {
		return
	}

	logger.Info("Found retryable jobs", zap.Int("count", len(jobs)))

	for _, job := range jobs {
		if err := ws.jobRepo.UpdateStatus(ctx, job.ID, "pending", ""); err != nil {
			logger.Error("Failed to update job status for retry",
				zap.String("job_id", job.ID),
				zap.Error(err))
			continue
		}

		if err := ws.queue.Enqueue(ctx, job); err != nil {
			logger.Error("Failed to re-enqueue job",
				zap.String("job_id", job.ID),
				zap.Error(err))
			continue
		}

		logger.Info("Job re-queued for retry", zap.String("job_id", job.ID))
	}
}

func (ws *WorkerService) runHealthMonitor(ctx context.Context) {
	interval := ws.workerConfig.HealthMonitorInterval
	ws.logger.Info(fmt.Sprintf("Health Monitor interval: %s", interval))
	if interval <= 0 {
		interval = 1 * time.Minute
		ws.logger.Warn("Invalid health monitor interval, using default",
			zap.Duration("configured", ws.workerConfig.HealthMonitorInterval),
			zap.Duration("default", interval))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	healthLogger := ws.logger.With(zap.String("component", "worker_service"))
	healthLogger.Info("Starting health monitor", zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			healthLogger.Info("Health monitor stopping")
			return
		case <-ticker.C:
			stats := ws.workerPool.GetStats()
			healthLogger.Info("Worker pool stats",
				zap.Int("active_workers", stats.ActiveWorkers),
				zap.Int("processing_jobs", stats.ProcessingJobs),
				zap.Int64("total_processed", stats.TotalProcessed),
				zap.Int64("total_failed", stats.TotalFailed),
				zap.Any("queue_depths", stats.QueueDepths))
		}
	}
}
