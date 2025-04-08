package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/eventbus/sns"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/command"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/rds"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/storage/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
)

// AlbumService handles album command operations
type AlbumService struct {
	albumRepo      *rds.AlbumRepository
	storageService *s3.StorageService
	fileManager    *s3.FileManager
	eventBus       *sns.EventBus
	validator      *Validator
}

// NewAlbumService creates a new album service
func NewAlbumService(
	albumRepo *rds.AlbumRepository,
	storageService *s3.StorageService,
	fileManager *s3.FileManager,
	eventBus *sns.EventBus,
	validator *Validator,
) *AlbumService {
	return &AlbumService{
		albumRepo:      albumRepo,
		storageService: storageService,
		fileManager:    fileManager,
		eventBus:       eventBus,
		validator:      validator,
	}
}

// CreateAlbum creates a new album with an image
func (s *AlbumService) CreateAlbum(ctx context.Context, cmd *command.AlbumCreateCommand, imageData []byte, contentType string) (*command.Album, error) {
	// Validate command
	if err := s.validator.ValidateAlbumCreate(cmd); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Generate unique ID
	albumID := uuid.New().String()

	// Process and store image
	imageInfo, err := s.fileManager.ProcessAndStoreImage(ctx, albumID, imageData, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Create album
	album := command.NewAlbum(
		albumID,
		cmd.Artist,
		cmd.Title,
		cmd.Year,
		imageInfo.ID,
		imageInfo.Size,
	)

	// Save to database
	if err := s.albumRepo.SaveAlbum(ctx, album); err != nil {
		// Clean up the uploaded image if database operation fails
		cleanupErr := s.storageService.DeleteImage(ctx, albumID, imageInfo.ID)
		if cleanupErr != nil {
			logging.GetLogger().Error("Failed to clean up image after album creation failure",
				zap.Error(cleanupErr),
				zap.String("albumId", albumID),
				zap.String("imageId", imageInfo.ID),
			)
		}
		return nil, fmt.Errorf("failed to save album: %w", err)
	}

	// Publish event
	event := &command.AlbumCreatedEvent{
		AlbumID:   albumID,
		Artist:    cmd.Artist,
		Title:     cmd.Title,
		Year:      cmd.Year,
		ImageID:   imageInfo.ID,
		ImageSize: imageInfo.Size,
	}
	if err := s.eventBus.PublishAlbumCreated(ctx, event); err != nil {
		logging.GetLogger().Error("Failed to publish album created event",
			zap.Error(err),
			zap.String("albumId", albumID),
		)
		// Don't fail the operation if event publishing fails
	}

	// Also publish image uploaded event
	imageEvent := &command.ImageUploadedEvent{
		AlbumID:            albumID,
		ImageID:            imageInfo.ID,
		ImageSize:          imageInfo.Size,
		RequiresProcessing: true, // Assume all images need processing
	}
	if err := s.eventBus.PublishImageUploaded(ctx, imageEvent); err != nil {
		logging.GetLogger().Error("Failed to publish image uploaded event",
			zap.Error(err),
			zap.String("albumId", albumID),
			zap.String("imageId", imageInfo.ID),
		)
		// Don't fail the operation if event publishing fails
	}

	return album, nil
}

// LikeAlbum records a like for an album
func (s *AlbumService) LikeAlbum(ctx context.Context, albumID string) error {
	// Verify album exists
	album, err := s.albumRepo.GetAlbum(ctx, albumID)
	if err != nil {
		return fmt.Errorf("failed to get album: %w", err)
	}

	// Save like to database
	if err := s.albumRepo.SaveLike(ctx, albumID, true); err != nil {
		return fmt.Errorf("failed to save like: %w", err)
	}

	// Publish event
	event := &command.AlbumReviewEvent{
		AlbumID: albumID,
		Liked:   true,
	}
	if err := s.eventBus.PublishAlbumReview(ctx, event); err != nil {
		logging.GetLogger().Error("Failed to publish album like event",
			zap.Error(err),
			zap.String("albumId", albumID),
		)
		// Don't fail the operation if event publishing fails
	}

	return nil
}

// DislikeAlbum records a dislike for an album
func (s *AlbumService) DislikeAlbum(ctx context.Context, albumID string) error {
	// Verify album exists
	album, err := s.albumRepo.GetAlbum(ctx, albumID)
	if err != nil {
		return fmt.Errorf("failed to get album: %w", err)
	}

	// Save dislike to database
	if err := s.albumRepo.SaveLike(ctx, albumID, false); err != nil {
		return fmt.Errorf("failed to save dislike: %w", err)
	}

	// Publish event
	event := &command.AlbumReviewEvent{
		AlbumID: albumID,
		Liked:   false,
	}
	if err := s.eventBus.PublishAlbumReview(ctx, event); err != nil {
		logging.GetLogger().Error("Failed to publish album dislike event",
			zap.Error(err),
			zap.String("albumId", albumID),
		)
		// Don't fail the operation if event publishing fails
	}

	return nil
}

// BatchCreateAlbums creates multiple albums in batch
func (s *AlbumService) BatchCreateAlbums(ctx context.Context, commands []*command.AlbumCreateCommand,
	imageDataArray [][]byte, contentTypes []string) ([]*command.Album, error) {

	if len(commands) != len(imageDataArray) || len(commands) != len(contentTypes) {
		return nil, errors.New("number of commands, images, and content types must match")
	}

	var createdAlbums []*command.Album
	for i, cmd := range commands {
		album, err := s.CreateAlbum(ctx, cmd, imageDataArray[i], contentTypes[i])
		if err != nil {
			return nil, fmt.Errorf("failed to create album %d: %w", i, err)
		}
		createdAlbums = append(createdAlbums, album)
	}

	return createdAlbums, nil
}
