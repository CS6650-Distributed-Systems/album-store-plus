package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	dynamoRepo "github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/dynamodb"
	reviewSvc "github.com/CS6650-Distributed-Systems/album-store-plus/internal/service/review"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/config"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	// Create background context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize AWS SDK Configuration
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWS.Region),
	)
	if err != nil {
		log.Fatalf("Failed to initialize AWS config: %v", err)
	}

	// Initialize AWS clients
	dynamoDBClient := dynamodb.NewFromConfig(awsCfg)
	sqsClient := sqs.NewFromConfig(awsCfg)

	// Initialize DynamoDB table
	dynamoDB := &dynamoRepo.Database{
		Client:    dynamoDBClient,
		TableName: cfg.DynamoDB.TableName,
	}
	if err := dynamoDB.EnsureTableExists(ctx); err != nil {
		log.Fatalf("Failed to ensure DynamoDB table exists: %v", err)
	}

	// Initialize repositories
	dynamoDBReviewRepo := dynamoRepo.NewReviewRepository(dynamoDBClient, cfg.DynamoDB.TableName)

	// Initialize message processor
	processor := reviewSvc.NewMessageProcessor(
		dynamoDBReviewRepo,
		sqsClient,
		cfg.SQS.QueueUrl,
	)

	// Start the processor
	log.Println("Starting review message processor...")
	processor.Start(ctx)
	log.Println("Review message processor started")

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	processor.Stop()
	log.Println("Worker shutdown complete")
}
