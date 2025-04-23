package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	dynamoRepo "github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/dynamodb"
	mysqlRepo "github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/mysql"
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

	// Variable to hold the repository implementation
	var reviewRepo domain.ReviewRepository

	// Choose between MySQL and DynamoDB based on feature flag
	if cfg.Features.UseDynamoDBForReviews {
		// Initialize DynamoDB table
		dynamoDB := &dynamoRepo.Database{
			Client:    dynamoDBClient,
			TableName: cfg.DynamoDB.TableName,
		}
		if err := dynamoDB.EnsureTableExists(ctx); err != nil {
			log.Fatalf("Failed to ensure DynamoDB table exists: %v", err)
		}

		// Initialize DynamoDB repository
		reviewRepo = dynamoRepo.NewReviewRepository(dynamoDBClient, cfg.DynamoDB.TableName)
		log.Println("Using DynamoDB for review storage")
	} else {
		// Initialize MySQL database connection
		mysqlDB, err := mysqlRepo.Connect(cfg.MySQL)
		if err != nil {
			log.Fatalf("Failed to connect to MySQL: %v", err)
		}
		defer mysqlDB.Close()

		// Initialize MySQL repository
		reviewRepo = mysqlRepo.NewReviewRepository(mysqlDB.DB)
		log.Println("Using MySQL for review storage")
	}

	// Initialize message processor with the chosen repository
	processor := reviewSvc.NewMessageProcessor(
		reviewRepo,
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
