package worker

import (
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/internal/job/entities"
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/pkg/sqs"
	"context"
	"time"

	job "backend/service-platform/app/internal/job/constants"

	"go.uber.org/zap"
)

type Services struct {
	WorkerService      *WorkerService
	SQSListenerService *SQSListenerService
}

type JobManagerInterface interface {
	CreateJob(ctx context.Context, req CreateJobRequest) (*entity.Job, error)
}

type CreateJobRequest struct {
	Type        string                 `json:"type"`
	Priority    job.Priority           `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	MaxAttempts int                    `json:"max_attempts,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

func NewServices(res runtime.Resource, workerConfig config.WorkerConfig) *Services {
	workerService := NewWorkerService(res, workerConfig)

	res.Logger.Info("Initializing SQS services")

	return &Services{
		WorkerService:      workerService,
		SQSListenerService: nil,
	}
}

func NewServicesWithJobManager(res runtime.Resource, workerConfig config.WorkerConfig, rawJobManager interface{}) *Services {
	workerService := NewWorkerService(res, workerConfig)

	res.Logger.Info("Initializing SQS services")

	sqsConfig := convertToSQSConfig(res.Config.AwsConfig)

	sqsListenerConfig := SQSListenerConfig{
		SQSConfig:  sqsConfig,
		JobManager: rawJobManager,
		QueueURLs:  []string{},
	}

	sqsListenerService, err := NewSQSListenerService(res, sqsListenerConfig)
	if err != nil {
		res.Logger.Error("Failed to create SQS listener service", zap.Error(err))
		panic(err)
	}
	res.Logger.Info("SQS listener service created successfully")

	return &Services{
		WorkerService:      workerService,
		SQSListenerService: sqsListenerService,
	}
}

func convertToSQSConfig(awsConfig config.AwsConfig) sqs.Config {
	return sqs.Config{
		Region:   awsConfig.Region,
		Endpoint: awsConfig.Endpoint,
		QueueURLs: sqs.QueueURLs{
			SqsScheduledJobQueue: awsConfig.Sqs.QueueURLs.SqsScheduledJobQueue,
		},
		Polling: sqs.PollingConfig{
			MaxMessages:              awsConfig.Sqs.Polling.MaxMessages,
			WaitTimeSeconds:          awsConfig.Sqs.Polling.WaitTimeSeconds,
			VisibilityTimeoutSeconds: awsConfig.Sqs.Polling.VisibilityTimeoutSeconds,
			PollingInterval:          awsConfig.Sqs.Polling.PollingInterval,
		},
		Message: sqs.MessageConfig{
			MaxRetries:     awsConfig.Sqs.Message.MaxRetries,
			BaseRetryDelay: awsConfig.Sqs.Message.BaseRetryDelay,
			MaxRetryDelay:  awsConfig.Sqs.Message.MaxRetryDelay,
		},
	}
}
