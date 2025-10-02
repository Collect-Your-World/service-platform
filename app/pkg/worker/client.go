package worker

import (
	"context"
	"fmt"
	"time"

	"backend/service-platform/app/pkg/locker"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/go-co-op/gocron/v2"
)

type Client interface {
	AddJob(ctx context.Context, taskID uuid.UUID, jo []gocron.JobOption, taskFunc any, args ...any) error
	AddJobWithDelay(ctx context.Context, taskID uuid.UUID, delay time.Duration, jo []gocron.JobOption, taskFunc any, args ...any) error
	AddUniqueJob(ctx context.Context, taskID uuid.UUID, key string, jo []gocron.JobOption, taskFunc any, args ...any) error
	IsJobValid(jobID uuid.UUID) bool
	StopJob(jobID uuid.UUID)
	StopAllJobs()
	AddJobName(ctx context.Context, taskID uuid.UUID, jo []gocron.JobOption, taskName string, args ...any) error
	AddJobNameWithDelay(ctx context.Context, taskID uuid.UUID, delay time.Duration, jo []gocron.JobOption, taskName string, args ...any) error
	AddUniqueJobName(ctx context.Context, taskID uuid.UUID, key string, jo []gocron.JobOption, taskName string, args ...any) error
	AddDurationJob(ctx context.Context, taskID uuid.UUID, duration time.Duration, jo []gocron.JobOption, taskFunc any, args ...any) error
}

type DefaultClient struct {
	log       *zap.Logger
	locker    gocron.Locker
	scheduler Scheduler
	jobPool   JobPool
}

func NewClient(
	log *zap.Logger,
	locker locker.Locker,
	scheduler Scheduler,
) (Client, error) {
	return &DefaultClient{
		log:       log,
		scheduler: scheduler,
		locker:    locker,
		jobPool:   NewJobPool(),
	}, nil
}

func (d *DefaultClient) IsJobValid(jobID uuid.UUID) bool {
	return d.jobPool.IsJobValid(jobID)
}

func (d *DefaultClient) StopJob(jobID uuid.UUID) {
	d.jobPool.StopJob(jobID)
	if err := d.scheduler.RemoveJob(jobID); err != nil {
		d.log.Warn("failed to remove job",
			zap.String("jobID", jobID.String()),
			zap.Error(err),
		)
	}
}

func (d *DefaultClient) StopAllJobs() {
	d.scheduler.StopJobs()
	d.jobPool.StopAllJobs()
}

func (d *DefaultClient) AddJobName(ctx context.Context, taskID uuid.UUID, jo []gocron.JobOption, taskName string,
	args ...any) error {
	taskFunc := d.scheduler.GetTask(taskName)
	if taskFunc == nil {
		return fmt.Errorf("task %s is not defined", taskName)
	}
	return d.AddJob(ctx, taskID, jo, taskFunc, args...)
}

func (d *DefaultClient) AddJobNameWithDelay(ctx context.Context, taskID uuid.UUID, delay time.Duration,
	jo []gocron.JobOption, taskName string, args ...any) error {
	taskFunc := d.scheduler.GetTask(taskName)
	if taskFunc == nil {
		return fmt.Errorf("task %s is not defined", taskName)
	}
	return d.AddJobWithDelay(ctx, taskID, delay, jo, taskFunc, args...)
}

func (d *DefaultClient) AddUniqueJobName(ctx context.Context, taskID uuid.UUID, key string, jo []gocron.JobOption,
	taskName string, args ...any) error {
	taskFunc := d.scheduler.GetTask(taskName)
	if taskFunc == nil {
		return fmt.Errorf("task %s is not defined", taskName)
	}
	return d.AddUniqueJob(ctx, taskID, key, jo, taskFunc, args...)
}

func (d *DefaultClient) AddJob(ctx context.Context, taskID uuid.UUID, jo []gocron.JobOption, taskFunc any,
	args ...any) error {
	jd := gocron.OneTimeJob(gocron.OneTimeJobStartImmediately())
	ctx = d.jobPool.NewJob(ctx, taskID)
	args = append([]any{ctx}, args...)
	jo = append(jo, gocron.WithIdentifier(taskID))
	_, err := d.scheduler.NewJob(jd, gocron.NewTask(taskFunc, args...), jo...)
	return err
}

func (d *DefaultClient) AddJobWithDelay(ctx context.Context, taskID uuid.UUID, delay time.Duration,
	jo []gocron.JobOption, taskFunc any, args ...any) error {
	jd := gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(delay)))
	ctx = d.jobPool.NewJob(ctx, taskID)
	args = append([]any{ctx}, args...)
	jo = append(jo, gocron.WithIdentifier(taskID))
	_, err := d.scheduler.NewJob(jd, gocron.NewTask(taskFunc, args...), jo...)
	return err
}

func (d *DefaultClient) AddUniqueJob(ctx context.Context, taskID uuid.UUID, key string, jo []gocron.JobOption, taskFunc any, args ...any) error {
	jd := gocron.OneTimeJob(gocron.OneTimeJobStartImmediately())
	ctx = d.jobPool.NewJob(ctx, taskID)
	args = append([]any{ctx}, args...)
	jo = append(jo,
		gocron.WithIdentifier(taskID),
		gocron.WithName(key),
		gocron.WithDistributedJobLocker(d.locker),
	)
	_, err := d.scheduler.NewJob(jd, gocron.NewTask(taskFunc, args...), jo...)
	return err
}

func (d *DefaultClient) AddDurationJob(ctx context.Context, taskID uuid.UUID, duration time.Duration, jo []gocron.JobOption, taskFunc any, args ...any) error {
	jd := gocron.DurationJob(duration)
	ctx = d.jobPool.NewJob(ctx, taskID)
	args = append([]any{ctx}, args...)
	jo = append(jo, gocron.WithIdentifier(taskID))
	_, err := d.scheduler.NewJob(jd, gocron.NewTask(taskFunc, args...), jo...)

	return err
}
