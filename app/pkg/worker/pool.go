package worker

import (
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/pkg/queue"
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Pool interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetStats() PoolStats
}

type PoolStats struct {
	TotalWorkers   int              `json:"total_workers"`
	ActiveWorkers  int              `json:"active_workers"`
	ProcessingJobs int              `json:"processing_jobs"`
	TotalProcessed int64            `json:"total_processed"`
	TotalFailed    int64            `json:"total_failed"`
	QueueDepths    map[string]int64 `json:"queue_depths"`
}

type workerPool struct {
	workers         int
	queue           queue.Queue
	jobRepo         repository.JobRepository
	handlerRegistry JobHandlerRegistry
	logger          *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	stats      PoolStats
	statsMutex sync.RWMutex
}

func NewWorkerPool(
	workers int,
	queue queue.Queue,
	jobRepo repository.JobRepository,
	handlerRegistry JobHandlerRegistry,
	logger *zap.Logger,
) Pool {
	return &workerPool{
		workers:         workers,
		queue:           queue,
		jobRepo:         jobRepo,
		handlerRegistry: handlerRegistry,
		logger:          logger,
		stats: PoolStats{
			QueueDepths: make(map[string]int64),
		},
	}
}

func (p *workerPool) Start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)

	p.logger.Info("Starting worker pool", zap.Int("workers", p.workers))

	// Set the total workers count
	p.statsMutex.Lock()
	p.stats.TotalWorkers = p.workers
	p.statsMutex.Unlock()

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.runWorker(i)
	}

	p.wg.Add(1)
	go p.runStatsCollector()

	return nil
}

func (p *workerPool) Stop(ctx context.Context) error {
	p.logger.Info("Stopping worker pool")

	if p.cancel != nil {
		p.cancel()
	}

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("Worker pool stopped successfully")
		return nil
	case <-ctx.Done():
		p.logger.Warn("Worker pool stop timeout")
		return ctx.Err()
	}
}

func (p *workerPool) GetStats() PoolStats {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()

	stats := p.stats
	stats.QueueDepths = make(map[string]int64)
	for k, v := range p.stats.QueueDepths {
		stats.QueueDepths[k] = v
	}

	return stats
}

