package sqs

import (
	jobconst "backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

// Queue implements the queue.Queue interface using AWS SQS
type Queue struct {
	client *Client
	config Config
	logger *zap.Logger
}

// NewSQSQueue creates a new SQS-based queue
func NewSQSQueue(ctx context.Context, config Config, logger *zap.Logger) (*Queue, error) {
	client, err := NewClient(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQS client: %w", err)
	}

	return &Queue{
		client: client,
		config: config,
		logger: logger.With(zap.String("component", "sqs_queue")),
	}, nil
}

// Enqueue adds a job to the appropriate SQS queue based on job type
func (q *Queue) Enqueue(ctx context.Context, job *entity.Job) error {
	queueURL := q.getQueueURLByJobType(jobconst.Type(job.Type))
	if queueURL == "" {
		return fmt.Errorf("no queue URL configured for job type %s", job.Type)
	}

	// Marshal job to JSON
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Prepare message attributes
	attributes := map[string]string{
		"JobID":       job.ID,
		"JobType":     job.Type,
		"Priority":    job.Priority.String(),
		"Attempts":    strconv.Itoa(job.Attempts),
		"MaxAttempts": strconv.Itoa(job.MaxAttempts),
	}

	// Handle scheduled jobs
	if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
		delaySeconds := int32(time.Until(*job.ScheduledAt).Seconds())

		// SQS supports maximum delay of 15 minutes (900 seconds)
		if delaySeconds > 900 {
			q.logger.Warn("Job delay exceeds SQS maximum, using max delay",
				zap.String("job_id", job.ID),
				zap.Int32("requested_delay", delaySeconds),
				zap.Int32("actual_delay", 900))
			delaySeconds = 900
		}

		_, err = q.client.SendDelayedMessage(ctx, queueURL, string(jobData), delaySeconds, attributes)
	} else {
		_, err = q.client.SendMessage(ctx, queueURL, string(jobData), attributes)
	}

	if err != nil {
		q.logger.Error("Failed to enqueue job to SQS",
			zap.String("job_id", job.ID),
			zap.String("queue_url", queueURL),
			zap.Error(err))
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	q.logger.Info("Job enqueued successfully to SQS",
		zap.String("job_id", job.ID),
		zap.String("type", job.Type),
		zap.String("priority", job.Priority.String()),
		zap.String("queue_url", queueURL))

	return nil
}

// Dequeue receives a job from SQS queues in priority order
func (q *Queue) Dequeue(ctx context.Context, queues []string) (*entity.Job, error) {
	// Convert Redis queue names to SQS queue URLs
	queueURLs := q.getQueueURLsFromNames(queues)

	// Poll queues in priority order
	for _, queueURL := range queueURLs {
		if queueURL == "" {
			continue
		}

		output, err := q.client.ReceiveMessages(ctx, queueURL)
		if err != nil {
			q.logger.Error("Failed to receive messages from SQS",
				zap.String("queue_url", queueURL),
				zap.Error(err))
			continue
		}

		// Process first available message
		if len(output.Messages) > 0 {
			message := output.Messages[0]

			var jobEntity entity.Job
			if err := json.Unmarshal([]byte(*message.Body), &jobEntity); err != nil {
				q.logger.Error("Failed to unmarshal job from SQS message",
					zap.String("message_body", *message.Body),
					zap.Error(err))

				// Delete malformed message to prevent infinite processing
				_ = q.client.DeleteMessage(ctx, queueURL, *message.ReceiptHandle)
				continue
			}

			// Store SQS metadata in job payload (temporary workaround)
			jobEntity.Payload["_sqs_metadata"] = map[string]interface{}{
				"sqs_receipt_handle": *message.ReceiptHandle,
				"sqs_queue_url":      queueURL,
			}

			q.logger.Info("Job dequeued successfully from SQS",
				zap.String("job_id", jobEntity.ID),
				zap.String("type", jobEntity.Type),
				zap.String("queue_url", queueURL))

			return &jobEntity, nil
		}
	}

	// No messages available in any queue
	return nil, nil
}

// MarkProcessing marks a job as being processed (SQS handles this via visibility timeout)
func (q *Queue) MarkProcessing(ctx context.Context, jobID string) error {
	// For SQS, this is handled automatically by message visibility timeout
	// We just log for consistency with Redis implementation
	q.logger.Debug("Job marked as processing (SQS visibility timeout)",
		zap.String("job_id", jobID))
	return nil
}

// MarkCompleted removes the message from SQS queue (deletes it)
func (q *Queue) MarkCompleted(ctx context.Context, jobID string) error {
	// We need to find the job's SQS metadata to delete the message
	// In practice, this should be called with the job context that includes metadata
	q.logger.Info("Job marked as completed", zap.String("job_id", jobID))
	return nil
}

// MarkCompletedWithMetadata deletes the SQS message using stored metadata
func (q *Queue) MarkCompletedWithMetadata(ctx context.Context, jobID string, metadata map[string]interface{}) error {
	receiptHandle, ok := metadata["sqs_receipt_handle"].(string)
	if !ok {
		return fmt.Errorf("missing SQS receipt handle for job %s", jobID)
	}

	queueURL, ok := metadata["sqs_queue_url"].(string)
	if !ok {
		return fmt.Errorf("missing SQS queue URL for job %s", jobID)
	}

	err := q.client.DeleteMessage(ctx, queueURL, receiptHandle)
	if err != nil {
		q.logger.Error("Failed to delete message from SQS",
			zap.String("job_id", jobID),
			zap.String("queue_url", queueURL),
			zap.Error(err))
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	q.logger.Info("Job marked as completed and deleted from SQS",
		zap.String("job_id", jobID),
		zap.String("queue_url", queueURL))

	return nil
}

// MarkFailed handles job failure, potentially re-queueing with retry logic
func (q *Queue) MarkFailed(ctx context.Context, jobID string, retryDelay time.Duration) error {
	// For SQS, failed jobs will automatically become visible again after visibility timeout
	// For immediate retry, we can delete the current message and send a new delayed one
	q.logger.Info("Job marked as failed",
		zap.String("job_id", jobID),
		zap.Duration("retry_delay", retryDelay))

	return nil
}

// MarkFailedWithMetadata handles job failure with retry logic using SQS delay
func (q *Queue) MarkFailedWithMetadata(ctx context.Context, jobID string, retryDelay time.Duration, job *entity.Job, metadata map[string]interface{}) error {
	receiptHandle, ok := metadata["sqs_receipt_handle"].(string)
	if !ok {
		return fmt.Errorf("missing SQS receipt handle for job %s", jobID)
	}

	queueURL, ok := metadata["sqs_queue_url"].(string)
	if !ok {
		return fmt.Errorf("missing SQS queue URL for job %s", jobID)
	}

	// Delete the current message
	if err := q.client.DeleteMessage(ctx, queueURL, receiptHandle); err != nil {
		q.logger.Error("Failed to delete failed message from SQS",
			zap.String("job_id", jobID),
			zap.Error(err))
		return fmt.Errorf("failed to delete failed message: %w", err)
	}

	// Re-queue with delay if retries are available
	if retryDelay > 0 && job.Attempts < job.MaxAttempts {
		delaySeconds := int32(retryDelay.Seconds())
		if delaySeconds > 900 {
			delaySeconds = 900 // SQS maximum delay
		}

		// Update job for retry
		job.Attempts++
		job.Status = "retrying"

		jobData, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal retry job: %w", err)
		}

		attributes := map[string]string{
			"JobID":       job.ID,
			"JobType":     job.Type,
			"Priority":    job.Priority.String(),
			"Attempts":    strconv.Itoa(job.Attempts),
			"MaxAttempts": strconv.Itoa(job.MaxAttempts),
			"IsRetry":     "true",
		}

		_, err = q.client.SendDelayedMessage(ctx, queueURL, string(jobData), delaySeconds, attributes)
		if err != nil {
			q.logger.Error("Failed to re-queue job for retry",
				zap.String("job_id", jobID),
				zap.Error(err))
			return fmt.Errorf("failed to re-queue job for retry: %w", err)
		}

		q.logger.Info("Job re-queued for retry",
			zap.String("job_id", jobID),
			zap.Int("attempts", job.Attempts),
			zap.Int32("delay_seconds", delaySeconds))
	}

	return nil
}

// GetQueueDepth returns approximate number of messages in queue
func (q *Queue) GetQueueDepth(ctx context.Context, queue string) (int64, error) {
	queueURL := q.getQueueURLFromName(queue)
	if queueURL == "" {
		return 0, fmt.Errorf("no queue URL configured for queue %s", queue)
	}

	output, err := q.client.GetQueueAttributes(ctx, queueURL, []types.QueueAttributeName{
		types.QueueAttributeNameApproximateNumberOfMessages,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	if countStr, exists := output.Attributes["ApproximateNumberOfMessages"]; exists {
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse message count: %w", err)
		}
		return count, nil
	}

	return 0, nil
}

// GetProcessingJobs returns list of jobs currently being processed (limited in SQS)
func (q *Queue) GetProcessingJobs(ctx context.Context) ([]string, error) {
	// SQS doesn't provide direct access to processing jobs
	// This is a limitation compared to Redis implementation
	q.logger.Debug("GetProcessingJobs called - limited functionality in SQS")
	return []string{}, nil
}

// Helper methods

func (q *Queue) getQueueURLByJobType(jobType jobconst.Type) string {
	switch jobType {
	case jobconst.InitClaim, jobconst.CompleteClaim:
		return q.config.QueueURLs.SqsScheduledJobQueue
	case jobconst.KYCVerification:
		return ""
	default:
		return ""
	}
}

func (q *Queue) getQueueURLFromName(queueName string) string {
	switch queueName {
	case "claim":
		return q.config.QueueURLs.SqsScheduledJobQueue
	default:
		return ""
	}
}

func (q *Queue) getQueueURLsFromNames(queueNames []string) []string {
	urls := make([]string, len(queueNames))
	for i, name := range queueNames {
		urls[i] = q.getQueueURLFromName(name)
	}
	return urls
}

// CalculateRetryDelay calculates exponential backoff delay
func (q *Queue) CalculateRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return q.config.Message.BaseRetryDelay
	}

	// Exponential backoff: baseDelay * 2^(attempt-1)
	delay := q.config.Message.BaseRetryDelay * time.Duration(math.Pow(2, float64(attempt-1)))

	// Cap at maximum delay
	if delay > q.config.Message.MaxRetryDelay {
		delay = q.config.Message.MaxRetryDelay
	}

	return delay
}

