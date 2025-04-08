package sns

import (
	"context"
	"fmt"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/command"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// EventBus is an implementation of the event bus using SNS
type EventBus struct {
	publisher *Publisher
}

// NewEventBus creates a new SNS event bus
func NewEventBus(sess *session.Session, topicARN string) *EventBus {
	return &EventBus{
		publisher: NewPublisher(sess, topicARN),
	}
}

// PublishEvent publishes a domain event
func (eb *EventBus) PublishEvent(ctx context.Context, event *command.DomainEvent) error {
	return eb.publisher.Publish(ctx, event)
}

// PublishAlbumCreated publishes an album created event
func (eb *EventBus) PublishAlbumCreated(ctx context.Context, event *command.AlbumCreatedEvent) error {
	domainEvent := command.NewDomainEvent(
		event.AlbumID,
		command.EventTypeAlbumCreated,
		event,
		map[string]string{
			"albumId": event.AlbumID,
		},
	)
	return eb.PublishEvent(ctx, domainEvent)
}

// PublishAlbumReview publishes an album review event (like/dislike)
func (eb *EventBus) PublishAlbumReview(ctx context.Context, event *command.AlbumReviewEvent) error {
	eventType := command.EventTypeAlbumLiked
	if !event.Liked {
		eventType = command.EventTypeAlbumDisliked
	}

	domainEvent := command.NewDomainEvent(
		event.AlbumID,
		eventType,
		event,
		map[string]string{
			"albumId": event.AlbumID,
			"liked":   fmt.Sprintf("%t", event.Liked),
		},
	)
	return eb.PublishEvent(ctx, domainEvent)
}

// PublishImageUploaded publishes an image uploaded event
func (eb *EventBus) PublishImageUploaded(ctx context.Context, event *command.ImageUploadedEvent) error {
	domainEvent := command.NewDomainEvent(
		event.AlbumID,
		command.EventTypeImageUploaded,
		event,
		map[string]string{
			"albumId": event.AlbumID,
			"imageId": event.ImageID,
		},
	)
	return eb.PublishEvent(ctx, domainEvent)
}

// Config returns the configuration for this event bus
type Config struct {
	Region          string
	Endpoint        string
	TopicName       string
	TopicARN        string
	CreateIfMissing bool
}

// NewEventBusFromConfig creates a new event bus from configuration
func NewEventBusFromConfig(ctx context.Context, sess *session.Session, config *Config) (*EventBus, error) {
	// Create SNS client
	snsClient := sns.New(sess)

	// Set endpoint for local development
	if config.Endpoint != "" {
		snsClient.Endpoint = config.Endpoint
	}

	// Create publisher
	publisher := &Publisher{
		client:   snsClient,
		topicARN: config.TopicARN,
	}

	// Create topic if needed
	if config.CreateIfMissing && config.TopicARN == "" {
		topicARN, err := publisher.CreateTopicIfNotExists(ctx, config.TopicName)
		if err != nil {
			return nil, err
		}
		publisher.topicARN = topicARN
	}

	return &EventBus{
		publisher: publisher,
	}, nil
}