func (p *workerPool) runWorker(workerID int) {
	defer p.wg.Done()

	logger := p.logger.With(zap.Int("worker_id", workerID))
	logger.Info("Worker started")

	queues := queue.GetPriorityQueues()

	for {
		select {
		case <-p.ctx.Done():
			logger.Info("Worker stopping")
			return
		default:
			job, err := p.queue.Dequeue(p.ctx, queues)
			if err != nil {
				if p.ctx.Err() != nil {
					return
				}
				logger.Error("Failed to dequeue job", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			if job == nil {
				continue
			}

			p.incrementActiveWorkers()
			p.processJob(logger, job)
			p.decrementActiveWorkers()
		}
	}
}

func (p *workerPool) processJob(logger *zap.Logger, jobEntity *entity.Job) {
	jobLogger := logger.With(
		zap.String("job_id", jobEntity.ID),
		zap.String("job_type", jobEntity.Type),
		zap.String("priority", jobEntity.Priority.String()),
	)

	jobLogger.Info("Processing job")

	// Use background context for database operations to avoid cancellation during shutdown
	opCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := p.queue.MarkProcessing(opCtx, jobEntity.ID); err != nil {
		jobLogger.Error("Failed to mark job as processing", zap.Error(err))
		return
	}

	startTime := time.Now()
	if err := p.jobRepo.UpdateJobToProcessing(opCtx, jobEntity.ID, startTime); err != nil {
		jobLogger.Error("Failed to update job to processing state", zap.Error(err))
	}

	handler, exists := p.handlerRegistry.Get(jobEntity.Type)
	if !exists {
		err := fmt.Errorf("no handler found for job type: %s", jobEntity.Type)
		p.handleJobFailure(jobLogger, jobEntity, err)
		return
	}

	if err := handler.Handle(p.ctx, jobEntity); err != nil {
		p.handleJobFailure(jobLogger, jobEntity, err)
		return
	}

	p.handleJobSuccess(jobLogger, jobEntity)
}

func (p *workerPool) handleJobSuccess(logger *zap.Logger, jobEntity *entity.Job) {
	completedAt := time.Now()

	// Use background context for cleanup operations to avoid cancellation during shutdown
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := p.jobRepo.UpdateJobToCompleted(cleanupCtx, jobEntity.ID, completedAt); err != nil {
		logger.Error("Failed to update job to completed state", zap.Error(err))
	}

	if err := p.queue.MarkCompleted(cleanupCtx, jobEntity.ID); err != nil {
		logger.Error("Failed to mark job as completed in queue", zap.Error(err))
	}

	p.incrementTotalProcessed()
	logger.Info("Job completed successfully")
}

func (p *workerPool) handleJobFailure(logger *zap.Logger, jobEntity *entity.Job, jobErr error) {
	logger.Error("Job failed", zap.Error(jobErr))

	jobEntity.Attempts++

	// Use background context for cleanup operations to avoid cancellation during shutdown
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if jobEntity.Attempts >= jobEntity.MaxAttempts {
		if err := p.jobRepo.UpdateJobToFailed(cleanupCtx, jobEntity.ID, jobErr.Error()); err != nil {
			logger.Error("Failed to update job to failed state", zap.Error(err))
		}

		if err := p.queue.MarkFailed(cleanupCtx, jobEntity.ID, 0); err != nil {
			logger.Error("Failed to mark job as failed in queue", zap.Error(err))
		}

		logger.Info("Job permanently failed", zap.Int("attempts", jobEntity.Attempts))
	} else {
		retryDelay := p.calculateRetryDelay(jobEntity.Attempts)

		if err := p.jobRepo.UpdateJobToRetrying(cleanupCtx, jobEntity.ID, jobErr.Error()); err != nil {
			logger.Error("Failed to update job to retrying state", zap.Error(err))
		}

		if err := p.queue.MarkFailed(cleanupCtx, jobEntity.ID, retryDelay); err != nil {
			logger.Error("Failed to mark job for retry in queue", zap.Error(err))
		}

		logger.Info("Job scheduled for retry",
			zap.Int("attempts", jobEntity.Attempts),
			zap.Duration("retry_delay", retryDelay))
	}

	p.incrementTotalFailed()
}

func (p *workerPool) calculateRetryDelay(attempts int) time.Duration {
	baseDelay := 30 * time.Second
	backoff := time.Duration(math.Pow(2, float64(attempts-1))) * baseDelay

	maxDelay := 10 * time.Minute
	if backoff > maxDelay {
		backoff = maxDelay
	}

	return backoff
}

func (p *workerPool) runStatsCollector() {
	defer p.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.collectQueueStats()
		}
	}
}

func (p *workerPool) collectQueueStats() {
	queues := queue.GetPriorityQueues()

	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()

	for _, queueName := range queues {
		depth, err := p.queue.GetQueueDepth(p.ctx, queueName)
		if err != nil {
			p.logger.Error("Failed to get queue depth", zap.String("queue", queueName), zap.Error(err))
			continue
		}
		p.stats.QueueDepths[queueName] = depth
	}
}

func (p *workerPool) incrementActiveWorkers() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	p.stats.ActiveWorkers++
	p.stats.ProcessingJobs++
}

func (p *workerPool) decrementActiveWorkers() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	if p.stats.ActiveWorkers > 0 {
		p.stats.ActiveWorkers--
	}
	if p.stats.ProcessingJobs > 0 {
		p.stats.ProcessingJobs--
	}
}

func (p *workerPool) incrementTotalProcessed() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	p.stats.TotalProcessed++
}

func (p *workerPool) incrementTotalFailed() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	p.stats.TotalFailed++
}
