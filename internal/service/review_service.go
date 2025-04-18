package service

import (
	"context"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
)

// ReviewService defines the operations available for album reviews
type ReviewService interface {
	// GetReview retrieves a review by album ID
	GetReview(ctx context.Context, albumID string) (*domain.Review, error)

	// LikeAlbum increments the like count for an album
	LikeAlbum(ctx context.Context, albumID string) error

	// DislikeAlbum increments the dislike count for an album
	DislikeAlbum(ctx context.Context, albumID string) error

	// DeleteReview removes a review by album ID
	DeleteReview(ctx context.Context, albumID string) error
}
