package sqs_test

import (
	"backend/service-platform/app/pkg/sqs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateRetryDelay(t *testing.T) {
	config := sqs.Config{
		Message: sqs.MessageConfig{
			BaseRetryDelay: 5 * time.Second,
			MaxRetryDelay:  2 * time.Minute,
		},
	}

	// Create a mock queue with the test config
	queue := &MockQueue{config: config}

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "attempt 0",
			attempt:  0,
			expected: 5 * time.Second, // base delay
		},
		{
			name:     "attempt 1",
			attempt:  1,
			expected: 5 * time.Second, // base delay * 2^0 = 5s
		},
		{
			name:     "attempt 2",
			attempt:  2,
			expected: 10 * time.Second, // base delay * 2^1 = 10s
		},
		{
			name:     "attempt 3",
			attempt:  3,
			expected: 20 * time.Second, // base delay * 2^2 = 20s
		},
		{
			name:     "attempt 10 (capped)",
			attempt:  10,
			expected: 2 * time.Minute, // capped at max delay
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := queue.CalculateRetryDelay(tt.attempt)
			assert.Equal(t, tt.expected, delay)
		})
	}
}

func TestGetQueueURLByJobType(t *testing.T) {
	config := sqs.Config{
		QueueURLs: sqs.QueueURLs{
			SqsScheduledJobQueue: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
		},
	}

	queue := &MockQueue{config: config}

	tests := []struct {
		name     string
		jobType  string
		expected string
	}{
		{
			name:     "init claim",
			jobType:  "init_claim",
			expected: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
		},
		{
			name:     "claim processing",
			jobType:  "claim_processing",
			expected: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
		},
		{
			name:     "kyc verification",
			jobType:  "kyc_verification",
			expected: "", // Should return empty for non-claim jobs
		},
		{
			name:     "unknown job type",
			jobType:  "unknown_job",
			expected: "", // Should return empty for unknown job types
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := queue.GetQueueURLByJobType(tt.jobType)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestGetQueueURLFromName(t *testing.T) {
	config := sqs.Config{
		QueueURLs: sqs.QueueURLs{
			SqsScheduledJobQueue: "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
		},
	}

	queue := &MockQueue{config: config}

	tests := []struct {
		name      string
		queueName string
		expected  string
	}{
		{
			name:      "claim queue",
			queueName: "claim",
			expected:  "https://sqs.us-east-1.amazonaws.com/123456789012/claim",
		},
		{
			name:      "unknown queue",
			queueName: "unknown_queue",
			expected:  "", // Should return empty for unknown queues
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := queue.GetQueueURLFromName(tt.queueName)
			assert.Equal(t, tt.expected, url)
		})
	}
}

// MockQueue implements the queue routing logic for testing
type MockQueue struct {
	config sqs.Config
}

func (m *MockQueue) CalculateRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return m.config.Message.BaseRetryDelay
	}

	// Exponential backoff: baseDelay * 2^(attempt-1)
	delay := m.config.Message.BaseRetryDelay * time.Duration(1<<(attempt-1))

	// Cap at maximum delay
	if delay > m.config.Message.MaxRetryDelay {
		delay = m.config.Message.MaxRetryDelay
	}

	return delay
}

func (m *MockQueue) GetQueueURLByJobType(jobType string) string {
	switch jobType {
	case "init_claim", "complete_claim", "claim_processing":
		return m.config.QueueURLs.SqsScheduledJobQueue
	default:
		return ""
	}
}

func (m *MockQueue) GetQueueURLFromName(queueName string) string {
	switch queueName {
	case "claim":
		return m.config.QueueURLs.SqsScheduledJobQueue
	default:
		return ""
	}
}
