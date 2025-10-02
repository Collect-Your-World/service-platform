package service

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/sqs"
	"context"
	"time"

	"go.uber.org/zap"
)

type Services struct {
	WorkerService      *WorkerService
	SQSListenerService *SQSListenerService
}

// JobManagerInterface defines the interface we need from manager.JobManager to avoid import cycle
type JobManagerInterface interface {
	CreateJob(ctx context.Context, req CreateJobRequest) (*entity.Job, error)
}

// CreateJobRequest defines the request structure for creating jobs
type CreateJobRequest struct {
	Type        string                 `json:"type"`
	Priority    job.Priority           `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	MaxAttempts int                    `json:"max_attempts,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

func NewServices(res runtime.Resource, workerConfig config.WorkerConfig) *Services {

	// Init Worker service
	workerService := NewWorkerService(res, workerConfig)

	// Init SQS services
	res.Logger.Info("Initializing SQS services")

	// SQS listener service will be nil without a job manager
	// Use NewServicesWithJobManager instead

	return &Services{
		WorkerService:      workerService,
		SQSListenerService: nil, // Will be created in NewServicesWithJobManager
	}
}

// NewServicesWithJobManager creates services with a provided job manager to avoid circular dependencies
func NewServicesWithJobManager(res runtime.Resource, workerConfig config.WorkerConfig, rawJobManager interface{}) *Services {
	// Init Worker service
	workerService := NewWorkerService(res, workerConfig)

	// Init SQS services with provided job manager
	res.Logger.Info("Initializing SQS services")

	// Convert config to SQS package format
	sqsConfig := convertToSQSConfig(res.Config.AwsConfig)

	// Create SQS listener service with provided job manager
	sqsListenerConfig := SQSListenerConfig{
		SQSConfig:  sqsConfig,
		JobManager: rawJobManager,
		QueueURLs:  []string{}, // Use all queues
	}

	sqsListenerService, err := NewSQSListenerService(res, sqsListenerConfig)
	if err != nil {
		res.Logger.Error("Failed to create SQS listener service", zap.Error(err))
		panic(err) // SQS is mandatory, fail if it can't be initialized
	}
	res.Logger.Info("SQS listener service created successfully")

	return &Services{
		WorkerService:      workerService,
		SQSListenerService: sqsListenerService,
	}
}

// convertToSQSConfig converts application config to SQS package config
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
