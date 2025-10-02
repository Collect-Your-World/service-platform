package manager

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/pkg/queue"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type JobManager interface {
	CreateJob(ctx context.Context, req CreateJobRequest) (*entity.Job, error)
	GetJob(ctx context.Context, id uuid.UUID) (*entity.Job, error)
	GetJobsByStatus(ctx context.Context, status job.Status, limit int) ([]*entity.Job, error)
}

type CreateJobRequest struct {
	Type        string                 `json:"type"`
	Priority    job.Priority           `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	MaxAttempts int                    `json:"max_attempts,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

type jobManager struct {
	jobRepo repository.JobRepository
	queue   queue.Queue
	logger  *zap.Logger
}

func NewJobManager(
	jobRepo repository.JobRepository,
	queue queue.Queue,
	logger *zap.Logger,
) JobManager {
	return &jobManager{
		jobRepo: jobRepo,
		queue:   queue,
		logger:  logger,
	}
}

func (m *jobManager) CreateJob(ctx context.Context, req CreateJobRequest) (*entity.Job, error) {
	if req.Type == "" {
		return nil, fmt.Errorf("job type is required")
	}

	if req.MaxAttempts <= 0 {
		req.MaxAttempts = 3
	}

	if req.Payload == nil {
		req.Payload = make(map[string]interface{})
	}

	jobEntity := &entity.Job{
		ID:          uuid.New(),
		Type:        req.Type,
		Priority:    req.Priority,
		Payload:     entity.JobPayload(req.Payload),
		MaxAttempts: req.MaxAttempts,
		CreatedAt:   time.Now(),
		ScheduledAt: req.ScheduledAt,
		Status:      job.Pending,
	}

	if err := m.jobRepo.Create(ctx, jobEntity); err != nil {
		m.logger.Error("Failed to create job in database",
			zap.String("job_id", jobEntity.ID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Enqueue job
	if err := m.queue.Enqueue(ctx, jobEntity); err != nil {
		m.logger.Error("Failed to enqueue job",
			zap.String("job_id", jobEntity.ID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	m.logger.Info("Job created successfully",
		zap.String("job_id", jobEntity.ID.String()),
		zap.String("type", jobEntity.Type),
		zap.String("priority", jobEntity.Priority.String()))

	return jobEntity, nil
}

func (m *jobManager) GetJob(ctx context.Context, id uuid.UUID) (*entity.Job, error) {
	return m.jobRepo.GetByID(ctx, id)
}

func (m *jobManager) GetJobsByStatus(ctx context.Context, status job.Status, limit int) ([]*entity.Job, error) {
	if limit <= 0 {
		limit = 50
	}
	return m.jobRepo.GetJobsByStatus(ctx, status, limit)
}
