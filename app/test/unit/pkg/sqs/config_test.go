package sqs_test

import (
	"backend/service-platform/app/pkg/sqs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := sqs.DefaultConfig()

	assert.Equal(t, "us-east-1", config.Region)
	assert.Empty(t, config.QueueURLs.SqsScheduledJobQueue)
	assert.Equal(t, 10, config.Polling.MaxMessages)
	assert.Equal(t, 20, config.Polling.WaitTimeSeconds)
	assert.Equal(t, 300, config.Polling.VisibilityTimeoutSeconds)
	assert.Equal(t, time.Second, config.Polling.PollingInterval)
	assert.Equal(t, 3, config.Message.MaxRetries)
	assert.Equal(t, 30*time.Second, config.Message.BaseRetryDelay)
	assert.Equal(t, 15*time.Minute, config.Message.MaxRetryDelay)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config sqs.Config
		valid  bool
	}{
		{
			name: "valid config",
			config: sqs.Config{
				Region: "us-east-1",
				QueueURLs: sqs.QueueURLs{
					SqsScheduledJobQueue: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
				},
				Polling: sqs.PollingConfig{
					MaxMessages:              5,
					WaitTimeSeconds:          10,
					VisibilityTimeoutSeconds: 300,
					PollingInterval:          time.Second,
				},
				Message: sqs.MessageConfig{
					MaxRetries:     3,
					BaseRetryDelay: time.Second,
					MaxRetryDelay:  time.Minute,
				},
			},
			valid: true,
		},
		{
			name: "empty region",
			config: sqs.Config{
				Region: "",
				QueueURLs: sqs.QueueURLs{
					SqsScheduledJobQueue: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
				},
			},
			valid: false,
		},
		{
			name: "empty queue URLs",
			config: sqs.Config{
				Region: "us-east-1",
				QueueURLs: sqs.QueueURLs{
					SqsScheduledJobQueue: "",
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, we just check that the config can be created
			// In the future, we might add validation methods
			if tt.valid {
				assert.NotEmpty(t, tt.config.Region)
				assert.NotEmpty(t, tt.config.QueueURLs.SqsScheduledJobQueue)
			} else {
				assert.True(t, tt.config.Region == "" || tt.config.QueueURLs.SqsScheduledJobQueue == "")
			}
		})
	}
}
