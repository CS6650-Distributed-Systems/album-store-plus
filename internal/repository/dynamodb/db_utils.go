package dynamodb

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Config represents DynamoDB configuration
type Config struct {
	Region             string
	Endpoint           string // Use for local development with DynamoDB local
	TableName          string
	RankingsTableName  string
	ReadCapacityUnits  int64
	WriteCapacityUnits int64
}

// Connect creates a new DynamoDB client
func Connect(config *Config) (*dynamodb.DynamoDB, error) {
	awsConfig := &aws.Config{
		Region: aws.String(config.Region),
	}

	// Use endpoint for local development
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
	}

	// Create session
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create DynamoDB client
	return dynamodb.New(sess), nil
}

// CreateTablesIfNotExist creates required DynamoDB tables if they don't exist
func CreateTablesIfNotExist(client *dynamodb.DynamoDB, config *Config) error {
	// Create albums table
	if err := createAlbumsTable(client, config); err != nil {
		return err
	}

	// Create rankings table
	if err := createRankingsTable(client, config); err != nil {
		return err
	}

	return nil
}

// createAlbumsTable creates the albums table
func createAlbumsTable(client *dynamodb.DynamoDB, config *Config) error {
	_, err := client.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(config.TableName),
	})

	if err == nil {
		// Table already exists
		return nil
	}

	// Create table
	_, err = client.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(config.TableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("artist"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("year"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("ArtistIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("artist"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(config.ReadCapacityUnits),
					WriteCapacityUnits: aws.Int64(config.WriteCapacityUnits),
				},
			},
			{
				IndexName: aws.String("YearIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("year"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(config.ReadCapacityUnits),
					WriteCapacityUnits: aws.Int64(config.WriteCapacityUnits),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(config.ReadCapacityUnits),
			WriteCapacityUnits: aws.Int64(config.WriteCapacityUnits),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create albums table: %w", err)
	}

	// Wait for table to be active
	if err := waitForTableActive(client, config.TableName); err != nil {
		return err
	}

	return nil
}

// createRankingsTable creates the album rankings table
func createRankingsTable(client *dynamodb.DynamoDB, config *Config) error {
	rankingsTable := config.TableName + "-rankings"

	_, err := client.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(rankingsTable),
	})

	if err == nil {
		// Table already exists
		return nil
	}

	// Create table
	_, err = client.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(rankingsTable),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("albumId"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("rank"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("albumId"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("RankIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("rank"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(config.ReadCapacityUnits),
					WriteCapacityUnits: aws.Int64(config.WriteCapacityUnits),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(config.ReadCapacityUnits),
			WriteCapacityUnits: aws.Int64(config.WriteCapacityUnits),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create rankings table: %w", err)
	}

	// Wait for table to be active
	if err := waitForTableActive(client, rankingsTable); err != nil {
		return err
	}

	return nil
}

// waitForTableActive waits for a table to be in ACTIVE state
func waitForTableActive(client *dynamodb.DynamoDB, tableName string) error {
	maxRetries := 10
	delay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := client.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})

		if err != nil {
			return err
		}

		if *resp.Table.TableStatus == "ACTIVE" {
			return nil
		}

		time.Sleep(delay)
	}

	return fmt.Errorf("timed out waiting for table %s to become active", tableName)
}
