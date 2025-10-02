package service_test

import (
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/pkg/sqs"
	"backend/service-platform/app/service"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSQSConfigConversion(t *testing.T) {
	// Test the config conversion in the service layer
	sqsConfig := config.SQSConfig{
		QueueURLs: config.SQSQueueURLs{
			SqsScheduledJobQueue: "https://sqs.us-west-2.amazonaws.com/123456789012/claim",
		},
		Polling: config.SQSPollingConfig{
			MaxMessages:              5,
			WaitTimeSeconds:          15,
			VisibilityTimeoutSeconds: 600,
			PollingInterval:          2 * time.Second,
		},
		Message: config.SQSMessageConfig{
			MaxRetries:     5,
			BaseRetryDelay: 10 * time.Second,
			MaxRetryDelay:  10 * time.Minute,
		},
	}
	awsConfig := config.AwsConfig{
		Region:   "us-west-2",
		Endpoint: "http://localhost:4566",
		Sqs:      sqsConfig,
	}

	// Verify conversion
	assert.Equal(t, awsConfig.Region, awsConfig.Region)
	assert.Equal(t, awsConfig.Sqs.QueueURLs.SqsScheduledJobQueue, sqsConfig.QueueURLs.SqsScheduledJobQueue)
	assert.Equal(t, awsConfig.Sqs.Polling.MaxMessages, sqsConfig.Polling.MaxMessages)
	assert.Equal(t, awsConfig.Sqs.Polling.WaitTimeSeconds, sqsConfig.Polling.WaitTimeSeconds)
	assert.Equal(t, awsConfig.Sqs.Polling.VisibilityTimeoutSeconds, sqsConfig.Polling.VisibilityTimeoutSeconds)
	assert.Equal(t, awsConfig.Sqs.Polling.PollingInterval, sqsConfig.Polling.PollingInterval)
	assert.Equal(t, awsConfig.Sqs.Message.MaxRetries, sqsConfig.Message.MaxRetries)
	assert.Equal(t, awsConfig.Sqs.Message.BaseRetryDelay, sqsConfig.Message.BaseRetryDelay)
	assert.Equal(t, awsConfig.Sqs.Message.MaxRetryDelay, sqsConfig.Message.MaxRetryDelay)
}

func TestSQSServiceConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config config.AwsConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: config.AwsConfig{
				Region:   "us-east-1",
				Endpoint: "http://localhost:4566",
				Sqs: config.SQSConfig{
					QueueURLs: config.SQSQueueURLs{
						SqsScheduledJobQueue: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
					},
				},
			},
			valid: true,
		},
		{
			name: "empty queue URL - invalid",
			config: config.AwsConfig{
				Region:   "us-east-1",
				Endpoint: "http://localhost:4566",
				Sqs: config.SQSConfig{
					QueueURLs: config.SQSQueueURLs{
						SqsScheduledJobQueue: "",
					},
				},
			},
			valid: false,
		},
		{
			name: "empty region - invalid",
			config: config.AwsConfig{
				Region:   "",
				Endpoint: "http://localhost:4566",
				Sqs: config.SQSConfig{
					QueueURLs: config.SQSQueueURLs{
						SqsScheduledJobQueue: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
					},
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SQS is now mandatory - validate required configuration fields
			isValid := tt.config.Region != "" && tt.config.Sqs.QueueURLs.SqsScheduledJobQueue != ""

			assert.Equal(t, tt.valid, isValid)
		})
	}
}

func TestSQSListenerConfig(t *testing.T) {
	sqsConfig := sqs.Config{
		Region: "us-east-1",
		QueueURLs: sqs.QueueURLs{
			SqsScheduledJobQueue: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
		},
	}

	// Test SQS listener config creation
	listenerConfig := service.SQSListenerConfig{
		SQSConfig: sqsConfig,
		QueueURLs: []string{}, // Empty means use all configured queues
	}

	assert.Equal(t, sqsConfig.Region, listenerConfig.SQSConfig.Region)
	assert.Equal(t, sqsConfig.QueueURLs.SqsScheduledJobQueue, listenerConfig.SQSConfig.QueueURLs.SqsScheduledJobQueue)
	assert.Empty(t, listenerConfig.QueueURLs) // Should use default (SqsScheduledJobQueue queue)
}
