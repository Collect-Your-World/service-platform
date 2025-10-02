package worker

import (
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/pkg/locker"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type Scheduler interface {
	gocron.Scheduler
	ImportTasks(taskList map[string]any)
	GetTask(taskName string) any
}

type DefaultScheduler struct {
	gocron.Scheduler
	log      *zap.Logger
	taskList map[string]any
}

func NewScheduler(cfg config.WorkerConfig, log *zap.Logger, locker locker.Locker) (Scheduler, error) {
	cron, err := gocron.NewScheduler(
		gocron.WithLimitConcurrentJobs(uint(cfg.PoolSize), gocron.LimitModeWait),
		gocron.WithLogger(NewWorkerLog(log.Sugar())),
		// gocron.WithDistributedLocker(locker),
	)
	return &DefaultScheduler{
		log:       log,
		Scheduler: cron,
		taskList:  map[string]any{},
	}, err
}

func (d *DefaultScheduler) ImportTasks(taskList map[string]any) {
	for taskName, taskFunc := range taskList {
		d.taskList[taskName] = taskFunc
	}
}

func (d *DefaultScheduler) GetTask(taskName string) any {
	return d.taskList[taskName]
}
