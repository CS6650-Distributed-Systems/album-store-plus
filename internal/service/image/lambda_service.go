package image

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/storage"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// ProcessRequest represents the request payload for the image processing Lambda
type ProcessRequest struct {
	SourceKey      string `json:"sourceKey"`
	DestinationKey string `json:"destinationKey"`
}

// ProcessResponse represents the response from the image processing Lambda
type ProcessResponse struct {
	ProcessID string `json:"processId"`
	Status    string `json:"status"`
}

// LambdaService implements ImageService using AWS Lambda
type LambdaService struct {
	lambdaClient *lambda.Client
	storageRepo  storage.Repository
	functionName string
}

// NewLambdaService creates a new Lambda-based image service
func NewLambdaService(lambdaClient *lambda.Client, storageRepo storage.Repository, functionName string) *LambdaService {
	return &LambdaService{
		lambdaClient: lambdaClient,
		storageRepo:  storageRepo,
		functionName: functionName,
	}
}

// ProcessImage invokes the Lambda function to process an image
func (s *LambdaService) ProcessImage(ctx context.Context, originalKey, processedKey string) error {
	// Check if the original image exists
	exists, err := s.storageRepo.ObjectExists(ctx, originalKey)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("original image not found")
	}

	// Prepare the request payload
	request := ProcessRequest{
		SourceKey:      originalKey,
		DestinationKey: processedKey,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Invoke Lambda
	result, err := s.lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: &s.functionName,
		Payload:      payload,
	})

	if err != nil {
		return err
	}

	// Check Lambda execution status
	if result.FunctionError != nil {
		return errors.New("serverless function error: " + *result.FunctionError)
	}

	// Parse response
	var response ProcessResponse
	if err := json.Unmarshal(result.Payload, &response); err != nil {
		return err
	}

	// Check for 10 seconds if the processed image exists
	for i := 0; i < 10; i++ {
		exists, err := s.storageRepo.ObjectExists(ctx, processedKey)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return errors.New("timeout waiting for image processing")
}

// GetStatus checks the status of an image processing task
func (s *LambdaService) GetStatus(ctx context.Context, processID string) (string, error) {
	// In a real implementation, this would check a status table in DynamoDB
	// For simplicity, we'll just return "completed" or query the Lambda for status
	return "completed", nil
}
