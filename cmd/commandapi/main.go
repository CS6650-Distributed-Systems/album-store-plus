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

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/eventbus/sns"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/repository/rds"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/service/command"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/storage/s3"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/aws"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/config"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
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
	logger.Info("Starting command API",
		zap.String("environment", cfg.Environment),
		zap.String("version", "1.0.0"),
	)

	// Create AWS session
	awsConfig := &aws.SessionConfig{
		Region:   cfg.Database.Region,
		Endpoint: config.GetAWSEndpoint(),
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

	// Initialize database connection
	dbConfig := &rds.Config{
		Username:     cfg.Database.Username,
		Password:     cfg.Database.Password,
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		DatabaseName: cfg.Database.Name,
		MaxOpen:      cfg.Database.MaxOpen,
		MaxIdle:      cfg.Database.MaxIdle,
		MaxLifetime:  time.Duration(cfg.Database.Lifetime) * time.Second,
	}
	db, err := rds.Connect(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize schema if needed
	if err := rds.InitSchema(db); err != nil {
		logger.Fatal("Failed to initialize database schema", zap.Error(err))
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

	// Create bucket if needed
	if err := storageService.CreateBucketIfNotExists(context.Background()); err != nil {
		logger.Fatal("Failed to create S3 bucket", zap.Error(err))
	}

	// Initialize file manager
	fileManager := s3.NewFileManager(storageService)

	// Initialize event bus
	eventBusConfig := &sns.Config{
		Region:          cfg.SNS.Region,
		Endpoint:        cfg.SNS.Endpoint,
		TopicName:       "album-events",
		TopicARN:        cfg.SNS.TopicArn,
		CreateIfMissing: true,
	}
	eventBus, err := sns.NewEventBusFromConfig(context.Background(), awsSession.GetSession(), eventBusConfig)
	if err != nil {
		logger.Fatal("Failed to create event bus", zap.Error(err))
	}

	// Initialize repositories
	albumRepo := rds.NewAlbumRepository(db)

	// Initialize services
	validator := command.NewValidator()
	albumService := command.NewAlbumService(albumRepo, storageService, fileManager, eventBus, validator)

	// Create router
	router := chi.NewRouter()

	// Register routes
	routes := NewRoutes(albumService)
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
		logger.Info("Command API server started", zap.String("address", server.Addr))
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
