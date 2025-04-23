package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/handler"
	dynamoRepo "github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/dynamodb"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/mysql"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/service"
	albumSvc "github.com/CS6650-Distributed-Systems/album-store-plus/internal/service/album"
	imageSvc "github.com/CS6650-Distributed-Systems/album-store-plus/internal/service/image"
	reviewSvc "github.com/CS6650-Distributed-Systems/album-store-plus/internal/service/review"
	s3Storage "github.com/CS6650-Distributed-Systems/album-store-plus/internal/storage/s3"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/config"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func main() {
	// Create background context
	ctx := context.Background()

	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Print configuration for debugging purposes
	if cfg.Environment != "production" {
		cfg.PrintConfig()
	}

	// Initialize AWS SDK Configuration
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWS.Region),
	)
	if err != nil {
		log.Fatalf("Failed to initialize AWS config: %v", err)
	}

	// Initialize AWS clients
	s3Client := s3.NewFromConfig(awsCfg)
	dynamoDBClient := dynamodb.NewFromConfig(awsCfg)
	lambdaClient := lambda.NewFromConfig(awsCfg)
	snsClient := sns.NewFromConfig(awsCfg)

	// Initialize MySQL database connection
	mysqlDB, err := mysql.Connect(cfg.MySQL)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlDB.Close()

	// Initialize DynamoDB table
	dynamoDB := &dynamoRepo.Database{
		Client:    dynamoDBClient,
		TableName: cfg.DynamoDB.TableName,
	}
	if err := dynamoDB.EnsureTableExists(ctx); err != nil {
		log.Fatalf("Failed to ensure DynamoDB table exists: %v", err)
	}

	// Initialize repositories
	albumRepo := mysql.NewAlbumRepository(mysqlDB.DB)
	artistRepo := mysql.NewArtistRepository(mysqlDB.DB)
	mysqlReviewRepo := mysql.NewReviewRepository(mysqlDB.DB)
	dynamoDBReviewRepo := dynamoRepo.NewReviewRepository(dynamoDBClient, cfg.DynamoDB.TableName)

	// Initialize S3 storage repository
	storageRepo := s3Storage.NewRepository(s3Client, cfg.S3.ImagesBucket, awsCfg.Region)

	// Initialize services
	// Choose between local and Lambda image processing
	var imgProcessor service.ImageService
	if cfg.Features.UseLocalImageProcessing {
		imgProcessor = imageSvc.NewLocalService(storageRepo, 100, 100, 85)
	} else {
		imgProcessor = imageSvc.NewLambdaService(lambdaClient, storageRepo, cfg.Lambda.FunctionName)
	}

	// Choose between MySQL and DynamoDB for reviews
	var reviewService service.ReviewService
	if cfg.Features.UseDynamoDBForReviews {
		reviewService = reviewSvc.NewDynamoDBService(dynamoDBReviewRepo, snsClient, cfg.SNS.TopicArn)
	} else {
		reviewService = reviewSvc.NewMySQLService(mysqlReviewRepo, snsClient, cfg.SNS.TopicArn)
	}

	// Always use DynamoDB for reviews in the album service for better performance
	albumService := albumSvc.NewService(
		albumRepo,
		artistRepo,
		dynamoDBReviewRepo,
		storageRepo,
		imgProcessor,
	)

	// Set up API handlers
	albumHandler := handler.NewAlbumHandler(albumService)
	reviewHandler := handler.NewReviewHandler(reviewService)

	// Set up routes
	router := handler.SetupRoutes(albumHandler, reviewHandler)

	// Configure and start the HTTP server
	port := cfg.Server.Port
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
