package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment string         `json:"environment"`
	Server      ServerConfig   `json:"server"`
	AWS         AWSConfig      `json:"aws"`
	MySQL       MySQLConfig    `json:"mysql"`
	DynamoDB    DynamoDBConfig `json:"dynamodb"`
	S3          S3Config       `json:"storage"`
	SNS         SNSConfig      `json:"sns"`
	SQS         SQSConfig      `json:"sqs"`
	Lambda      LambdaConfig   `json:"serverless"`
	Features    FeaturesConfig `json:"features"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port int `json:"port"`
}

// AWSConfig contains AWS general configuration
type AWSConfig struct {
	Region string `json:"region"`
}

// MySQLConfig contains MySQL database configuration
type MySQLConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// DynamoDBConfig contains DynamoDB configuration
type DynamoDBConfig struct {
	TableName string `json:"tableName"`
}

// S3Config contains S3 configuration
type S3Config struct {
	ImagesBucket string `json:"imagesBucket"`
}

// SNSConfig contains SNS configuration
type SNSConfig struct {
	TopicArn string `json:"topicArn"`
}

// SQSConfig contains SQS configuration
type SQSConfig struct {
	QueueUrl string `json:"queueUrl"`
}

// LambdaConfig contains Lambda configuration
type LambdaConfig struct {
	FunctionName string `json:"functionName"`
}

// FeaturesConfig contains feature flags for experiments
type FeaturesConfig struct {
	UseLocalImageProcessing bool `json:"useLocalImageProcessing"`
	UseDynamoDBForReviews   bool `json:"useDynamoDBForReviews"`
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool retrieves a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Try to parse as boolean
	lowercaseValue := strings.ToLower(value)
	if lowercaseValue == "true" || lowercaseValue == "1" || lowercaseValue == "yes" {
		return true
	}
	if lowercaseValue == "false" || lowercaseValue == "0" || lowercaseValue == "no" {
		return false
	}

	// If not parsable, return default
	return defaultValue
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// LoadConfig loads the application configuration
func LoadConfig() (*Config, error) {
	// Create configuration with environment variables or defaults
	cfg := &Config{
		Environment: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		AWS: AWSConfig{
			Region: getEnv("AWS_REGION", "us-west-2"),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			Username: getEnv("MYSQL_USERNAME", "album_store_user"),
			Password: getEnv("MYSQL_PASSWORD", "password"),
			Database: getEnv("MYSQL_DATABASE", "album_store"),
		},
		DynamoDB: DynamoDBConfig{
			TableName: getEnv("DYNAMODB_TABLE_NAME", "album_reviews"),
		},
		S3: S3Config{
			ImagesBucket: getEnv("S3_IMAGES_BUCKET", "album-store-covers"),
		},
		SNS: SNSConfig{
			TopicArn: getEnv("SNS_TOPIC_ARN", "your-topic-arn"),
		},
		SQS: SQSConfig{
			QueueUrl: getEnv("SQS_QUEUE_URL", "your-queue-url"),
		},
		Lambda: LambdaConfig{
			FunctionName: getEnv("LAMBDA_FUNCTION_NAME", "album-image-processor"),
		},
		Features: FeaturesConfig{
			UseLocalImageProcessing: getEnvBool("FEATURE_USE_LOCAL_IMAGE_PROCESSING", false),
			UseDynamoDBForReviews:   getEnvBool("FEATURE_USE_DYNAMODB_FOR_REVIEWS", true),
		},
	}

	return cfg, nil
}

// PrintConfig print the current configuration for development
func (c *Config) PrintConfig() {
	fmt.Println("=== Application Configuration ===")
	fmt.Println("Environment:", c.Environment)
	fmt.Println("Server Port:", c.Server.Port)
	fmt.Println("AWS Region:", c.AWS.Region)
	fmt.Println("MySQL:", c.MySQL.Host, c.MySQL.Port, c.MySQL.Database)
	fmt.Println("DynamoDB Table:", c.DynamoDB.TableName)
	fmt.Println("S3 Bucket:", c.S3.ImagesBucket)
	fmt.Println("SNS Topic ARN:", c.SNS.TopicArn)
	fmt.Println("SQS Queue URL:", c.SQS.QueueUrl)
	fmt.Println("Lambda Function:", c.Lambda.FunctionName)
	fmt.Println("Feature Flags:")
	fmt.Println("  - Use Local Image Processing:", c.Features.UseLocalImageProcessing)
	fmt.Println("  - Use DynamoDB for Reviews:", c.Features.UseDynamoDBForReviews)
	fmt.Println("===============================")
}
