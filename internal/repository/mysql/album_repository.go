package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/google/uuid"
)

// AlbumRepository implements the domain.AlbumRepository interface
type AlbumRepository struct {
	db *sql.DB
}

// NewAlbumRepository creates a new MySQL album repository
func NewAlbumRepository(db *sql.DB) *AlbumRepository {
	return &AlbumRepository{
		db: db,
	}
}

// Create adds a new album
func (r *AlbumRepository) Create(ctx context.Context, album *domain.Album) error {
	if album.ID == "" {
		album.ID = uuid.New().String()
	}

	now := time.Now()
	album.CreatedAt = now
	album.UpdatedAt = now

	query := `
		INSERT INTO albums (
			album_id, artist_id, title, year, 
			original_image_key, processed_image_key,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		album.ID,
		album.ArtistID,
		album.Title,
		album.Year,
		album.OriginalImageKey,
		album.ProcessedImageKey,
		album.CreatedAt,
		album.UpdatedAt,
	)

	return err
}

// GetByID retrieves an album by ID
func (r *AlbumRepository) GetByID(ctx context.Context, id string) (*domain.Album, error) {
	query := `
		SELECT 
			album_id, artist_id, title, year,
			original_image_key, processed_image_key,
			created_at, updated_at
		FROM albums
		WHERE album_id = ?
	`

	var album domain.Album
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&album.ID,
		&album.ArtistID,
		&album.Title,
		&album.Year,
		&album.OriginalImageKey,
		&album.ProcessedImageKey,
		&album.CreatedAt,
		&album.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &album, nil
}

// GetByArtistID retrieves all albums for a specific artist
func (r *AlbumRepository) GetByArtistID(ctx context.Context, artistID string) ([]*domain.Album, error) {
	query := `
		SELECT 
			album_id, artist_id, title, year,
			original_image_key, processed_image_key,
			created_at, updated_at
		FROM albums
		WHERE artist_id = ?
		ORDER BY year DESC, title ASC
	`

	rows, err := r.db.QueryContext(ctx, query, artistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []*domain.Album
	for rows.Next() {
		var album domain.Album
		err := rows.Scan(
			&album.ID,
			&album.ArtistID,
			&album.Title,
			&album.Year,
			&album.OriginalImageKey,
			&album.ProcessedImageKey,
			&album.CreatedAt,
			&album.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		albums = append(albums, &album)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return albums, nil
}

// Update modifies an existing album
func (r *AlbumRepository) Update(ctx context.Context, album *domain.Album) error {
	album.UpdatedAt = time.Now()

	query := `
		UPDATE albums
		SET
			artist_id = ?,
			title = ?,
			year = ?,
			original_image_key = ?,
			processed_image_key = ?,
			updated_at = ?
		WHERE album_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		album.ArtistID,
		album.Title,
		album.Year,
		album.OriginalImageKey,
		album.ProcessedImageKey,
		album.UpdatedAt,
		album.ID,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("album not found")
	}

	return nil
}

// Delete removes an album
func (r *AlbumRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM albums WHERE album_id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("album not found")
	}

	return nil
}

// UpdateImageKeys updates the image keys for an album
func (r *AlbumRepository) UpdateImageKeys(ctx context.Context, id string, originalKey, processedKey string) error {
	now := time.Now()

	query := `
		UPDATE albums
		SET 
			original_image_key = ?,
			processed_image_key = ?,
			updated_at = ?
		WHERE album_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		originalKey,
		processedKey,
		now,
		id,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("album not found")
	}

	return nil
}
