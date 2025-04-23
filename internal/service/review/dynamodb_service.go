package review

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// DynamoDBService implements ReviewService using DynamoDB and SNS
type DynamoDBService struct {
	repo      *dynamodb.ReviewRepository
	snsClient *sns.Client
	topicARN  string
	metrics   struct {
		TotalRequests    int64
		TotalErrors      int64
		AverageLatencyMs int64
	}
}

// NewDynamoDBService creates a new DynamoDB review service
func NewDynamoDBService(repo *dynamodb.ReviewRepository, snsClient *sns.Client, topicARN string) *DynamoDBService {
	return &DynamoDBService{
		repo:      repo,
		snsClient: snsClient,
		topicARN:  topicARN,
	}
}

// GetReview retrieves a review by album ID
func (s *DynamoDBService) GetReview(ctx context.Context, albumID string) (*domain.Review, error) {
	return s.repo.GetByAlbumID(ctx, albumID)
}

// publishMessage publishes a review message to SNS
func (s *DynamoDBService) publishMessage(ctx context.Context, msgType ReviewMessageType, albumID string) error {
	startTime := time.Now()

	atomic.AddInt64(&s.metrics.TotalRequests, 1)

	msg := ReviewMessage{
		Type:    msgType,
		AlbumID: albumID,
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		atomic.AddInt64(&s.metrics.TotalErrors, 1)
		return err
	}

	// Publish to SNS
	msgString := string(msgJSON)
	_, err = s.snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: &s.topicARN,
		Message:  &msgString,
	})

	if err != nil {
		atomic.AddInt64(&s.metrics.TotalErrors, 1)
		return err
	}

	// Update average latency
	latencyMs := time.Since(startTime).Milliseconds()
	currentAvg := atomic.LoadInt64(&s.metrics.AverageLatencyMs)
	currentTotal := atomic.LoadInt64(&s.metrics.TotalRequests)

	// Calculate new average latency
	newAvg := (currentAvg*(currentTotal-1) + latencyMs) / currentTotal
	atomic.StoreInt64(&s.metrics.AverageLatencyMs, newAvg)

	return nil
}

// LikeAlbum queues a like operation for an album
func (s *DynamoDBService) LikeAlbum(ctx context.Context, albumID string) error {
	return s.publishMessage(ctx, AddLikeType, albumID)
}

// DislikeAlbum queues a dislike operation for an album
func (s *DynamoDBService) DislikeAlbum(ctx context.Context, albumID string) error {
	return s.publishMessage(ctx, AddDislikeType, albumID)
}

// DeleteReview removes a review for an album
func (s *DynamoDBService) DeleteReview(ctx context.Context, albumID string) error {
	return s.publishMessage(ctx, DeleteType, albumID)
}
