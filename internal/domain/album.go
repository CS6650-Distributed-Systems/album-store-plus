package domain

import (
	"context"
	"time"
)

// Album represents a music album
type Album struct {
	ID                string    `json:"id"`
	ArtistID          string    `json:"artist_id"`
	Title             string    `json:"title"`
	Year              int       `json:"year"`
	OriginalImageKey  string    `json:"original_image_key,omitempty"`
	ProcessedImageKey string    `json:"processed_image_key,omitempty"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
	// Extra fields for JSON serialization
	Artist *Artist `json:"artist,omitempty"`
	Review *Review `json:"review,omitempty"`
}

// AlbumRepository defines the operations for album storage
type AlbumRepository interface {
	// Create adds a new album
	Create(ctx context.Context, album *Album) error

	// GetByID retrieves an album by ID
	GetByID(ctx context.Context, id string) (*Album, error)

	// GetByArtistID retrieves all albums for a specific artist
	GetByArtistID(ctx context.Context, artistID string) ([]*Album, error)

	// Update modifies an existing album
	Update(ctx context.Context, album *Album) error

	// Delete removes an album
	Delete(ctx context.Context, id string) error

	// UpdateImageKeys updates the image keys for an album
	UpdateImageKeys(ctx context.Context, id string, originalKey, processedKey string) error
}
