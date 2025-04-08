package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/command"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"go.uber.org/zap"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
)

// EventHandlerFactory is a factory for creating event handlers
type EventHandlerFactory struct {
	sess          *session.Session
	lambdaClient  *lambda.Lambda
	lambdaFnName  string
	dynamoUpdater *DynamoDBUpdater
}

// DynamoDBUpdater encapsulates DynamoDB update logic
type DynamoDBUpdater struct {
	// In a real implementation, this would include DynamoDB client
	// and methods to update read models based on events
}

// NewEventHandlerFactory creates a new event handler factory
func NewEventHandlerFactory(sess *session.Session, lambdaFnName string) *EventHandlerFactory {
	return &EventHandlerFactory{
		sess:          sess,
		lambdaClient:  lambda.New(sess),
		lambdaFnName:  lambdaFnName,
		dynamoUpdater: &DynamoDBUpdater{},
	}
}

// HandleAlbumCreated handles album created events
func (f *EventHandlerFactory) HandleAlbumCreated(ctx context.Context, event *command.DomainEvent) error {
	logger := logging.GetLogger()
	logger.Info("Handling album created event", zap.String("eventId", event.ID))

	// Parse the event payload
	var albumCreated command.AlbumCreatedEvent
	if err := parseEventPayload(event, &albumCreated); err != nil {
		return err
	}

	// In a real implementation, this would:
	// 1. Update the read model in DynamoDB
	// 2. Trigger any additional processing

	return nil
}

// HandleAlbumLiked handles album liked events
func (f *EventHandlerFactory) HandleAlbumLiked(ctx context.Context, event *command.DomainEvent) error {
	logger := logging.GetLogger()
	logger.Info("Handling album liked event", zap.String("eventId", event.ID))

	// Parse the event payload
	var albumReview command.AlbumReviewEvent
	if err := parseEventPayload(event, &albumReview); err != nil {
		return err
	}

	// In a real implementation, this would update like counts in DynamoDB

	return nil
}

// HandleAlbumDisliked handles album disliked events
func (f *EventHandlerFactory) HandleAlbumDisliked(ctx context.Context, event *command.DomainEvent) error {
	logger := logging.GetLogger()
	logger.Info("Handling album disliked event", zap.String("eventId", event.ID))

	// Parse the event payload
	var albumReview command.AlbumReviewEvent
	if err := parseEventPayload(event, &albumReview); err != nil {
		return err
	}

	// In a real implementation, this would update dislike counts in DynamoDB

	return nil
}

// HandleImageUploaded handles image uploaded events
func (f *EventHandlerFactory) HandleImageUploaded(ctx context.Context, event *command.DomainEvent) error {
	logger := logging.GetLogger()
	logger.Info("Handling image uploaded event", zap.String("eventId", event.ID))

	// Parse the event payload
	var imageUploaded command.ImageUploadedEvent
	if err := parseEventPayload(event, &imageUploaded); err != nil {
		return err
	}

	// Only invoke Lambda if image processing is required
	if !imageUploaded.RequiresProcessing {
		logger.Info("Image does not require processing, skipping Lambda invocation")
		return nil
	}

	// Invoke Lambda function for image processing
	payload, err := json.Marshal(imageUploaded)
	if err != nil {
		return fmt.Errorf("failed to marshal Lambda payload: %w", err)
	}

	_, err = f.lambdaClient.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(f.lambdaFnName),
		Payload:      payload,
	})

	if err != nil {
		return fmt.Errorf("failed to invoke Lambda function: %w", err)
	}

	logger.Info("Successfully invoked Lambda function for image processing",
		zap.String("albumId", imageUploaded.AlbumID),
		zap.String("imageId", imageUploaded.ImageID),
	)

	return nil
}

// RegisterHandlers registers all event handlers with a consumer
func (f *EventHandlerFactory) RegisterHandlers(consumer *Consumer) {
	consumer.RegisterHandler(command.EventTypeAlbumCreated, f.HandleAlbumCreated)
	consumer.RegisterHandler(command.EventTypeAlbumLiked, f.HandleAlbumLiked)
	consumer.RegisterHandler(command.EventTypeAlbumDisliked, f.HandleAlbumDisliked)
	consumer.RegisterHandler(command.EventTypeImageUploaded, f.HandleImageUploaded)
}

// parseEventPayload parses the event payload into a specific type
func parseEventPayload(event *command.DomainEvent, dest interface{}) error {
	// If payload is already the correct type, just copy it
	if payload, ok := event.Payload.(map[string]interface{}); ok {
		// Convert payload map to JSON
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}

		// Unmarshal JSON to destination type
		if err := json.Unmarshal(payloadBytes, dest); err != nil {
			return fmt.Errorf("failed to unmarshal event payload: %w", err)
		}
	} else {
		return fmt.Errorf("unexpected payload type: %T", event.Payload)
	}

	return nil
}

// Config contains configuration for the SQS handler
type Config struct {
	Region       string
	Endpoint     string
	QueueName    string
	QueueURL     string
	LambdaFnName string
}
