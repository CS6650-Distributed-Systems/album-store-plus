package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/google/uuid"
)

// ReviewRepository implements the domain.ReviewRepository interface with MySQL
type ReviewRepository struct {
	db *sql.DB
}

// NewReviewRepository creates a new MySQL review repository
func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{
		db: db,
	}
}

// GetByAlbumID retrieves a review by album ID
func (r *ReviewRepository) GetByAlbumID(ctx context.Context, albumID string) (*domain.Review, error) {
	query := `
		SELECT 
			review_id, album_id, like_count, dislike_count,
			created_at, updated_at
		FROM reviews
		WHERE album_id = ?
	`

	var review domain.Review
	err := r.db.QueryRowContext(ctx, query, albumID).Scan(
		&review.ID,
		&review.AlbumID,
		&review.LikeCount,
		&review.DislikeCount,
		&review.CreatedAt,
		&review.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &review, nil
}

// addCount increments a specific counter (like or dislike) for an album
func (r *ReviewRepository) addCount(ctx context.Context, albumID string, columnName string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if review exists
	checkQuery := "SELECT review_id FROM reviews WHERE album_id = ? FOR UPDATE"
	var reviewID string
	err = tx.QueryRowContext(ctx, checkQuery, albumID).Scan(&reviewID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create a new review with default values
			reviewID = uuid.New().String()
			now := time.Now()

			// Initialize counters
			likeCount := 0
			dislikeCount := 0

			// Set the appropriate counter to 1 based on the current operation
			if columnName == "like_count" {
				likeCount = 1
			} else if columnName == "dislike_count" {
				dislikeCount = 1
			}

			insertQuery := `
                INSERT INTO reviews (
                    review_id, album_id, like_count, dislike_count, 
                    created_at, updated_at
                )
                VALUES (?, ?, ?, ?, ?, ?)
            `

			_, err = tx.ExecContext(
				ctx,
				insertQuery,
				reviewID,
				albumID,
				likeCount,
				dislikeCount,
				now,
				now,
			)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Increment only the specified counter for existing review
		updateQuery := fmt.Sprintf(`
            UPDATE reviews
            SET %s = %s + 1, updated_at = ?
            WHERE review_id = ?
        `, columnName, columnName)

		_, err = tx.ExecContext(ctx, updateQuery, time.Now(), reviewID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddLike atomically increments the like count for an album
func (r *ReviewRepository) AddLike(ctx context.Context, albumID string) error {
	return r.addCount(ctx, albumID, "like_count")
}

// AddDislike atomically increments the dislike count for an album
func (r *ReviewRepository) AddDislike(ctx context.Context, albumID string) error {
	return r.addCount(ctx, albumID, "dislike_count")
}

// Delete removes a review by album ID
func (r *ReviewRepository) Delete(ctx context.Context, albumID string) error {
	query := `DELETE FROM reviews WHERE album_id = ?`

	result, err := r.db.ExecContext(ctx, query, albumID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("review not found")
	}

	return nil
}
