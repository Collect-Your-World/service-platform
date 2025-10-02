package sqs

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/database/entity"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

// MessageHandler defines the interface for handling SQS messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, job *entity.Job) error
}

// JobManagerBridge bridges SQS messages to the job manager
type JobManagerBridge struct {
	jobManager interface{}
	logger     *zap.Logger
}

// NewJobManagerBridge creates a new bridge processor
func NewJobManagerBridge(jobManager interface{}, logger *zap.Logger) *JobManagerBridge {
	return &JobManagerBridge{
		jobManager: jobManager,
		logger:     logger.With(zap.String("component", "sqs_job_manager_bridge")),
	}
}

// HandleMessage implements the MessageHandler interface for SQS listener
func (b *JobManagerBridge) HandleMessage(ctx context.Context, sqsEvent *entity.Job) error {
	jobLogger := b.logger.With(
		zap.String("event_type", sqsEvent.Type),
		zap.String("sqs_event_id", sqsEvent.ID))

	jobLogger.Info("Processing SQS event, creating job via job manager")

	workerJobType := b.convertEventTypeToJobType(sqsEvent.Type)

	payload := make(map[string]interface{})
	if sqsEvent.Payload != nil {
		for k, v := range sqsEvent.Payload {
			payload[k] = v
		}
	}

	payload["_sqs_source"] = map[string]interface{}{
		"original_event_type": sqsEvent.Type,
		"sqs_event_id":        sqsEvent.ID,
	}

	managerCreateJobRequestType := reflect.TypeOf(struct {
		Type        string                 `json:"type"`
		Priority    job.Priority           `json:"priority"`
		Payload     map[string]interface{} `json:"payload"`
		MaxAttempts int                    `json:"max_attempts,omitempty"`
		ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	}{})

	createJobReqValue := reflect.New(managerCreateJobRequestType).Elem()
	createJobReqValue.FieldByName("Type").SetString(string(workerJobType))
	createJobReqValue.FieldByName("Priority").Set(reflect.ValueOf(sqsEvent.Priority))
	createJobReqValue.FieldByName("Payload").Set(reflect.ValueOf(payload))
	createJobReqValue.FieldByName("MaxAttempts").SetInt(int64(sqsEvent.MaxAttempts))

	jmValue := reflect.ValueOf(b.jobManager)
	createJobMethod := jmValue.MethodByName("CreateJob")

	if !createJobMethod.IsValid() {
		jobLogger.Error("Job manager does not have CreateJob method")
		return fmt.Errorf("job manager does not have CreateJob method")
	}

	results := createJobMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		createJobReqValue,
	})

	if len(results) != 2 {
		jobLogger.Error("Unexpected return values from CreateJob")
		return fmt.Errorf("unexpected return values from CreateJob")
	}

	if !results[1].IsNil() {
		err := results[1].Interface().(error)
		jobLogger.Error("Failed to create job via job manager", zap.Error(err))
		return fmt.Errorf("failed to create job via job manager: %w", err)
	}

	workerJob := results[0].Interface().(*entity.Job)

	jobLogger.Info("SQS event successfully processed via job manager",
		zap.String("worker_job_id", workerJob.ID),
		zap.String("worker_job_type", workerJob.Type))

	return nil
}

func (b *JobManagerBridge) convertEventTypeToJobType(eventType string) job.Type {
	switch eventType {
	case Claim.String():
		return job.InitClaim
	case KYCVerification.String():
		return job.KYCVerification
	default:
		b.logger.Warn("No job type mapping for event type", zap.String("event_type", eventType))
		return job.Type(eventType)
	}
}

// Listener manages SQS message polling and processing
type Listener struct {
	client      *Client
	config      Config
	logger      *zap.Logger
	handler     MessageHandler
	queueURLs   []string
	running     bool
	mu          sync.RWMutex
	stopCh      chan struct{}
	workerCount int
}

// ListenerConfig defines configuration for the SQS listener
type ListenerConfig struct {
	// Queue URLs to listen to (in priority order)
	QueueURLs []string

	// Number of concurrent workers per queue
	WorkerCount int

	// Handler for processing messages
	Handler MessageHandler
}

// NewListener creates a new SQS message listener
func NewListener(ctx context.Context, config Config, listenerConfig ListenerConfig, logger *zap.Logger) (*Listener, error) {
	client, err := NewClient(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQS client: %w", err)
	}

	workerCount := listenerConfig.WorkerCount
	if workerCount <= 0 {
		workerCount = 1
	}

	return &Listener{
		client:      client,
		config:      config,
		logger:      logger.With(zap.String("component", "sqs_listener")),
		handler:     listenerConfig.Handler,
		queueURLs:   listenerConfig.QueueURLs,
		stopCh:      make(chan struct{}),
		workerCount: workerCount,
	}, nil
}

