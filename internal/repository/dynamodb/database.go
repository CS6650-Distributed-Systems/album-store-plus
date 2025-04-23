package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Database represents a DynamoDB database connection
type Database struct {
	Client    *dynamodb.Client
	TableName string
}

// Connect creates a new DynamoDB client
func Connect(cfg aws.Config) *dynamodb.Client {
	return dynamodb.NewFromConfig(cfg)
}

// EnsureTableExists checks if the table exists and creates it if needed
func (db *Database) EnsureTableExists(ctx context.Context) error {
	// Check if table exists
	_, err := db.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(db.TableName),
	})

	// If no error, table exists
	if err == nil {
		return nil
	}

	// Create table for reviews with album_id as partition key
	_, err = db.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(db.TableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("album_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("album_id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest, // Use on-demand capacity mode
	})

	if err != nil {
		return fmt.Errorf("failed to create DynamoDB table: %w", err)
	}

	// Wait for table to be active
	waiter := dynamodb.NewTableExistsWaiter(db.Client)
	err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(db.TableName),
	}, 2*time.Minute)

	if err != nil {
		return fmt.Errorf("failed waiting for table creation: %w", err)
	}

	return nil
}
