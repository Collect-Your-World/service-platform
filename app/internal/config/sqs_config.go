package config

import (
	"time"
)

// SQSConfig represents SQS configuration for the application
type SQSConfig struct {
	// Queue URLs for different priorities
	QueueURLs SQSQueueURLs `mapstructure:"queue_urls"`

	// Polling configuration
	Polling SQSPollingConfig `mapstructure:"polling"`

	// Message configuration
	Message SQSMessageConfig `mapstructure:"message"`
}

// SQSQueueURLs defines the SQS queue URLs for business features
type SQSQueueURLs struct {
	SqsScheduledJobQueue string `mapstructure:"sqs_scheduled_job_queue"`
}

// SQSPollingConfig defines SQS polling behavior
type SQSPollingConfig struct {
	// Maximum number of messages to receive in a single request (1-10)
	MaxMessages int `mapstructure:"max_messages"`

	// Wait time for long polling (0-20 seconds)
	WaitTimeSeconds int `mapstructure:"wait_time_seconds"`

	// Message visibility timeout (0-43200 seconds)
	VisibilityTimeoutSeconds int `mapstructure:"visibility_timeout_seconds"`

	// Polling interval when no messages are received
	PollingInterval time.Duration `mapstructure:"polling_interval"`
}

// SQSMessageConfig defines message handling configuration
type SQSMessageConfig struct {
	// Maximum retry attempts
	MaxRetries int `mapstructure:"max_retries"`

	// Base delay for exponential backoff
	BaseRetryDelay time.Duration `mapstructure:"base_retry_delay"`

	// Maximum delay between retries
	MaxRetryDelay time.Duration `mapstructure:"max_retry_delay"`
}
