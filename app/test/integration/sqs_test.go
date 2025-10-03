package integration

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/manager"
	"backend/service-platform/app/pkg/queue"
	"backend/service-platform/app/pkg/sqs"
	service "backend/service-platform/app/service"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type SQSSuite struct {
	RouterSuite
	sqsListenerService *service.SQSListenerService
	workerService      *service.WorkerService
	jobManager         manager.JobManager
	sqsQueue           *sqs.Queue
	sqsClient          *sqs.Client
	testQueueURL       string
}

func TestSQSSuite(t *testing.T) {
	// SQS integration tests require LocalStack credentials to be set as environment variables:
	// AWS_ACCESS_KEY_ID=test
	// AWS_SECRET_ACCESS_KEY=test
	// AWS_DEFAULT_REGION=us-east-1
	// These should be set in .env.test or as actual environment variables
	suite.Run(t, new(SQSSuite))
}

func (s *SQSSuite) SetupTest() {
	s.RouterSuite.SetupTest()

	// Generate dynamic queue name for this test run to avoid conflicts
	testID := uuid.New().String()[:8]
	dynamicQueueURL := fmt.Sprintf("%s/000000000000/test-claim-%s", s.resource.Config.AwsConfig.Endpoint, testID)

	// Update the SQS config with dynamic queue URL
	// Note: LocalStack endpoint is read from config-test.yaml (aws.endpoint)
	// LocalStack credentials are read from environment variables:
	// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_DEFAULT_REGION
	s.resource.Config.AwsConfig.Sqs.QueueURLs.SqsScheduledJobQueue = dynamicQueueURL

	// Try to create SQS services
	s.createSQSServices()
}

func (s *SQSSuite) createSQSServices() {
	// Convert to SQS package config using the loaded YAML config
	sqsConfig := sqs.Config{
		Region:   s.resource.Config.AwsConfig.Region,
		Endpoint: s.resource.Config.AwsConfig.Endpoint, // Use AWS endpoint from config
		QueueURLs: sqs.QueueURLs{
			SqsScheduledJobQueue: s.resource.Config.AwsConfig.Sqs.QueueURLs.SqsScheduledJobQueue,
		},
		Polling: sqs.PollingConfig{
			MaxMessages:              s.resource.Config.AwsConfig.Sqs.Polling.MaxMessages,
			WaitTimeSeconds:          s.resource.Config.AwsConfig.Sqs.Polling.WaitTimeSeconds,
			VisibilityTimeoutSeconds: s.resource.Config.AwsConfig.Sqs.Polling.VisibilityTimeoutSeconds,
			PollingInterval:          s.resource.Config.AwsConfig.Sqs.Polling.PollingInterval,
		},
		Message: sqs.MessageConfig{
			MaxRetries:     s.resource.Config.AwsConfig.Sqs.Message.MaxRetries,
			BaseRetryDelay: s.resource.Config.AwsConfig.Sqs.Message.BaseRetryDelay,
			MaxRetryDelay:  s.resource.Config.AwsConfig.Sqs.Message.MaxRetryDelay,
		},
	}

	// Create SQS client first to create the queue
	var err error
	s.sqsClient, err = sqs.NewClient(s.ctx, sqsConfig, s.resource.Logger)
	if err != nil {
		s.T().Skipf("Failed to create SQS client (SQS not available): %v", err)
		return
	}

	// Extract queue name from URL for creation (LocalStack needs the queue to exist)
	s.testQueueURL = s.resource.Config.AwsConfig.Sqs.QueueURLs.SqsScheduledJobQueue
	queueName := s.extractQueueNameFromURL(s.testQueueURL)

	// Create the queue (this is idempotent in LocalStack)
	_, err = s.sqsClient.CreateQueue(s.ctx, queueName, nil)
	if err != nil {
		s.T().Skipf("Failed to create SQS queue %s: %v", queueName, err)
		return
	}
	s.T().Logf("✓ Created SQS queue: %s", queueName)

	// Create SQS queue wrapper
	s.sqsQueue, err = sqs.NewSQSQueue(s.ctx, sqsConfig, s.resource.Logger)
	if err != nil {
		s.T().Skipf("Failed to create SQS queue wrapper (SQS not available): %v", err)
		return
	}

	// Create worker service for SQS integration
	workerConfig := s.resource.Config.WorkerConfig
	s.workerService = service.NewWorkerService(s.resource, workerConfig)

	// Create Redis queue and job manager
	redisQueue := queue.NewRedisQueue(s.resource.Redis.GetUniversalClient(), s.resource.Logger)
	jobRepo := repository.NewJobRepository(s.resource)
	s.jobManager = manager.NewJobManager(jobRepo, redisQueue, s.resource.Logger)

	// Create SQS listener service using the interface{} approach
	sqsListenerConfig := service.SQSListenerConfig{
		SQSConfig:  sqsConfig,
		JobManager: s.jobManager,
		QueueURLs:  []string{}, // Use all queues
	}

	s.sqsListenerService, err = service.NewSQSListenerService(s.resource, sqsListenerConfig)
	if err != nil {
		s.T().Skipf("Failed to create SQS listener service (SQS not available): %v", err)
		return
	}

	// Start worker service in background
	go func() {
		ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
		defer cancel()

		if err := s.workerService.Start(ctx); err != nil {
			s.resource.Logger.Error("Worker service failed during test", zap.Error(err))
		}
	}()

	// Start SQS listener in background
	go func() {
		ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
		defer cancel()

		if err := s.sqsListenerService.Start(ctx); err != nil {
			s.resource.Logger.Error("SQS listener failed during test", zap.Error(err))
		}
	}()

	// Wait for SQS listener to be ready
	s.waitForSQSListenerReady()
}

