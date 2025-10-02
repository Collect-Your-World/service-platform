package repository

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/internal/runtime"
	"context"
	"time"
)

type JobRepository interface {
	Create(ctx context.Context, job *entity.Job) error
	GetByID(ctx context.Context, id string) (*entity.Job, error)
	GetBySQSMessageID(ctx context.Context, sqsMessageID string) (*entity.Job, error)
	UpdateStatus(ctx context.Context, id string, status job.Status, error string) error
	UpdateStartTime(ctx context.Context, id string, startedAt time.Time) error
	UpdateCompleteTime(ctx context.Context, id string, completedAt time.Time) error
	IncrementAttempts(ctx context.Context, id string) error
	UpdateJobToProcessing(ctx context.Context, id string, startedAt time.Time) error
	UpdateJobToCompleted(ctx context.Context, id string, completedAt time.Time) error
	UpdateJobToFailed(ctx context.Context, id string, errorMsg string) error
	UpdateJobToRetrying(ctx context.Context, id string, errorMsg string) error
	GetPendingJobs(ctx context.Context, limit int) ([]*entity.Job, error)
	GetJobsByStatus(ctx context.Context, status job.Status, limit int) ([]*entity.Job, error)
	GetRetryableJobs(ctx context.Context, beforeTime time.Time, limit int) ([]*entity.Job, error)
}

type jobRepository struct {
	res runtime.Resource
}

func NewJobRepository(res runtime.Resource) JobRepository {
	return &jobRepository{res: res}
}

func (r *jobRepository) Create(ctx context.Context, job *entity.Job) error {
	_, err := r.res.DB.NewInsert().Model(job).Exec(ctx)
	return err
}

func (r *jobRepository) GetByID(ctx context.Context, id string) (*entity.Job, error) {
	job := &entity.Job{}
	err := r.res.DB.NewSelect().Model(job).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (r *jobRepository) UpdateStatus(ctx context.Context, id string, status job.Status, errorMsg string) error {
	update := r.res.DB.NewUpdate().
		Model((*entity.Job)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL")

	if errorMsg != "" {
		update = update.Set("error = ?", errorMsg)
	}

	_, err := update.Exec(ctx)
	return err
}

func (r *jobRepository) UpdateStartTime(ctx context.Context, id string, startedAt time.Time) error {
	_, err := r.res.DB.NewUpdate().
		Model((*entity.Job)(nil)).
		Set("started_at = ?", startedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

func (r *jobRepository) UpdateCompleteTime(ctx context.Context, id string, completedAt time.Time) error {
	_, err := r.res.DB.NewUpdate().
		Model((*entity.Job)(nil)).
		Set("completed_at = ?", completedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *jobRepository) IncrementAttempts(ctx context.Context, id string) error {
	_, err := r.res.DB.NewUpdate().
		Model((*entity.Job)(nil)).
		Set("attempts = attempts + 1").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *jobRepository) GetPendingJobs(ctx context.Context, limit int) ([]*entity.Job, error) {
	var jobs []*entity.Job
	err := r.res.DB.NewSelect().
		Model(&jobs).
		Where("status = ?", job.Pending).
		Where("deleted_at IS NULL").
		Where("scheduled_at IS NULL OR scheduled_at <= ?", time.Now()).
		Order("priority DESC").
		Order("created_at ASC").
		Limit(limit).
		Scan(ctx)
	return jobs, err
}

func (r *jobRepository) GetJobsByStatus(ctx context.Context, status job.Status, limit int) ([]*entity.Job, error) {
	var jobs []*entity.Job
	err := r.res.DB.NewSelect().
		Model(&jobs).
		Where("status = ?", status).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	return jobs, err
}

func (r *jobRepository) UpdateJobToProcessing(ctx context.Context, id string, startedAt time.Time) error {
	_, err := r.res.DB.NewUpdate().
		Model((*entity.Job)(nil)).
		Set("status = ?", job.Processing).
		Set("started_at = ?", startedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

func (r *jobRepository) UpdateJobToCompleted(ctx context.Context, id string, completedAt time.Time) error {
	_, err := r.res.DB.NewUpdate().
		Model((*entity.Job)(nil)).
		Set("status = ?", job.Completed).
		Set("completed_at = ?", completedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

func (r *jobRepository) UpdateJobToFailed(ctx context.Context, id string, errorMsg string) error {
	update := r.res.DB.NewUpdate().
		Model((*entity.Job)(nil)).
		Set("status = ?", job.Failed).
		Set("attempts = attempts + 1").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL")

	if errorMsg != "" {
		update = update.Set("error = ?", errorMsg)
	}

	_, err := update.Exec(ctx)
	return err
}

func (r *jobRepository) UpdateJobToRetrying(ctx context.Context, id string, errorMsg string) error {
	update := r.res.DB.NewUpdate().
		Model((*entity.Job)(nil)).
		Set("status = ?", job.Retrying).
		Set("attempts = attempts + 1").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL")

	if errorMsg != "" {
		update = update.Set("error = ?", errorMsg)
	}

	_, err := update.Exec(ctx)
	return err
}

func (r *jobRepository) GetRetryableJobs(ctx context.Context, beforeTime time.Time, limit int) ([]*entity.Job, error) {
	var jobs []*entity.Job
	err := r.res.DB.NewSelect().
		Model(&jobs).
		Where("status = ?", job.Failed).
		Where("attempts < max_attempts").
		Where("deleted_at IS NULL").
		Where("updated_at < ?", beforeTime).
		Order("priority DESC").
		Order("updated_at ASC").
		Limit(limit).
		Scan(ctx)
	return jobs, err
}

func (r *jobRepository) GetBySQSMessageID(ctx context.Context, sqsMessageID string) (*entity.Job, error) {
	job := &entity.Job{}
	err := r.res.DB.NewSelect().
		Model(job).
		Where("payload->'_sqs_metadata'->>'sqs_message_id' = ?", sqsMessageID).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return job, nil
}
