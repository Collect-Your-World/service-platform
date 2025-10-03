package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

// Client wraps the AWS SQS client with additional functionality
type Client struct {
	sqs    *sqs.Client
	config Config
	logger *zap.Logger
}

// NewClient creates a new SQS client
func NewClient(ctx context.Context, config Config, logger *zap.Logger) (*Client, error) {
	var awsCfg aws.Config
	var err error

	if config.Endpoint != "" {
		// Configure for custom endpoint (e.g., LocalStack)
		logger.Info("Configuring SQS client for custom endpoint", zap.String("endpoint", config.Endpoint))
		awsCfg, err = awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(config.Region),
			awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "test",
					SecretAccessKey: "test",
					Source:          "CustomEndpoint",
				}, nil
			})),
			awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:               config.Endpoint,
						SigningRegion:     config.Region,
						HostnameImmutable: true,
					}, nil
				},
			)),
		)
	} else {
		// Load standard AWS configuration
		awsCfg, err = awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(config.Region),
		)
	}

	if err != nil {
		return nil, err
	}

	// Create SQS client
	sqsClient := sqs.NewFromConfig(awsCfg)

	return &Client{
		sqs:    sqsClient,
		config: config,
		logger: logger.With(zap.String("component", "sqs_client")),
	}, nil
}

// SendMessage sends a message to the specified queue
func (c *Client) SendMessage(ctx context.Context, queueURL, messageBody string, attributes map[string]string) (*sqs.SendMessageOutput, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(messageBody),
	}

	// Add message attributes if provided
	if len(attributes) > 0 {
		messageAttributes := make(map[string]types.MessageAttributeValue)
		for key, value := range attributes {
			messageAttributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(value),
			}
		}
		input.MessageAttributes = messageAttributes
	}

	c.logger.Debug("Sending message to SQS",
		zap.String("queue_url", queueURL),
		zap.Int("attributes_count", len(attributes)))

	return c.sqs.SendMessage(ctx, input)
}

// SendDelayedMessage sends a message with a delay
func (c *Client) SendDelayedMessage(ctx context.Context, queueURL, messageBody string, delaySeconds int32, attributes map[string]string) (*sqs.SendMessageOutput, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:     aws.String(queueURL),
		MessageBody:  aws.String(messageBody),
		DelaySeconds: delaySeconds,
	}

	// Add message attributes if provided
	if len(attributes) > 0 {
		messageAttributes := make(map[string]types.MessageAttributeValue)
		for key, value := range attributes {
			messageAttributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(value),
			}
		}
		input.MessageAttributes = messageAttributes
	}

	c.logger.Debug("Sending delayed message to SQS",
		zap.String("queue_url", queueURL),
		zap.Int32("delay_seconds", delaySeconds),
		zap.Int("attributes_count", len(attributes)))

	return c.sqs.SendMessage(ctx, input)
}

// ReceiveMessages receives messages from the specified queue
func (c *Client) ReceiveMessages(ctx context.Context, queueURL string) (*sqs.ReceiveMessageOutput, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: int32(c.config.Polling.MaxMessages),
		WaitTimeSeconds:     int32(c.config.Polling.WaitTimeSeconds),
		VisibilityTimeout:   int32(c.config.Polling.VisibilityTimeoutSeconds),
		MessageAttributeNames: []string{
			"All", // Receive all message attributes
		},
	}

	c.logger.Debug("Receiving messages from SQS",
		zap.String("queue_url", queueURL),
		zap.Int("max_messages", c.config.Polling.MaxMessages),
		zap.Int("wait_time", c.config.Polling.WaitTimeSeconds))

	return c.sqs.ReceiveMessage(ctx, input)
}

// DeleteMessage deletes a message from the queue
func (c *Client) DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	}

	c.logger.Debug("Deleting message from SQS",
		zap.String("queue_url", queueURL))

	_, err := c.sqs.DeleteMessage(ctx, input)
	return err
}

// ChangeMessageVisibility changes the visibility timeout of a message
func (c *Client) ChangeMessageVisibility(ctx context.Context, queueURL, receiptHandle string, visibilityTimeout int32) error {
	input := &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(queueURL),
		ReceiptHandle:     aws.String(receiptHandle),
		VisibilityTimeout: visibilityTimeout,
	}

	c.logger.Debug("Changing message visibility",
		zap.String("queue_url", queueURL),
		zap.Int32("visibility_timeout", visibilityTimeout))

	_, err := c.sqs.ChangeMessageVisibility(ctx, input)
	return err
}

// GetQueueAttributes gets queue attributes
func (c *Client) GetQueueAttributes(ctx context.Context, queueURL string, attributeNames []types.QueueAttributeName) (*sqs.GetQueueAttributesOutput, error) {
	input := &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: attributeNames,
	}

	return c.sqs.GetQueueAttributes(ctx, input)
}

// CreateQueue creates a new SQS queue
func (c *Client) CreateQueue(ctx context.Context, queueName string, attributes map[string]string) (*sqs.CreateQueueOutput, error) {
	input := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	}

	// Add queue attributes if provided
	if len(attributes) > 0 {
		input.Attributes = attributes
	}

	c.logger.Info("Creating SQS queue",
		zap.String("queue_name", queueName),
		zap.Int("attributes_count", len(attributes)))

	return c.sqs.CreateQueue(ctx, input)
}

// DeleteQueue deletes an SQS queue
func (c *Client) DeleteQueue(ctx context.Context, queueURL string) error {
	input := &sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueURL),
	}

	c.logger.Info("Deleting SQS queue",
		zap.String("queue_url", queueURL))

	_, err := c.sqs.DeleteQueue(ctx, input)
	return err
}
