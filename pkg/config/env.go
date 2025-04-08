package config

import (
	"os"
	"strconv"
	"time"
)

// Environment constants
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTesting     = "testing"
)

// IsProduction returns true if the current environment is production
func IsProduction(config *Config) bool {
	return config.Environment == EnvProduction
}

// IsDevelopment returns true if the current environment is development
func IsDevelopment(config *Config) bool {
	return config.Environment == EnvDevelopment
}

// IsTesting returns true if the current environment is testing
func IsTesting(config *Config) bool {
	return config.Environment == EnvTesting
}

// GetEnv retrieves an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt retrieves an environment variable as int or returns a default value
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool retrieves an environment variable as bool or returns a default value
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvDuration retrieves an environment variable as time.Duration or returns a default value
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
	}
	return defaultValue
}

// GetAWSRegion retrieves the AWS region from environment or returns a default
func GetAWSRegion(defaultRegion string) string {
	return GetEnv("AWS_REGION", defaultRegion)
}

// GetAWSEndpoint retrieves the AWS endpoint from environment (for local development)
func GetAWSEndpoint() string {
	return GetEnv("AWS_ENDPOINT", "")
}

// UseLocalstack returns true if we should use LocalStack (local AWS emulator)
func UseLocalstack() bool {
	return GetEnvBool("USE_LOCALSTACK", false)
}
