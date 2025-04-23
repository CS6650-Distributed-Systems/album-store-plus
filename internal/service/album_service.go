package service

import (
	"context"
	"io"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
)

// AlbumService defines operations available for albums
type AlbumService interface {
	// CreateAlbum creates a new album
	CreateAlbum(ctx context.Context, album *domain.Album) error

	// GetAlbum retrieves an album by ID (with optional related data)
	GetAlbum(ctx context.Context, id string, includeArtist, includeReview bool) (*domain.Album, error)

	// GetAlbumsByArtist retrieves all albums for a specific artist
	GetAlbumsByArtist(ctx context.Context, artistID string) ([]*domain.Album, error)

	// UpdateAlbum modifies an existing album
	UpdateAlbum(ctx context.Context, album *domain.Album) error

	// DeleteAlbum removes an album
	DeleteAlbum(ctx context.Context, id string) error

	// UploadAlbumCover uploads and processes an album cover image
	UploadAlbumCover(ctx context.Context, albumID string, imageData io.Reader, filename string) error

	// GetAlbumCoverURL gets the URL for an album's processed cover image
	GetAlbumCoverURL(ctx context.Context, albumID string) (string, error)
}
