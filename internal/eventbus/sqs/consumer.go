package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/command"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
)

// Consumer consumes events from SQS
type Consumer struct {
	client     *sqs.SQS
	queueURL   string
	handlers   map[command.EventType][]EventHandler
	maxRetries int
}

// EventHandler is a function that handles domain events
type EventHandler func(context.Context, *command.DomainEvent) error

// NewConsumer creates a new SQS consumer
func NewConsumer(sess *session.Session, queueURL string) *Consumer {
	return &Consumer{
		client:     sqs.New(sess),
		queueURL:   queueURL,
		handlers:   make(map[command.EventType][]EventHandler),
		maxRetries: 3,
	}
}

// RegisterHandler registers a handler for a specific event type
func (c *Consumer) RegisterHandler(eventType command.EventType, handler EventHandler) {
	if _, exists := c.handlers[eventType]; !exists {
		c.handlers[eventType] = []EventHandler{}
	}
	c.handlers[eventType] = append(c.handlers[eventType], handler)
}

// Start starts consuming messages from the queue
func (c *Consumer) Start(ctx context.Context) error {
	logger := logging.GetLogger()
	logger.Info("Starting SQS consumer", zap.String("queueURL", c.queueURL))

	for {
		select {
		case <-ctx.Done():
			logger.Info("SQS consumer stopped")
			return nil
		default:
			// Receive messages
			result, err := c.client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:              aws.String(c.queueURL),
				MaxNumberOfMessages:   aws.Int64(10),
				WaitTimeSeconds:       aws.Int64(20), // Long polling
				MessageAttributeNames: aws.StringSlice([]string{"All"}),
			})

			if err != nil {
				logger.Error("Failed to receive messages from SQS", zap.Error(err))
				time.Sleep(5 * time.Second) // Backoff before retrying
				continue
			}

			// Process messages
			for _, msg := range result.Messages {
				go c.processMessage(ctx, msg)
			}
		}
	}
}

// processMessage processes a single SQS message
func (c *Consumer) processMessage(ctx context.Context, msg *sqs.Message) {
	logger := logging.GetLogger()
	logger.Debug("Processing SQS message", zap.String("messageId", *msg.MessageId))

	// Parse message body as domain event
	var event command.DomainEvent
	if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
		logger.Error("Failed to unmarshal message body",
			zap.Error(err),
			zap.String("body", *msg.Body),
		)
		c.deleteMessage(ctx, msg)
		return
	}

	// Get event type from message attributes if available
	if attributeValue, ok := msg.MessageAttributes["EventType"]; ok {
		if attributeValue.StringValue != nil {
			event.Type = command.EventType(*attributeValue.StringValue)
		}
	}

	// Find handlers for this event type
	handlers, ok := c.handlers[event.Type]
	if !ok || len(handlers) == 0 {
		logger.Warn("No handlers registered for event type",
			zap.String("eventType", string(event.Type)),
		)
		c.deleteMessage(ctx, msg)
		return
	}

	// Process event with all registered handlers
	var handlerErrors []error
	for _, handler := range handlers {
		if err := handler(ctx, &event); err != nil {
			handlerErrors = append(handlerErrors, err)
			logger.Error("Handler failed to process event",
				zap.Error(err),
				zap.String("eventType", string(event.Type)),
				zap.String("eventId", event.ID),
			)
		}
	}

	// Delete message if all handlers succeeded or if we've exceeded retries
	if len(handlerErrors) == 0 || c.getApproximateReceiveCount(msg) > c.maxRetries {
		c.deleteMessage(ctx, msg)
	}
}

// deleteMessage deletes a message from the queue
func (c *Consumer) deleteMessage(ctx context.Context, msg *sqs.Message) {
	_, err := c.client.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		logging.GetLogger().Error("Failed to delete message from SQS",
			zap.Error(err),
			zap.String("messageId", *msg.MessageId),
		)
	}
}

// getApproximateReceiveCount gets the approximate number of times a message has been received
func (c *Consumer) getApproximateReceiveCount(msg *sqs.Message) int {
	if attributeValue, ok := msg.Attributes["ApproximateReceiveCount"]; ok {
		count := 0
		fmt.Sscanf(*attributeValue, "%d", &count)
		return count
	}
	return 0
}

// CreateQueueIfNotExists creates the SQS queue if it doesn't exist
func (c *Consumer) CreateQueueIfNotExists(ctx context.Context, queueName string) (string, error) {
	// Check if we already have a queue URL
	if c.queueURL != "" {
		return c.queueURL, nil
	}

	// Create queue
	result, err := c.client.CreateQueueWithContext(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]*string{
			"VisibilityTimeout":             aws.String("30"),
			"MessageRetentionPeriod":        aws.String("86400"), // 1 day
			"ReceiveMessageWaitTimeSeconds": aws.String("20"),    // Enable long polling
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to create SQS queue: %w", err)
	}

	// Save queue URL
	c.queueURL = *result.QueueUrl

	return c.queueURL, nil
}
