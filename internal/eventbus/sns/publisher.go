package sns

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/command"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// Publisher publishes events to SNS
type Publisher struct {
	client   *sns.SNS
	topicARN string
}

// NewPublisher creates a new SNS publisher
func NewPublisher(sess *session.Session, topicARN string) *Publisher {
	return &Publisher{
		client:   sns.New(sess),
		topicARN: topicARN,
	}
}

// Publish publishes an event to the SNS topic
func (p *Publisher) Publish(ctx context.Context, event *command.DomainEvent) error {
	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create SNS message attributes for event type
	messageAttributes := map[string]*sns.MessageAttributeValue{
		"EventType": {
			DataType:    aws.String("String"),
			StringValue: aws.String(string(event.Type)),
		},
	}

	// Publish message
	_, err = p.client.PublishWithContext(ctx, &sns.PublishInput{
		TopicArn:               aws.String(p.topicARN),
		Message:                aws.String(string(eventJSON)),
		MessageAttributes:      messageAttributes,
		MessageGroupId:         aws.String(event.ID), // For FIFO topics
		MessageDeduplicationId: aws.String(event.ID), // For FIFO topics
	})

	if err != nil {
		return fmt.Errorf("failed to publish event to SNS: %w", err)
	}

	return nil
}

// PublishBatch publishes multiple events to the SNS topic in batch
func (p *Publisher) PublishBatch(ctx context.Context, events []*command.DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// CreateTopicIfNotExists creates the SNS topic if it doesn't exist
func (p *Publisher) CreateTopicIfNotExists(ctx context.Context, topicName string) (string, error) {
	// Check if we already have a topic ARN
	if p.topicARN != "" {
		return p.topicARN, nil
	}

	// Create topic
	result, err := p.client.CreateTopicWithContext(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})

	if err != nil {
		return "", fmt.Errorf("failed to create SNS topic: %w", err)
	}

	// Save topic ARN
	p.topicARN = *result.TopicArn

	return p.topicARN, nil
}

// SubscribeQueue subscribes an SQS queue to the SNS topic
func (p *Publisher) SubscribeQueue(ctx context.Context, queueARN string) (string, error) {
	// Subscribe queue to topic
	result, err := p.client.SubscribeWithContext(ctx, &sns.SubscribeInput{
		TopicArn: aws.String(p.topicARN),
		Protocol: aws.String("sqs"),
		Endpoint: aws.String(queueARN),
	})

	if err != nil {
		return "", fmt.Errorf("failed to subscribe queue to SNS topic: %w", err)
	}

	return *result.SubscriptionArn, nil
}
