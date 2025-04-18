package review

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/mysql"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// MySQLService implements ReviewService using MySQL and SNS
type MySQLService struct {
	repo      *mysql.ReviewRepository
	snsClient *sns.Client
	topicARN  string
	metrics   struct {
		TotalRequests    int64
		TotalErrors      int64
		AverageLatencyMs int64
	}
}

// NewMySQLService creates a new MySQL review service
func NewMySQLService(repo *mysql.ReviewRepository, snsClient *sns.Client, topicARN string) *MySQLService {
	return &MySQLService{
		repo:      repo,
		snsClient: snsClient,
		topicARN:  topicARN,
	}
}

// GetReview retrieves a review by album ID
func (s *MySQLService) GetReview(ctx context.Context, albumID string) (*domain.Review, error) {
	return s.repo.GetByAlbumID(ctx, albumID)
}

// publishMessage publishes a review message to SNS
func (s *MySQLService) publishMessage(ctx context.Context, msgType ReviewMessageType, albumID string) error {
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
func (s *MySQLService) LikeAlbum(ctx context.Context, albumID string) error {
	return s.publishMessage(ctx, AddLikeType, albumID)
}

// DislikeAlbum queues a dislike operation for an album
func (s *MySQLService) DislikeAlbum(ctx context.Context, albumID string) error {
	return s.publishMessage(ctx, AddDislikeType, albumID)
}

// DeleteReview removes a review for an album
func (s *MySQLService) DeleteReview(ctx context.Context, albumID string) error {
	return s.publishMessage(ctx, DeleteType, albumID)
}