// Start begins listening for messages from configured SQS queues
func (l *Listener) Start(ctx context.Context) error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return fmt.Errorf("listener is already running")
	}
	l.running = true
	l.mu.Unlock()

	l.logger.Info("Starting SQS listener",
		zap.Strings("queue_urls", l.queueURLs),
		zap.Int("worker_count", l.workerCount))

	// Start workers for each queue
	var wg sync.WaitGroup
	for _, queueURL := range l.queueURLs {
		if queueURL == "" {
			l.logger.Warn("Skipping empty queue URL")
			continue
		}

		for i := 0; i < l.workerCount; i++ {
			wg.Add(1)
			go func(queueURL string, workerID int) {
				defer wg.Done()
				l.runWorker(ctx, queueURL, workerID)
			}(queueURL, i)
		}
	}

	// Wait for context cancellation or stop signal
	select {
	case <-ctx.Done():
		l.logger.Info("Context cancelled, stopping SQS listener")
	case <-l.stopCh:
		l.logger.Info("Stop signal received, stopping SQS listener")
	}

	// Signal all workers to stop
	l.mu.Lock()
	l.running = false
	l.mu.Unlock()

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		l.logger.Info("All SQS workers stopped gracefully")
	case <-time.After(30 * time.Second):
		l.logger.Warn("Timeout waiting for SQS workers to stop")
	}

	return nil
}

// Stop gracefully stops the listener
func (l *Listener) Stop(ctx context.Context) error {
	l.mu.RLock()
	running := l.running
	l.mu.RUnlock()

	if !running {
		return nil
	}

	l.logger.Info("Stopping SQS listener")

	select {
	case l.stopCh <- struct{}{}:
	default:
	}

	return nil
}

// runWorker runs a single worker that polls a specific queue
func (l *Listener) runWorker(ctx context.Context, queueURL string, workerID int) {
	workerLogger := l.logger.With(
		zap.String("queue_url", queueURL),
		zap.Int("worker_id", workerID))

	workerLogger.Info("Starting SQS worker")

	for {
		// Check if we should stop
		l.mu.RLock()
		running := l.running
		l.mu.RUnlock()

		if !running {
			workerLogger.Info("Worker stopping")
			return
		}

		// Check context
		select {
		case <-ctx.Done():
			workerLogger.Info("Worker context cancelled")
			return
		default:
		}

		// Poll for messages
		if err := l.pollMessages(ctx, queueURL, workerLogger); err != nil {
			workerLogger.Error("Error polling messages", zap.Error(err))

			// Brief pause before retrying to avoid rapid polling on errors
			select {
			case <-time.After(l.config.Polling.PollingInterval):
			case <-ctx.Done():
				return
			}
		}
	}
}

// pollMessages polls for messages from the specified queue
func (l *Listener) pollMessages(ctx context.Context, queueURL string, logger *zap.Logger) error {
	output, err := l.client.ReceiveMessages(ctx, queueURL)
	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	if len(output.Messages) == 0 {
		logger.Debug("No messages received")
		return nil
	}

	logger.Debug("Received messages", zap.Int("count", len(output.Messages)))

	// Process each message
	for _, message := range output.Messages {
		if err := l.processMessage(ctx, queueURL, message, logger); err != nil {
			logger.Error("Failed to process message",
				zap.String("message_id", getMessageID(message)),
				zap.Error(err))
		}
	}

	return nil
}

// processMessage processes a single SQS message
func (l *Listener) processMessage(ctx context.Context, queueURL string, message types.Message, logger *zap.Logger) error {
	messageID := getMessageID(message)
	logger = logger.With(zap.String("message_id", messageID))

	logger.Debug("Processing message")

	sqsMessage, err := l.parseMessage(ctx, queueURL, message, logger)
	if err != nil {
		return err
	}

	eventJob, err := l.createJobFromMessage(sqsMessage, messageID, queueURL)
	if err != nil {
		_ = l.deleteMessage(ctx, queueURL, *message.ReceiptHandle, logger)
		return err
	}

	logger = logger.With(
		zap.String("job_id", eventJob.ID),
		zap.String("job_type", eventJob.Type))

	logger.Info("Processing job from SQS")

	return l.handleJobAndCleanup(ctx, queueURL, message, eventJob, logger)
}

