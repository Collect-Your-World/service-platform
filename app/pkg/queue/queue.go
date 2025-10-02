package queue

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	QueueKey      = "{jobs}:queue"
	QueueKeyRetry = "{jobs}:retry"

	ProcessingSetKey = "{jobs}:processing"
	RetrySetKey      = "{jobs}:retry_schedule"
)

type Queue interface {
	Enqueue(ctx context.Context, job *entity.Job) error
	Dequeue(ctx context.Context, queues []string) (*entity.Job, error)
	MarkProcessing(ctx context.Context, jobID string) error
	MarkCompleted(ctx context.Context, jobID string) error
	MarkFailed(ctx context.Context, jobID string, retryDelay time.Duration) error
	GetQueueDepth(ctx context.Context, queue string) (int64, error)
	GetProcessingJobs(ctx context.Context) ([]string, error)
}

type redisQueue struct {
	client redis.UniversalClient
	logger *zap.Logger
}

func NewRedisQueue(client redis.UniversalClient, logger *zap.Logger) Queue {
	return &redisQueue{
		client: client,
		logger: logger,
	}
}

func (q *redisQueue) Enqueue(ctx context.Context, job *entity.Job) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	queueKey := q.getQueueKey(job.Priority)

	pipe := q.client.TxPipeline()
	pipe.LPush(ctx, queueKey, jobData)

	if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
		score := float64(job.ScheduledAt.Unix())
		pipe.ZAdd(ctx, RetrySetKey, redis.Z{
			Score:  score,
			Member: job.ID,
		})
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		q.logger.Error("Failed to enqueue job", zap.String("job_id", job.ID), zap.Error(err))
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	q.logger.Info("Job enqueued successfully",
		zap.String("job_id", job.ID),
		zap.String("type", job.Type),
		zap.String("priority", job.Priority.String()),
		zap.String("queue", queueKey))

	return nil
}

func (q *redisQueue) Dequeue(ctx context.Context, queues []string) (*entity.Job, error) {
	result, err := q.client.BRPop(ctx, 5*time.Second, queues...).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) != 2 {
		return nil, fmt.Errorf("unexpected result format from BRPOP")
	}

	var job entity.Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		q.logger.Error("Failed to unmarshal job", zap.String("data", result[1]), zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	q.logger.Info("Job dequeued successfully",
		zap.String("job_id", job.ID),
		zap.String("type", job.Type),
		zap.String("queue", result[0]))

	return &job, nil
}

func (q *redisQueue) MarkProcessing(ctx context.Context, jobID string) error {
	timestamp := time.Now().Unix()
	err := q.client.ZAdd(ctx, ProcessingSetKey, redis.Z{
		Score:  float64(timestamp),
		Member: jobID,
	}).Err()

	if err != nil {
		q.logger.Error("Failed to mark job as processing", zap.String("job_id", jobID), zap.Error(err))
		return fmt.Errorf("failed to mark job as processing: %w", err)
	}

	return nil
}

func (q *redisQueue) MarkCompleted(ctx context.Context, jobID string) error {
	err := q.client.ZRem(ctx, ProcessingSetKey, jobID).Err()
	if err != nil {
		q.logger.Error("Failed to mark job as completed", zap.String("job_id", jobID), zap.Error(err))
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	q.logger.Info("Job marked as completed", zap.String("job_id", jobID))
	return nil
}

func (q *redisQueue) MarkFailed(ctx context.Context, jobID string, retryDelay time.Duration) error {
	pipe := q.client.TxPipeline()

	pipe.ZRem(ctx, ProcessingSetKey, jobID)

	if retryDelay > 0 {
		retryTime := time.Now().Add(retryDelay)
		pipe.ZAdd(ctx, RetrySetKey, redis.Z{
			Score:  float64(retryTime.Unix()),
			Member: jobID,
		})
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		q.logger.Error("Failed to mark job as failed", zap.String("job_id", jobID), zap.Error(err))
		return fmt.Errorf("failed to mark job as failed: %w", err)
	}

	q.logger.Info("Job marked as failed",
		zap.String("job_id", jobID),
		zap.Duration("retry_delay", retryDelay))

	return nil
}

func (q *redisQueue) GetQueueDepth(ctx context.Context, queue string) (int64, error) {
	return q.client.LLen(ctx, queue).Result()
}

func (q *redisQueue) GetProcessingJobs(ctx context.Context) ([]string, error) {
	return q.client.ZRange(ctx, ProcessingSetKey, 0, -1).Result()
}

func (q *redisQueue) getQueueKey(priority job.Priority) string {
	// All jobs go to the same queue regardless of priority
	return QueueKey
}

func GetPriorityQueues() []string {
	// Single queue for all jobs
	return []string{QueueKey}
}
