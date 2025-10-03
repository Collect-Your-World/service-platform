package worker

import (
	"backend/service-platform/app/database/entity"
	"context"

	"go.uber.org/zap"
)

type JobHandler interface {
	Handle(ctx context.Context, job *entity.Job) error
	CanHandle(jobType string) bool
	GetType() string
}

type JobHandlerRegistry interface {
	Register(handler JobHandler)
	Get(jobType string) (JobHandler, bool)
	GetAll() map[string]JobHandler
}

type jobHandlerRegistry struct {
	handlers map[string]JobHandler
	logger   *zap.Logger
}

func NewJobHandlerRegistry(logger *zap.Logger) JobHandlerRegistry {
	return &jobHandlerRegistry{
		handlers: make(map[string]JobHandler),
		logger:   logger,
	}
}

func (r *jobHandlerRegistry) Register(handler JobHandler) {
	jobType := handler.GetType()
	r.handlers[jobType] = handler
	r.logger.Debug("Registered job handler", zap.String("type", jobType))
}

func (r *jobHandlerRegistry) Get(jobType string) (JobHandler, bool) {
	handler, exists := r.handlers[jobType]
	return handler, exists
}

func (r *jobHandlerRegistry) GetAll() map[string]JobHandler {
	return r.handlers
}
