package rds

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/command"
	"github.com/google/uuid"
)

// AlbumRepository handles album operations on RDS
type AlbumRepository struct {
	db *sql.DB
}

// NewAlbumRepository creates a new AlbumRepository
func NewAlbumRepository(db *sql.DB) *AlbumRepository {
	return &AlbumRepository{
		db: db,
	}
}

// SaveAlbum saves an album to the database
func (r *AlbumRepository) SaveAlbum(ctx context.Context, album *command.Album) error {
	query := `
		INSERT INTO albums (id, artist, title, year, image_id, image_size, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		album.ID, album.Artist, album.Title, album.Year, album.ImageID, album.ImageSize,
		album.CreatedAt, album.UpdatedAt,
	)
	return err
}

// GetAlbum retrieves an album by ID
func (r *AlbumRepository) GetAlbum(ctx context.Context, id string) (*command.Album, error) {
	query := `
		SELECT id, artist, title, year, image_id, image_size, created_at, updated_at
		FROM albums WHERE id = ?
	`
	var album command.Album
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&album.ID, &album.Artist, &album.Title, &album.Year, &album.ImageID, &album.ImageSize,
		&album.CreatedAt, &album.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("album not found: %s", id)
		}
		return nil, err
	}
	return &album, nil
}

// SaveLike saves a like for an album
func (r *AlbumRepository) SaveLike(ctx context.Context, albumID string, liked bool) error {
	query := `
		INSERT INTO album_reviews (id, album_id, liked, created_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		uuid.New().String(), albumID, liked, time.Now(),
	)
	return err
}

// DeleteAlbum deletes an album by ID
func (r *AlbumRepository) DeleteAlbum(ctx context.Context, id string) error {
	query := "DELETE FROM albums WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// SaveEvent saves a domain event to the events table
func (r *AlbumRepository) SaveEvent(ctx context.Context, event *command.DomainEvent) error {
	// Implement event sourcing for CQRS
	return nil
}
