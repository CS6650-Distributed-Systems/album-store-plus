package dynamodb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/query"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// AlbumRepository handles album queries using DynamoDB
type AlbumRepository struct {
	client    *dynamodb.DynamoDB
	tableName string
}

// NewAlbumRepository creates a new AlbumRepository
func NewAlbumRepository(client *dynamodb.DynamoDB, tableName string) *AlbumRepository {
	return &AlbumRepository{
		client:    client,
		tableName: tableName,
	}
}

// GetAlbumView retrieves an album view by ID
func (r *AlbumRepository) GetAlbumView(ctx context.Context, id string) (*query.AlbumView, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	result, err := r.client.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get album from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("album not found: %s", id)
	}

	var album query.AlbumView
	if err := dynamodbattribute.UnmarshalMap(result.Item, &album); err != nil {
		return nil, fmt.Errorf("failed to unmarshal album: %w", err)
	}

	return &album, nil
}

// SearchAlbums searches for albums based on query parameters
func (r *AlbumRepository) SearchAlbums(ctx context.Context, searchQuery *query.AlbumSearchQuery) (*query.SearchResult, error) {
	// Build filter expression
	var builder expression.Builder
	var filter expression.ConditionBuilder
	var filterSet bool

	if searchQuery.Artist != "" {
		filter = expression.Name("artist").Contains(searchQuery.Artist)
		filterSet = true
	}

	if searchQuery.Title != "" {
		titleFilter := expression.Name("title").Contains(searchQuery.Title)
		if filterSet {
			filter = filter.And(titleFilter)
		} else {
			filter = titleFilter
			filterSet = true
		}
	}

	if searchQuery.Year != "" {
		yearFilter := expression.Name("year").Equal(expression.Value(searchQuery.Year))
		if filterSet {
			filter = filter.And(yearFilter)
		} else {
			filter = yearFilter
			filterSet = true
		}
	}

	if searchQuery.MinLikes != nil {
		likesFilter := expression.Name("likeCount").GreaterThanEqual(expression.Value(*searchQuery.MinLikes))
		if filterSet {
			filter = filter.And(likesFilter)
		} else {
			filter = likesFilter
			filterSet = true
		}
	}

	// Add filter if any conditions were set
	if filterSet {
		builder = builder.WithFilter(filter)
	}

	// Build projection (select all fields)
	builder = builder.WithProjection(expression.NamesList(
		expression.Name("id"),
		expression.Name("artist"),
		expression.Name("title"),
		expression.Name("year"),
		expression.Name("imageId"),
		expression.Name("imageSize"),
		expression.Name("likeCount"),
		expression.Name("dislikeCount"),
		expression.Name("popularity"),
		expression.Name("createdAt"),
		expression.Name("updatedAt"),
	))

	// Build expression
	expr, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	// Set limit and pagination
	limit := int64(10) // Default
	if searchQuery.Limit > 0 {
		limit = int64(searchQuery.Limit)
	}

	// Prepare scan input
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(r.tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		Limit:                     aws.Int64(limit),
	}

	// Set pagination if offset is provided
	if searchQuery.Offset > 0 {
		// Note: DynamoDB doesn't directly support offset
		// You would need to implement a more sophisticated pagination approach for production
	}

	// Execute scan
	result, err := r.client.ScanWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB: %w", err)
	}

	// Unmarshal items
	var albums []query.AlbumView
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &albums); err != nil {
		return nil, fmt.Errorf("failed to unmarshal albums: %w", err)
	}

	// Get total count (estimating from scan)
	totalCount := int(*result.Count)
	if result.ScannedCount != nil {
		totalCount = int(*result.ScannedCount)
	}

	return &query.SearchResult{
		Albums:     albums,
		TotalCount: totalCount,
		Limit:      int(limit),
		Offset:     searchQuery.Offset,
	}, nil
}

// GetAlbumStats retrieves statistics for an album
func (r *AlbumRepository) GetAlbumStats(ctx context.Context, albumID string) (*query.AlbumStats, error) {
	// Get album view first
	albumView, err := r.GetAlbumView(ctx, albumID)
	if err != nil {
		return nil, err
	}

	// Get album rank (this is a simplification - in production you might have a separate table for rankings)
	rankInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName + "-rankings"),
		KeyConditionExpression: aws.String("albumId = :albumId"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":albumId": {
				S: aws.String(albumID),
			},
		},
	}

	rankResult, err := r.client.QueryWithContext(ctx, rankInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get album rank: %w", err)
	}

	var rank int = 0
	if len(rankResult.Items) > 0 {
		rankAttr, ok := rankResult.Items[0]["rank"]
		if ok && rankAttr.N != nil {
			rankInt, _ := strconv.Atoi(*rankAttr.N)
			rank = rankInt
		}
	}

	// Calculate like ratio
	var likeRatio float64 = 0
	totalReviews := albumView.LikeCount + albumView.DislikeCount
	if totalReviews > 0 {
		likeRatio = float64(albumView.LikeCount) / float64(totalReviews)
	}

	// Create album stats
	return &query.AlbumStats{
		AlbumID:        albumID,
		ViewCount:      0, // This would be tracked separately
		LikeCount:      albumView.LikeCount,
		DislikeCount:   albumView.DislikeCount,
		LikeRatio:      likeRatio,
		PopularityRank: rank,
	}, nil
}

// FacetedSearch performs a search with faceted results
func (r *AlbumRepository) FacetedSearch(ctx context.Context, facetedSearch *query.FacetedSearch) (*query.FacetedSearchResult, error) {
	// This is a simplified implementation
	// In a real application, you would implement faceting logic here or use a search service

	// First, get the base search results
	searchResult, err := r.SearchAlbums(ctx, &facetedSearch.Query)
	if err != nil {
		return nil, err
	}

	// Simple facet calculation (this would be more sophisticated in production)
	var facets []query.Facet

	// Return the results with facets
	return &query.FacetedSearchResult{
		SearchResult: *searchResult,
		Facets:       facets,
	}, nil
}

// UpdateViewCount increments the view count for an album
func (r *AlbumRepository) UpdateViewCount(ctx context.Context, albumID string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(albumID),
			},
		},
		UpdateExpression: aws.String("ADD viewCount :inc"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":inc": {
				N: aws.String("1"),
			},
		},
	}

	_, err := r.client.UpdateItemWithContext(ctx, input)
	return err
}
