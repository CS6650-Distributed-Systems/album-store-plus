package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/query"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/dynamodb"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/storage/s3"
	"go.uber.org/zap"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
)

// AlbumQueryService handles album query operations
type AlbumQueryService struct {
	albumRepo      *dynamodb.AlbumRepository
	storageService *s3.StorageService
	cacheTTL       time.Duration
}

// NewAlbumQueryService creates a new album query service
func NewAlbumQueryService(
	albumRepo *dynamodb.AlbumRepository,
	storageService *s3.StorageService,
) *AlbumQueryService {
	return &AlbumQueryService{
		albumRepo:      albumRepo,
		storageService: storageService,
		cacheTTL:       5 * time.Minute, // Default cache TTL
	}
}

// GetAlbum retrieves an album by ID
func (s *AlbumQueryService) GetAlbum(ctx context.Context, albumID string) (*query.AlbumView, error) {
	if albumID == "" {
		return nil, errors.New("album ID is required")
	}

	// Track view count asynchronously
	go func() {
		if err := s.albumRepo.UpdateViewCount(context.Background(), albumID); err != nil {
			logging.GetLogger().Error("Failed to update view count",
				zap.Error(err),
				zap.String("albumId", albumID),
			)
		}
	}()

	// Get album from DynamoDB
	album, err := s.albumRepo.GetAlbumView(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album: %w", err)
	}

	return album, nil
}

// GetAlbumStats retrieves statistics for an album
func (s *AlbumQueryService) GetAlbumStats(ctx context.Context, albumID string) (*query.AlbumStats, error) {
	if albumID == "" {
		return nil, errors.New("album ID is required")
	}

	// Get album statistics from DynamoDB
	stats, err := s.albumRepo.GetAlbumStats(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album statistics: %w", err)
	}

	return stats, nil
}

// SearchAlbums searches for albums based on query parameters
func (s *AlbumQueryService) SearchAlbums(ctx context.Context, searchQuery *query.AlbumSearchQuery) (*query.SearchResult, error) {
	if searchQuery == nil {
		searchQuery = &query.AlbumSearchQuery{
			Limit:  10,
			Offset: 0,
		}
	}

	// Validate and normalize search parameters
	if searchQuery.Limit <= 0 {
		searchQuery.Limit = 10
	} else if searchQuery.Limit > 100 {
		searchQuery.Limit = 100 // Cap the limit
	}

	if searchQuery.Offset < 0 {
		searchQuery.Offset = 0
	}

	// Execute search
	result, err := s.albumRepo.SearchAlbums(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search albums: %w", err)
	}

	return result, nil
}

// FacetedSearch performs a search with faceted results
func (s *AlbumQueryService) FacetedSearch(ctx context.Context, facetedSearch *query.FacetedSearch) (*query.FacetedSearchResult, error) {
	if facetedSearch == nil {
		return nil, errors.New("faceted search parameters are required")
	}

	// Execute faceted search
	result, err := s.albumRepo.FacetedSearch(ctx, facetedSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to perform faceted search: %w", err)
	}

	return result, nil
}

// GetImageURL generates a pre-signed URL for an album image
func (s *AlbumQueryService) GetImageURL(ctx context.Context, albumID, imageID string) (string, error) {
	if albumID == "" || imageID == "" {
		return "", errors.New("both album ID and image ID are required")
	}

	// Generate pre-signed URL with TTL
	url, err := s.storageService.GetImageURL(ctx, albumID, imageID, s.cacheTTL)
	if err != nil {
		return "", fmt.Errorf("failed to generate image URL: %w", err)
	}

	return url, nil
}

// SetCacheTTL sets the cache time-to-live for pre-signed URLs
func (s *AlbumQueryService) SetCacheTTL(ttl time.Duration) {
	if ttl > 0 {
		s.cacheTTL = ttl
	}
}
