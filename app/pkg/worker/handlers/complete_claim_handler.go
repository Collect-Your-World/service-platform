package handlers

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"context"
	"time"

	"go.uber.org/zap"
)

type CompleteClaimHandler struct {
	logger *zap.Logger
}

func NewCompleteClaimHandler(logger *zap.Logger) *CompleteClaimHandler {
	return &CompleteClaimHandler{
		logger: logger.With(zap.String("handler", "complete_claim")),
	}
}

func (h *CompleteClaimHandler) Handle(ctx context.Context, job *entity.Job) error {
	h.logger.Info("Processing complete claim job",
		zap.String("job_id", job.ID),
		zap.Any("payload", job.Payload))

	time.Sleep(1 * time.Second)

	h.logger.Info("Complete claim completed",
		zap.String("job_id", job.ID))

	return nil
}

func (h *CompleteClaimHandler) CanHandle(jobType string) bool {
	return jobType == string(job.CompleteClaim)
}

func (h *CompleteClaimHandler) GetType() string {
	return string(job.CompleteClaim)
}
