package review

import (
	"context"
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// MessageProcessor processes review messages from SQS
type MessageProcessor struct {
	repo            domain.ReviewRepository
	sqsClient       *sqs.Client
	queueURL        string
	maxMessages     int32
	waitTimeSeconds int32
	stopChan        chan struct{}
	metrics         struct {
		TotalMessages           int64
		ProcessedMessages       int64
		FailedMessages          int64
		AverageProcessingTimeMs int64
	}
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(repo domain.ReviewRepository, sqsClient *sqs.Client, queueURL string) *MessageProcessor {
	return &MessageProcessor{
		repo:            repo,
		sqsClient:       sqsClient,
		queueURL:        queueURL,
		maxMessages:     10,
		waitTimeSeconds: 5,
		stopChan:        make(chan struct{}),
	}
}

// Start begins processing messages
func (p *MessageProcessor) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-p.stopChan:
				return
			case <-ctx.Done():
				return
			default:
				if err := p.processMessages(ctx); err != nil {
					log.Printf("Error processing messages: %v", err)
					// Sleep briefly before retrying
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()
}

// Stop stops the message processor
func (p *MessageProcessor) Stop() {
	close(p.stopChan)
}

// processMessages retrieves and processes messages from SQS
func (p *MessageProcessor) processMessages(ctx context.Context) error {
	// Receive messages from SQS
	msgResult, err := p.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &p.queueURL,
		MaxNumberOfMessages: p.maxMessages,
		WaitTimeSeconds:     p.waitTimeSeconds,
	})

	if err != nil {
		return err
	}

	if len(msgResult.Messages) == 0 {
		return nil
	}

	atomic.AddInt64(&p.metrics.TotalMessages, int64(len(msgResult.Messages)))

	// Process each message
	for _, message := range msgResult.Messages {
		if err := p.handleMessage(ctx, message); err != nil {
			log.Printf("Error handling message: %v", err)
			atomic.AddInt64(&p.metrics.FailedMessages, 1)
			continue
		}

		atomic.AddInt64(&p.metrics.ProcessedMessages, 1)

		// Delete the processed message
		_, err := p.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      &p.queueURL,
			ReceiptHandle: message.ReceiptHandle,
		})

		if err != nil {
			log.Printf("Error deleting message: %v", err)
		}
	}

	return nil
}

// handleMessage processes a single SQS message
func (p *MessageProcessor) handleMessage(ctx context.Context, message types.Message) error {
	startTime := time.Now()

	var snsMessage struct {
		Message string `json:"Message"`
	}

	if err := json.Unmarshal([]byte(*message.Body), &snsMessage); err != nil {
		return err
	}

	var reviewMsg ReviewMessage
	if err := json.Unmarshal([]byte(snsMessage.Message), &reviewMsg); err != nil {
		return err
	}

	var err error

	// Process based on message type
	switch reviewMsg.Type {
	case AddLikeType:
		err = p.repo.AddLike(ctx, reviewMsg.AlbumID)
	case AddDislikeType:
		err = p.repo.AddDislike(ctx, reviewMsg.AlbumID)
	case DeleteType:
		err = p.repo.Delete(ctx, reviewMsg.AlbumID)
	default:
		log.Printf("Unknown message type: %s", reviewMsg.Type)
		return nil
	}

	// Update average processing time
	if err == nil {
		processingTimeMs := time.Since(startTime).Milliseconds()
		currentAvg := atomic.LoadInt64(&p.metrics.AverageProcessingTimeMs)
		processedCount := atomic.LoadInt64(&p.metrics.ProcessedMessages)

		// Calculate new average processing time
		if processedCount > 0 {
			newAvg := (currentAvg*(processedCount-1) + processingTimeMs) / processedCount
			atomic.StoreInt64(&p.metrics.AverageProcessingTimeMs, newAvg)
		} else {
			atomic.StoreInt64(&p.metrics.AverageProcessingTimeMs, processingTimeMs)
		}
	}

	return err
}
