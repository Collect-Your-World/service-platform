package handlers

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
)

type KYCVerificationHandler struct {
	logger *zap.Logger
}

func NewKYCVerificationHandler(logger *zap.Logger) *KYCVerificationHandler {
	return &KYCVerificationHandler{
		logger: logger.With(zap.String("handler", "kyc_verification")),
	}
}

func (h *KYCVerificationHandler) Handle(ctx context.Context, job *entity.Job) error {
	h.logger.Info("Processing KYC verification job",
		zap.String("job_id", job.ID),
		zap.Any("payload", job.Payload))

	// Simulate KYC processing with potential failure for testing
	time.Sleep(3 * time.Second)

	// Simulate occasional failures for retry testing
	if job.Attempts > 0 && job.Attempts%2 == 0 {
		h.logger.Warn("Simulating KYC verification failure for retry testing",
			zap.String("job_id", job.ID),
			zap.Int("attempt", job.Attempts))
		return errors.New("KYC service temporarily unavailable")
	}

	h.logger.Info("KYC verification completed",
		zap.String("job_id", job.ID))

	return nil
}

func (h *KYCVerificationHandler) CanHandle(jobType string) bool {
	return jobType == string(job.KYCVerification)
}

func (h *KYCVerificationHandler) GetType() string {
	return string(job.KYCVerification)
}
