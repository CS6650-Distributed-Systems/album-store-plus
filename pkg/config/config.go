package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the application configuration
type Config struct {
	Environment string         `json:"environment"`
	Server      ServerConfig   `json:"server"`
	Database    DatabaseConfig `json:"database"`
	DynamoDB    DynamoDBConfig `json:"dynamoDB"`
	S3          S3Config       `json:"s3"`
	SNS         SNSConfig      `json:"sns"`
	SQS         SQSConfig      `json:"sqs"`
	Lambda      LambdaConfig   `json:"lambda"`
	Logging     LoggingConfig  `json:"logging"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
}

// DatabaseConfig represents the RDS database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	MaxOpen  int    `json:"maxOpen"`
	MaxIdle  int    `json:"maxIdle"`
	Lifetime int    `json:"lifetime"`
}

// DynamoDBConfig represents the DynamoDB configuration
type DynamoDBConfig struct {
	Region             string `json:"region"`
	Endpoint           string `json:"endpoint"`
	TableName          string `json:"tableName"`
	ReadCapacityUnits  int64  `json:"readCapacityUnits"`
	WriteCapacityUnits int64  `json:"writeCapacityUnits"`
}

// S3Config represents the S3 configuration
type S3Config struct {
	Region     string `json:"region"`
	Endpoint   string `json:"endpoint"`
	BucketName string `json:"bucketName"`
}

// SNSConfig represents the SNS configuration
type SNSConfig struct {
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	TopicArn string `json:"topicArn"`
}

// SQSConfig represents the SQS configuration
type SQSConfig struct {
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	QueueUrl string `json:"queueUrl"`
}

// LambdaConfig represents the Lambda configuration
type LambdaConfig struct {
	Region       string `json:"region"`
	Endpoint     string `json:"endpoint"`
	FunctionName string `json:"functionName"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputPath string `json:"outputPath"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Apply environment overrides
	if err := applyEnvironmentOverrides(&config); err != nil {
		return nil, fmt.Errorf("error applying environment overrides: %w", err)
	}

	return &config, nil
}

// GetDefaultConfigPath returns the default configuration path
func GetDefaultConfigPath() string {
	// First, check the environment variable
	if path := os.Getenv("CONFIG_FILE"); path != "" {
		return path
	}

	// Next, check if configs/config.json exists in the current directory
	if _, err := os.Stat("configs/config.json"); err == nil {
		return "configs/config.json"
	}

	// Try parent directory
	if _, err := os.Stat("../configs/config.json"); err == nil {
		return "../configs/config.json"
	}

	// Default to "configs/config.json" if no config file is found
	return "configs/config.json"
}

// applyEnvironmentOverrides applies environment variable overrides to config
func applyEnvironmentOverrides(config *Config) error {
	// Override environment if set
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	// Database overrides
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		config.Database.Port = port
	}
	if user := os.Getenv("DB_USERNAME"); user != "" {
		config.Database.Username = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		config.Database.Password = pass
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		config.Database.Name = name
	}

	// AWS overrides
	if region := os.Getenv("AWS_REGION"); region != "" {
		config.DynamoDB.Region = region
		config.S3.Region = region
		config.SNS.Region = region
		config.SQS.Region = region
		config.Lambda.Region = region
	}

	// Endpoint overrides for local development
	if endpoint := os.Getenv("AWS_ENDPOINT"); endpoint != "" {
		config.DynamoDB.Endpoint = endpoint
		config.S3.Endpoint = endpoint
		config.SNS.Endpoint = endpoint
		config.SQS.Endpoint = endpoint
		config.Lambda.Endpoint = endpoint
	}

	return nil
}
