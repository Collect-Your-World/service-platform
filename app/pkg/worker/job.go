package worker

import (
	"context"
	"sync"

	ctxutil "backend/service-platform/app/pkg/util/context"

	"github.com/google/uuid"
)

const TaskIDKey ctxutil.ContextKey[uuid.UUID] = "background_task_id"

type Job struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewJob(ctx context.Context, cancel context.CancelFunc) *Job {
	return &Job{
		ctx:    ctx,
		cancel: cancel,
	}
}

type JobPool struct {
	jobList      map[uuid.UUID]*Job
	mutexJobList *sync.RWMutex
}

const MAX_JOB_POOL_SIZE = 10000

func NewJobPool() JobPool {
	return JobPool{
		jobList:      make(map[uuid.UUID]*Job, MAX_JOB_POOL_SIZE),
		mutexJobList: &sync.RWMutex{},
	}
}

func (jp *JobPool) NewJob(ctx context.Context, taskID uuid.UUID) context.Context {
	// Decouple a context from its original cancellation mechanism
	ctx = context.WithoutCancel(ctx)
	ctx = TaskIDKey.Set(ctx, taskID)
	ctx, cancel := context.WithCancel(ctx)
	jp.mutexJobList.Lock()
	defer jp.mutexJobList.Unlock()
	jp.jobList[taskID] = NewJob(ctx, cancel)
	return ctx
}

func (jp JobPool) IsJobValid(jobID uuid.UUID) bool {
	jp.mutexJobList.RLock()
	defer jp.mutexJobList.RUnlock()
	_, ok := jp.jobList[jobID]
	return ok
}

func (jp *JobPool) StopJob(jobID uuid.UUID) {
	jp.mutexJobList.Lock()
	defer jp.mutexJobList.Unlock()
	if job, ok := jp.jobList[jobID]; ok {
		job.cancel()
		delete(jp.jobList, jobID)
	}
}

func (jp *JobPool) StopAllJobs() {
	jp.mutexJobList.Lock()
	defer jp.mutexJobList.Unlock()
	for _, job := range jp.jobList {
		job.cancel()
	}
	jp.jobList = make(map[uuid.UUID]*Job, MAX_JOB_POOL_SIZE)
}