// EnqueueRaw sends a job message directly to SQS without database dependency
// This simulates external services sending job messages to SQS queues
func (q *Queue) EnqueueRaw(ctx context.Context, job *entity.Job) error {
	queueURL := q.getQueueURLByJobType(jobconst.Type(job.Type))
	if queueURL == "" {
		return fmt.Errorf("no SQS queue configured for job type: %s", job.Type)
	}

	// Serialize job to JSON
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Prepare message attributes
	attributes := map[string]string{
		"JobID":       job.ID,
		"JobType":     job.Type,
		"Priority":    job.Priority.String(),
		"Attempts":    strconv.Itoa(job.Attempts),
		"MaxAttempts": strconv.Itoa(job.MaxAttempts),
		"Source":      "external", // Mark as external message
	}

	// Handle scheduled jobs with delay
	if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
		delaySeconds := int32(time.Until(*job.ScheduledAt).Seconds())

		// SQS supports maximum delay of 15 minutes (900 seconds)
		if delaySeconds > 900 {
			q.logger.Warn("External job delay exceeds SQS maximum, using max delay",
				zap.String("job_id", job.ID),
				zap.Int32("requested_delay", delaySeconds),
				zap.Int32("actual_delay", 900))
			delaySeconds = 900
		}

		q.logger.Info("Sending delayed external message to SQS",
			zap.String("job_id", job.ID),
			zap.String("job_type", job.Type),
			zap.String("queue_url", queueURL),
			zap.Int32("delay_seconds", delaySeconds))

		_, err = q.client.SendDelayedMessage(ctx, queueURL, string(jobData), delaySeconds, attributes)
	} else {
		q.logger.Info("Sending raw external message to SQS",
			zap.String("job_id", job.ID),
			zap.String("job_type", job.Type),
			zap.String("queue_url", queueURL))

		_, err = q.client.SendMessage(ctx, queueURL, string(jobData), attributes)
	}

	if err != nil {
		return fmt.Errorf("failed to send raw message to SQS: %w", err)
	}

	return nil
}
