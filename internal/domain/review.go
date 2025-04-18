package domain

import (
	"context"
	"time"
)

// Review represents an album review with like/dislike counts
type Review struct {
	ID           string    `json:"id,omitempty"` // Used by MySQL, optional for DynamoDB
	AlbumID      string    `json:"album_id"`     // Primary identifier in DynamoDB
	LikeCount    uint      `json:"like_count"`
	DislikeCount uint      `json:"dislike_count"`
	CreatedAt    time.Time `json:"created_at,omitempty"` // Used by MySQL, optional for DynamoDB
	UpdatedAt    time.Time `json:"updated_at,omitempty"` // Used by MySQL, optional for DynamoDB
}

// ReviewRepository interface defines operations for review storage
type ReviewRepository interface {
	// GetByAlbumID retrieves a review by album ID
	GetByAlbumID(ctx context.Context, albumID string) (*Review, error)

	// AddLike atomically increments the like count for an album
	AddLike(ctx context.Context, albumID string) error

	// AddDislike atomically increments the dislike count for an album
	AddDislike(ctx context.Context, albumID string) error

	// Delete removes a review
	Delete(ctx context.Context, albumID string) error
}
