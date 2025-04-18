package dynamodb

import (
	"context"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ReviewRepository implements the domain.ReviewRepository interface with DynamoDB
type ReviewRepository struct {
	client    *dynamodb.Client
	tableName string
}

// DynamoDBReview is the internal representation for DynamoDB
// Note: We don't use ID field in DynamoDB, as album_id is the primary key
type DynamoDBReview struct {
	AlbumID      string `dynamodbav:"album_id"`
	LikeCount    int    `dynamodbav:"like_count"`
	DislikeCount int    `dynamodbav:"dislike_count"`
}

// NewReviewRepository creates a new DynamoDB review repository
func NewReviewRepository(client *dynamodb.Client, tableName string) *ReviewRepository {
	return &ReviewRepository{
		client:    client,
		tableName: tableName,
	}
}

// GetByAlbumID retrieves a review by album ID
func (r *ReviewRepository) GetByAlbumID(ctx context.Context, albumID string) (*domain.Review, error) {
	// Get item by album_id
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"album_id": &types.AttributeValueMemberS{Value: albumID},
		},
	})
	if err != nil {
		return nil, err
	}

	// Check if item exists
	if result.Item == nil || len(result.Item) == 0 {
		return nil, nil
	}

	// Convert DynamoDB item to review
	var dbReview DynamoDBReview
	err = attributevalue.UnmarshalMap(result.Item, &dbReview)
	if err != nil {
		return nil, err
	}

	// Convert to domain model - note we don't set ID since it's not used in DynamoDB
	review := &domain.Review{
		AlbumID:      dbReview.AlbumID,
		LikeCount:    uint(dbReview.LikeCount),
		DislikeCount: uint(dbReview.DislikeCount),
	}

	return review, nil
}

// createReview adds a new review record (private helper method)
// Note: For DynamoDB, we ignore the ID field and use AlbumID as the primary key
func (r *ReviewRepository) createReview(ctx context.Context, review *domain.Review) error {
	// Convert domain model to DynamoDB representation
	dbReview := DynamoDBReview{
		AlbumID:      review.AlbumID,
		LikeCount:    int(review.LikeCount),
		DislikeCount: int(review.DislikeCount),
	}

	// Convert to DynamoDB attribute values
	item, err := attributevalue.MarshalMap(dbReview)
	if err != nil {
		return err
	}

	// Create/update item in DynamoDB
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})

	return err
}

// addCount atomically increments a specific counter (like or dislike) for an album
func (r *ReviewRepository) addCount(ctx context.Context, albumID string, counterName string) error {
	// Check if review exists
	review, err := r.GetByAlbumID(ctx, albumID)
	if err != nil {
		return err
	}

	// Create a new review if it doesn't exist
	if review == nil {
		newReview := &domain.Review{
			AlbumID: albumID,
		}

		// Initialize the appropriate counter
		if counterName == "like_count" {
			newReview.LikeCount = 1
		} else if counterName == "dislike_count" {
			newReview.DislikeCount = 1
		}

		return r.createReview(ctx, newReview)
	}

	// For existing reviews, use atomic counter to increment the specific count
	// This is where DynamoDB shines compared to MySQL
	updateExpression := "SET " + counterName + " = if_not_exists(" + counterName + ", :zero) + :incr"

	updateParams := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"album_id": &types.AttributeValueMemberS{Value: albumID},
		},
		UpdateExpression: aws.String(updateExpression),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":incr": &types.AttributeValueMemberN{Value: "1"},
			":zero": &types.AttributeValueMemberN{Value: "0"},
		},
	}

	_, err = r.client.UpdateItem(ctx, updateParams)
	return err
}

// AddLike adds a like to an album
func (r *ReviewRepository) AddLike(ctx context.Context, albumID string) error {
	return r.addCount(ctx, albumID, "like_count")
}

// AddDislike adds a dislike to an album
func (r *ReviewRepository) AddDislike(ctx context.Context, albumID string) error {
	return r.addCount(ctx, albumID, "dislike_count")
}

// Delete removes a review by album ID
func (r *ReviewRepository) Delete(ctx context.Context, albumID string) error {
	// Define delete parameters
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"album_id": &types.AttributeValueMemberS{Value: albumID},
		},
		ConditionExpression: aws.String("attribute_exists(album_id)"),
	}

	// Execute delete
	_, err := r.client.DeleteItem(ctx, params)
	if err != nil {
		return err
	}

	return nil
}