func (l *Listener) parseMessage(ctx context.Context, queueURL string, message types.Message, logger *zap.Logger) (*SqsMessage, error) {
	var sqsMessage SqsMessage
	if err := json.Unmarshal([]byte(*message.Body), &sqsMessage); err != nil {
		logger.Error("Failed to unmarshal SQS message", zap.Error(err))
		_ = l.deleteMessage(ctx, queueURL, *message.ReceiptHandle, logger)
		return nil, fmt.Errorf("failed to unmarshal SQS message: %w", err)
	}

	if err := l.validateMessage(&sqsMessage); err != nil {
		logger.Error("Invalid message", zap.Error(err))
		_ = l.deleteMessage(ctx, queueURL, *message.ReceiptHandle, logger)
		return nil, err
	}

	return &sqsMessage, nil
}

func (l *Listener) validateMessage(sqsMessage *SqsMessage) error {
	if sqsMessage.Type == "" {
		return fmt.Errorf("message missing required 'type' field")
	}

	if sqsMessage.Payload == nil {
		return fmt.Errorf("message missing required 'payload' field")
	}

	return nil
}

func (l *Listener) createJobFromMessage(sqsMessage *SqsMessage, messageID, queueURL string) (*entity.Job, error) {
	payloadMap, err := l.convertPayloadToMap(sqsMessage.Payload)
	if err != nil {
		return nil, err
	}

	eventJob := &entity.Job{
		Type:    sqsMessage.Type,
		Payload: payloadMap,
	}

	l.addSQSMetadata(eventJob, messageID, queueURL)
	return eventJob, nil
}

func (l *Listener) convertPayloadToMap(payload interface{}) (map[string]interface{}, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return payloadMap, nil
}

func (l *Listener) addSQSMetadata(job *entity.Job, messageID, queueURL string) {
	job.Payload["_sqs_metadata"] = map[string]interface{}{
		"sqs_message_id": messageID,
		"sqs_queue_url":  queueURL,
	}
}

func (l *Listener) handleJobAndCleanup(ctx context.Context, queueURL string, message types.Message, eventJob *entity.Job, logger *zap.Logger) error {
	if err := l.handler.HandleMessage(ctx, eventJob); err != nil {
		logger.Error("Failed to handle job", zap.Error(err))

		if l.shouldRetry(eventJob, err) {
			logger.Info("Job will be retried automatically by SQS")
			return nil
		}

		_ = l.deleteMessage(ctx, queueURL, *message.ReceiptHandle, logger)
		logger.Info("Deleted failed message from SQS")
		return err
	}

	if err := l.deleteMessage(ctx, queueURL, *message.ReceiptHandle, logger); err != nil {
		logger.Error("Failed to delete processed message", zap.Error(err))
		return fmt.Errorf("failed to delete processed message: %w", err)
	}

	logger.Info("Job processed successfully and message deleted")
	return nil
}

func (l *Listener) deleteMessage(ctx context.Context, queueURL, receiptHandle string, logger *zap.Logger) error {
	if err := l.client.DeleteMessage(ctx, queueURL, receiptHandle); err != nil {
		logger.Error("Failed to delete message", zap.Error(err))
		return err
	}
	return nil
}

// shouldRetry determines if a job should be retried based on the error and job configuration
func (l *Listener) shouldRetry(job *entity.Job, err error) bool {
	// Check if job has remaining attempts
	if job.Attempts >= job.MaxAttempts {
		l.logger.Info("Job exceeded max attempts",
			zap.String("job_id", job.ID),
			zap.Int("attempts", job.Attempts),
			zap.Int("max_attempts", job.MaxAttempts))
		return false
	}

	// Check if error is retryable (this could be more sophisticated)
	// For now, we'll retry most errors except for specific types
	switch err.Error() {
	case "invalid_payload", "malformed_data", "authentication_failed":
		// Don't retry these types of errors
		return false
	default:
		// Retry other errors
		return true
	}
}

// IsRunning returns whether the listener is currently running
func (l *Listener) IsRunning() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.running
}

// GetStats returns listener statistics
func (l *Listener) GetStats() ListenerStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return ListenerStats{
		Running:     l.running,
		QueueURLs:   l.queueURLs,
		WorkerCount: l.workerCount,
	}
}

// ListenerStats contains listener statistics
type ListenerStats struct {
	Running     bool     `json:"running"`
	QueueURLs   []string `json:"queue_urls"`
	WorkerCount int      `json:"worker_count"`
}

// Helper functions

func getMessageID(message types.Message) string {
	if message.MessageId != nil {
		return *message.MessageId
	}
	return "unknown"
}

// getMessageAttribute safely gets a message attribute value
func getMessageAttribute(message types.Message, key string) string {
	if message.MessageAttributes == nil {
		return ""
	}

	attr, exists := message.MessageAttributes[key]
	if !exists || attr.StringValue == nil {
		return ""
	}

	return *attr.StringValue
}
