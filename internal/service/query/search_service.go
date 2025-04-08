package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/query"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/dynamodb"
	"go.uber.org/zap"

	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
)

// SearchService provides advanced search capabilities
type SearchService struct {
	albumRepo *dynamodb.AlbumRepository
}

// NewSearchService creates a new search service
func NewSearchService(albumRepo *dynamodb.AlbumRepository) *SearchService {
	return &SearchService{
		albumRepo: albumRepo,
	}
}

// Search performs a search for albums
func (s *SearchService) Search(ctx context.Context, searchQuery *query.AlbumSearchQuery) (*query.SearchResult, error) {
	// Apply default values
	if searchQuery == nil {
		searchQuery = &query.AlbumSearchQuery{
			Limit:  10,
			Offset: 0,
		}
	}

	// Normalize query parameters
	s.normalizeSearchQuery(searchQuery)

	// Execute search
	result, err := s.albumRepo.SearchAlbums(ctx, searchQuery)
	if err != nil {
		logging.GetLogger().Error("Failed to search albums",
			zap.Error(err),
			zap.Any("query", searchQuery),
		)
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return result, nil
}

// FacetedSearch performs a search with facets
func (s *SearchService) FacetedSearch(ctx context.Context, query *query.AlbumSearchQuery, facetNames []string) (*query.FacetedSearchResult, error) {
	// Normalize query parameters
	s.normalizeSearchQuery(query)

	// Create facets
	facets := s.createFacets(facetNames)

	// Create faceted search
	facetedSearch := &query.FacetedSearch{
		Query:  *query,
		Facets: facets,
	}

	// Execute faceted search
	result, err := s.albumRepo.FacetedSearch(ctx, facetedSearch)
	if err != nil {
		logging.GetLogger().Error("Failed to perform faceted search",
			zap.Error(err),
			zap.Any("query", query),
			zap.Any("facets", facetNames),
		)
		return nil, fmt.Errorf("faceted search failed: %w", err)
	}

	return result, nil
}

// normalizeSearchQuery applies normalization to search query parameters
func (s *SearchService) normalizeSearchQuery(query *query.AlbumSearchQuery) {
	// Apply default values
	if query.Limit <= 0 {
		query.Limit = 10
	} else if query.Limit > 100 {
		query.Limit = 100 // Cap maximum
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	// Normalize sort parameters
	if query.SortBy == "" {
		query.SortBy = "popularity" // Default sort
	} else {
		query.SortBy = strings.ToLower(query.SortBy)
	}

	if query.SortOrder == "" {
		query.SortOrder = "desc" // Default order
	} else {
		query.SortOrder = strings.ToLower(query.SortOrder)
		if query.SortOrder != "asc" && query.SortOrder != "desc" {
			query.SortOrder = "desc" // Default if invalid
		}
	}

	// Trim search terms
	if query.Artist != "" {
		query.Artist = strings.TrimSpace(query.Artist)
	}
	if query.Title != "" {
		query.Title = strings.TrimSpace(query.Title)
	}
	if query.Year != "" {
		query.Year = strings.TrimSpace(query.Year)
	}
}

// createFacets creates facet definitions based on requested facet names
func (s *SearchService) createFacets(facetNames []string) []query.Facet {
	var facets []query.Facet

	for _, name := range facetNames {
		switch strings.ToLower(name) {
		case "year":
			facets = append(facets, query.Facet{
				Name: query.FacetYear,
				Type: "term",
			})
		case "artist":
			facets = append(facets, query.Facet{
				Name: query.FacetArtist,
				Type: "term",
			})
		case "likes":
			facets = append(facets, query.Facet{
				Name: query.FacetLikes,
				Type: "range",
				Values: []query.FacetValue{
					{Value: query.LikeRange{From: 0, To: 10}, Count: 0},
					{Value: query.LikeRange{From: 11, To: 50}, Count: 0},
					{Value: query.LikeRange{From: 51, To: 100}, Count: 0},
					{Value: query.LikeRange{From: 101, To: -1}, Count: 0}, // -1 means no upper bound
				},
			})
		}
	}

	return facets
}
