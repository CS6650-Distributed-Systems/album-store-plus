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

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/dynamodb"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/service/query"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/storage/s3"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/aws"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/config"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	loggingCfg := &logging.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	}
	if err := logging.InitLogger(loggingCfg); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Sync()

	logger := logging.GetLogger()
	logger.Info("Starting query API",
		zap.String("environment", cfg.Environment),
		zap.String("version", "1.0.0"),
	)

	// Create AWS session
	awsConfig := &aws.SessionConfig{
		Region:   cfg.DynamoDB.Region,
		Endpoint: cfg.DynamoDB.Endpoint,
	}
	awsSession, err := aws.NewSession(awsConfig)
	if err != nil {
		logger.Fatal("Failed to create AWS session", zap.Error(err))
	}

	// Configure AWS credentials
	credsConfig := &aws.CredentialConfig{
		UseProfile: config.GetEnvBool("AWS_USE_PROFILE", false),
		Profile:    config.GetEnv("AWS_PROFILE", "default"),
	}
	aws.ConfigureSession(awsSession.GetSession(), credsConfig)

	// Create DynamoDB client
	dynamoClient := dynamodb.New(awsSession.GetSession())

	// Create DynamoDB tables if needed
	dynamoCfg := &dynamodb.Config{
		Region:             cfg.DynamoDB.Region,
		Endpoint:           cfg.DynamoDB.Endpoint,
		TableName:          cfg.DynamoDB.TableName,
		ReadCapacityUnits:  cfg.DynamoDB.ReadCapacityUnits,
		WriteCapacityUnits: cfg.DynamoDB.WriteCapacityUnits,
	}
	if err := dynamodb.CreateTablesIfNotExist(dynamoClient, dynamoCfg); err != nil {
		logger.Fatal("Failed to create DynamoDB tables", zap.Error(err))
	}

	// Initialize S3 storage service
	s3Config := &s3.Config{
		Region:     cfg.S3.Region,
		BucketName: cfg.S3.BucketName,
		Endpoint:   cfg.S3.Endpoint,
	}
	storageService, err := s3.NewStorageService(s3Config)
	if err != nil {
		logger.Fatal("Failed to create S3 storage service", zap.Error(err))
	}

	// Initialize repositories
	albumRepo := dynamodb.NewAlbumRepository(dynamoClient, cfg.DynamoDB.TableName)

	// Initialize services
	albumQueryService := query.NewAlbumQueryService(albumRepo, storageService)
	searchService := query.NewSearchService(albumRepo)

	// Create router
	router := chi.NewRouter()

	// Register routes
	routes := NewRoutes(albumQueryService, searchService)
	routes.RegisterRoutes(router)

	// Start HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Server shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.Info("Query API server started", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed", zap.Error(err))
	}

	logger.Info("Server stopped")
}
