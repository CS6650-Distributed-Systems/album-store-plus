package album

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/service"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/storage"
	"github.com/google/uuid"
)

// Service implements the AlbumService interface
type Service struct {
	albumRepo      domain.AlbumRepository
	artistRepo     domain.ArtistRepository
	reviewRepo     domain.ReviewRepository
	storageRepo    storage.Repository
	imageProcessor service.ImageService
}

// NewService creates a new album service
func NewService(
	albumRepo domain.AlbumRepository,
	artistRepo domain.ArtistRepository,
	reviewRepo domain.ReviewRepository,
	storageRepo storage.Repository,
	imageProcessor service.ImageService,
) *Service {
	return &Service{
		albumRepo:      albumRepo,
		artistRepo:     artistRepo,
		reviewRepo:     reviewRepo,
		storageRepo:    storageRepo,
		imageProcessor: imageProcessor,
	}
}

// CreateAlbum creates a new album
func (s *Service) CreateAlbum(ctx context.Context, album *domain.Album) error {
	// Validate artist exists
	artist, err := s.artistRepo.GetByID(ctx, album.ArtistID)
	if err != nil {
		return err
	}
	if artist == nil {
		return errors.New("artist not found")
	}

	return s.albumRepo.Create(ctx, album)
}

// GetAlbum retrieves an album by ID
func (s *Service) GetAlbum(ctx context.Context, id string, includeArtist, includeReview bool) (*domain.Album, error) {
	album, err := s.albumRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if album == nil {
		return nil, nil
	}

	// Include artist data if requested
	if includeArtist {
		artist, err := s.artistRepo.GetByID(ctx, album.ArtistID)
		if err != nil {
			return nil, err
		}
		album.Artist = artist
	}

	// Include review data if requested
	if includeReview {
		review, err := s.reviewRepo.GetByAlbumID(ctx, id)
		if err != nil {
			return nil, err
		}
		album.Review = review
	}

	return album, nil
}

// GetAlbumsByArtist retrieves all albums for a specific artist
func (s *Service) GetAlbumsByArtist(ctx context.Context, artistID string) ([]*domain.Album, error) {
	// Validate artist exists
	artist, err := s.artistRepo.GetByID(ctx, artistID)
	if err != nil {
		return nil, err
	}
	if artist == nil {
		return nil, errors.New("artist not found")
	}

	return s.albumRepo.GetByArtistID(ctx, artistID)
}

// UpdateAlbum modifies an existing album
func (s *Service) UpdateAlbum(ctx context.Context, album *domain.Album) error {
	// Verify album exists
	existingAlbum, err := s.albumRepo.GetByID(ctx, album.ID)
	if err != nil {
		return err
	}
	if existingAlbum == nil {
		return errors.New("album not found")
	}

	// Validate artist exists if artist ID is changing
	if album.ArtistID != existingAlbum.ArtistID {
		artist, err := s.artistRepo.GetByID(ctx, album.ArtistID)
		if err != nil {
			return err
		}
		if artist == nil {
			return errors.New("artist not found")
		}
	}

	// Preserve image keys if not explicitly changed
	if album.OriginalImageKey == "" {
		album.OriginalImageKey = existingAlbum.OriginalImageKey
	}
	if album.ProcessedImageKey == "" {
		album.ProcessedImageKey = existingAlbum.ProcessedImageKey
	}

	return s.albumRepo.Update(ctx, album)
}

// DeleteAlbum removes an album
func (s *Service) DeleteAlbum(ctx context.Context, id string) error {
	// Verify album exists
	album, err := s.albumRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if album == nil {
		return errors.New("album not found")
	}

	// Delete associated images if they exist
	if album.OriginalImageKey != "" {
		if err := s.storageRepo.DeleteObject(ctx, album.OriginalImageKey); err != nil {
			// Log error but continue
			// log.Printf("Error deleting original image: %v", err)
		}
	}
	if album.ProcessedImageKey != "" {
		if err := s.storageRepo.DeleteObject(ctx, album.ProcessedImageKey); err != nil {
			// Log error but continue
			// log.Printf("Error deleting processed image: %v", err)
		}
	}

	return s.albumRepo.Delete(ctx, id)
}

// UploadAlbumCover uploads and processes an album cover image
func (s *Service) UploadAlbumCover(ctx context.Context, albumID string, imageData io.Reader, filename string) error {
	// Verify album exists
	album, err := s.albumRepo.GetByID(ctx, albumID)
	if err != nil {
		return err
	}
	if album == nil {
		return errors.New("album not found")
	}

	// Generate a unique key for the original image
	ext := filepath.Ext(filename)
	originalKey := "albums/" + albumID + "/original/" + uuid.New().String() + ext

	// Upload the original image
	if err := s.storageRepo.UploadObject(ctx, originalKey, imageData, "image/"+ext[1:]); err != nil {
		return err
	}

	// Request image processing (this will be handled differently in Lambda vs. local implementations)
	processedKey := "albums/" + albumID + "/processed/" + uuid.New().String() + ".jpg"
	if err := s.imageProcessor.ProcessImage(ctx, originalKey, processedKey); err != nil {
		// If processing fails, still keep the original but log the error
		return err
	}

	// Update album with the new image keys
	return s.albumRepo.UpdateImageKeys(ctx, albumID, originalKey, processedKey)
}

// GetAlbumCoverURL gets the URL for an album's processed cover image
func (s *Service) GetAlbumCoverURL(ctx context.Context, albumID string) (string, error) {
	// Get the album to find the processed image key
	album, err := s.albumRepo.GetByID(ctx, albumID)
	if err != nil {
		return "", err
	}
	if album == nil {
		return "", errors.New("album not found")
	}

	if album.ProcessedImageKey == "" {
		return "", errors.New("album has no cover image")
	}

	// Get a URL for the processed image
	return s.storageRepo.GetObjectURL(ctx, album.ProcessedImageKey)
}
