package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/google/uuid"
)

// ArtistRepository implements the domain.ArtistRepository interface
type ArtistRepository struct {
	db *sql.DB
}

// NewArtistRepository creates a new MySQL artist repository
func NewArtistRepository(db *sql.DB) *ArtistRepository {
	return &ArtistRepository{
		db: db,
	}
}

// Create adds a new artist
func (r *ArtistRepository) Create(ctx context.Context, artist *domain.Artist) error {
	if artist.ID == "" {
		artist.ID = uuid.New().String()
	}

	now := time.Now()
	artist.CreatedAt = now
	artist.UpdatedAt = now

	query := `
		INSERT INTO artists (
			artist_id, name, created_at, updated_at
		)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		artist.ID,
		artist.Name,
		artist.CreatedAt,
		artist.UpdatedAt,
	)

	return err
}

// GetByID retrieves an artist by ID
func (r *ArtistRepository) GetByID(ctx context.Context, id string) (*domain.Artist, error) {
	query := `
		SELECT 
			artist_id, name, created_at, updated_at
		FROM artists
		WHERE artist_id = ?
	`

	var artist domain.Artist
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&artist.ID,
		&artist.Name,
		&artist.CreatedAt,
		&artist.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &artist, nil
}

// Update modifies an existing artist
func (r *ArtistRepository) Update(ctx context.Context, artist *domain.Artist) error {
	artist.UpdatedAt = time.Now()

	query := `
		UPDATE artists
		SET
			name = ?,
			updated_at = ?
		WHERE artist_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		artist.Name,
		artist.UpdatedAt,
		artist.ID,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("artist not found")
	}

	return nil
}

// Delete removes an artist
func (r *ArtistRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM artists WHERE artist_id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("artist not found")
	}

	return nil
}
