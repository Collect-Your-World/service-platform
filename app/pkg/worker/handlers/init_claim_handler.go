package handlers

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"context"
	"time"

	"go.uber.org/zap"
)

type InitClaimHandler struct {
	logger *zap.Logger
}

func NewInitClaimHandler(logger *zap.Logger) *InitClaimHandler {
	return &InitClaimHandler{
		logger: logger.With(zap.String("handler", "init_claim")),
	}
}

func (h *InitClaimHandler) Handle(ctx context.Context, job *entity.Job) error {
	h.logger.Info("Processing init claim job",
		zap.String("job_id", job.ID),
		zap.Any("payload", job.Payload))

	time.Sleep(2 * time.Second)

	h.logger.Info("Init claim completed",
		zap.String("job_id", job.ID))

	return nil
}

func (h *InitClaimHandler) CanHandle(jobType string) bool {
	return jobType == string(job.InitClaim)
}

func (h *InitClaimHandler) GetType() string {
	return string(job.InitClaim)
}