func (s *SQSSuite) TearDownTest() {
	// Stop SQS listener service
	if s.sqsListenerService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.sqsListenerService.Stop(ctx)
	}

	// Stop worker service
	if s.workerService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.workerService.Stop(ctx)
	}

	// Clean up test queue (optional for LocalStack, but good practice)
	if s.sqsClient != nil && s.testQueueURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.sqsClient.DeleteQueue(ctx, s.testQueueURL); err != nil {
			s.T().Logf("Warning: Failed to delete test queue %s: %v", s.testQueueURL, err)
		} else {
			s.T().Logf("✓ Cleaned up test queue: %s", s.testQueueURL)
		}
	}

	s.RouterSuite.TearDownTest()
}

// waitForSQSListenerReady waits for the SQS listener to be running
func (s *SQSSuite) waitForSQSListenerReady() {
	if s.sqsListenerService == nil {
		return
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		stats := s.sqsListenerService.GetStats()
		if stats.Running {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// extractQueueNameFromURL extracts the queue name from a LocalStack SQS URL
// Example: http://localhost:4566/000000000000/test-claim-12345 -> test-claim-12345
func (s *SQSSuite) extractQueueNameFromURL(queueURL string) string {
	// Split by '/' and get the last part
	parts := strings.Split(queueURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "test-queue"
}

// sendExternalMessage simulates an external service sending a message directly to SQS
// Returns the SQS message ID for tracking
func (s *SQSSuite) sendExternalMessage(ctx context.Context, message map[string]interface{}) (string, error) {
	queueURL := s.resource.Config.AwsConfig.Sqs.QueueURLs.SqsScheduledJobQueue
	messageBody, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	// Create direct SQS client to simulate external service
	sqsClient, err := sqs.NewClient(ctx, sqs.Config{
		Region:   s.resource.Config.AwsConfig.Region,
		Endpoint: s.resource.Config.AwsConfig.Endpoint,
		QueueURLs: sqs.QueueURLs{
			SqsScheduledJobQueue: queueURL,
		},
	}, s.resource.Logger)
	if err != nil {
		return "", err
	}

	// Send message as external service would (no JobID - external service doesn't know our internal IDs)
	output, err := sqsClient.SendMessage(ctx, queueURL, string(messageBody), map[string]string{
		"JobType": message["type"].(string),
		"Source":  "external_service",
	})
	if err != nil {
		return "", err
	}

	return *output.MessageId, nil
}

func (s *SQSSuite) TestSQSJobEnqueueAndProcess() {
	if s.sqsQueue == nil {
		s.T().Skip("SQS not available for testing")
	}

	// Test true event-driven SQS flow: External Service → SQS Queue → SQS Listener → SQS Processor → Database Job Creation → Job Handler → Status Updates
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	// Step 1: External service sends message to SQS queue (external service → SQS queue)
	// External services don't know our internal job IDs - they send their own data
	// Follow the strict sample format: {"type":"claim","payload":{"user":"a"}}
	// Using structured payload for better type safety
	claimPayload := sqs.ClaimPayload{
		User:   "external-service-user",
		Amount: 1000.0,
	}

	externalMessage := sqs.SqsMessage{
		Type:    "claim",
		Payload: claimPayload,
	}

	// Convert structured message to map for sending
	messageMap := map[string]interface{}{
		"type":    externalMessage.Type,
		"payload": externalMessage.Payload,
	}

	sqsMessageID, err := s.sendExternalMessage(ctx, messageMap)
	s.r.NoError(err)
	s.T().Logf("✓ External service sent message to SQS (message ID: %s)", sqsMessageID)

	// Steps 2-6: Wait for our system to process the external message
	// SQS Listener → SQS Processor → Database Job Creation → Job Handler → Status Updates
	var processedJob *entity.Job
	s.a.Eventually(func() bool {
		jobEntity, err := s.repositories.JobRepository.GetBySQSMessageID(ctx, sqsMessageID)
		if err != nil {
			return false // Job not created yet
		}
		processedJob = jobEntity
		return jobEntity.Status == job.Completed
	}, 15*time.Second, 500*time.Millisecond, "Job should be created and completed from external SQS message")

	// Step 7: Verify complete event-driven flow
	s.r.NotNil(processedJob, "Job entity should exist after processing")
	s.a.Equal(job.Completed, processedJob.Status, "Job should be completed")
	s.a.Equal("external-service-user", processedJob.Payload["user"], "Job payload user should match external message")
	s.a.Equal(1000.0, processedJob.Payload["amount"], "Job payload amount should match external message")

	// Verify SQS metadata is stored
	sqsMetadata, exists := processedJob.Payload["_sqs_metadata"]
	s.a.True(exists, "SQS metadata should be stored in job payload")
	s.a.Equal(sqsMessageID, sqsMetadata.(map[string]interface{})["sqs_message_id"], "SQS message ID should match")

	s.T().Logf("✅ Event-driven flow: External Service → SQS → Our System → Job Status: %s", processedJob.Status)
}
