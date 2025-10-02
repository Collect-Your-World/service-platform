package sqs

import "time"

// Config holds SQS configuration
type Config struct {
	// AWS Region
	Region string `yaml:"region" mapstructure:"region"`

	// AWS Endpoint (for LocalStack or custom endpoints)
	Endpoint string `yaml:"endpoint" mapstructure:"endpoint"`

	// Queue URLs for different priorities
	QueueURLs QueueURLs `yaml:"queue_urls" mapstructure:"queue_urls"`

	// Polling configuration
	Polling PollingConfig `yaml:"polling" mapstructure:"polling"`

	// Message configuration
	Message MessageConfig `yaml:"message" mapstructure:"message"`
}

// QueueURLs defines the SQS queue URLs for business features
type QueueURLs struct {
	SqsScheduledJobQueue string `mapstructure:"sqs_scheduled_job_queue"`
}

// PollingConfig defines SQS polling behavior
type PollingConfig struct {
	// Maximum number of messages to receive in a single request (1-10)
	MaxMessages int `yaml:"max_messages" mapstructure:"max_messages"`

	// Wait time for long polling (0-20 seconds)
	WaitTimeSeconds int `yaml:"wait_time_seconds" mapstructure:"wait_time_seconds"`

	// Message visibility timeout (0-43200 seconds)
	VisibilityTimeoutSeconds int `yaml:"visibility_timeout_seconds" mapstructure:"visibility_timeout_seconds"`

	// Polling interval when no messages are received
	PollingInterval time.Duration `yaml:"polling_interval" mapstructure:"polling_interval"`
}

// MessageConfig defines message handling configuration
type MessageConfig struct {
	// Maximum retry attempts
	MaxRetries int `yaml:"max_retries" mapstructure:"max_retries"`

	// Base delay for exponential backoff
	BaseRetryDelay time.Duration `yaml:"base_retry_delay" mapstructure:"base_retry_delay"`

	// Maximum delay between retries
	MaxRetryDelay time.Duration `yaml:"max_retry_delay" mapstructure:"max_retry_delay"`
}

// DefaultConfig returns a default SQS configuration
func DefaultConfig() Config {
	return Config{
		Region: "us-east-1",
		QueueURLs: QueueURLs{
			SqsScheduledJobQueue: "",
		},
		Polling: PollingConfig{
			MaxMessages:              10,
			WaitTimeSeconds:          20,
			VisibilityTimeoutSeconds: 300, // 5 minutes
			PollingInterval:          time.Second,
		},
		Message: MessageConfig{
			MaxRetries:     3,
			BaseRetryDelay: 30 * time.Second,
			MaxRetryDelay:  15 * time.Minute,
		},
	}
}
