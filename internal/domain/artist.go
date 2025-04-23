package domain

import (
	"context"
	"time"
)

// Artist represents a music artist
type Artist struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ArtistRepository defines the operations for artist storage
type ArtistRepository interface {
	// Create adds a new artist
	Create(ctx context.Context, artist *Artist) error

	// GetByID retrieves an artist by ID
	GetByID(ctx context.Context, id string) (*Artist, error)

	// Update modifies an existing artist
	Update(ctx context.Context, artist *Artist) error

	// Delete removes an artist
	Delete(ctx context.Context, id string) error
}
