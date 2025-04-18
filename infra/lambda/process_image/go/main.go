package _go

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"io"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

// Configuration values matching the local_service.go implementation
const (
	maxWidth  = 100
	maxHeight = 100
	quality   = 85
)

// ProcessRequest matches the request format defined in lambda_service.go
type ProcessRequest struct {
	SourceKey      string `json:"sourceKey"`
	DestinationKey string `json:"destinationKey"`
}

// ProcessResponse matches the response format defined in lambda_service.go
type ProcessResponse struct {
	ProcessID string `json:"processId"`
	Status    string `json:"status"`
}

// S3Client is a simple interface for S3 operations needed by the Lambda
type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// Handler processes the Lambda request
func Handler(ctx context.Context, request ProcessRequest) (ProcessResponse, error) {
	// Generate a unique process ID
	processID := uuid.New().String()

	// Create an AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ProcessResponse{
			ProcessID: processID,
			Status:    "error",
		}, fmt.Errorf("error loading AWS configuration: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Process the image
	err = processImage(ctx, s3Client, request.SourceKey, request.DestinationKey)
	if err != nil {
		return ProcessResponse{
			ProcessID: processID,
			Status:    "error",
		}, err
	}

	// Return success response
	return ProcessResponse{
		ProcessID: processID,
		Status:    "completed",
	}, nil
}

// processImage downloads an image from S3, resizes it, and uploads the processed version back to S3
func processImage(ctx context.Context, s3Client S3Client, sourceKey, destinationKey string) error {
	// Extract bucket name from environment variable (set in Lambda configuration)
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		return errors.New("S3_BUCKET_NAME environment variable not set")
	}

	// Download the original image
	getResult, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(sourceKey),
	})
	if err != nil {
		return fmt.Errorf("error downloading original image: %w", err)
	}
	defer getResult.Body.Close()

	// Decode the image
	img, _, err := image.Decode(getResult.Body)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}

	// Resize the image using the same parameters as local_service.go
	resized := resize.Thumbnail(maxWidth, maxHeight, img, resize.Lanczos3)

	// Create a pipe for streaming the processed image
	pr, pw := io.Pipe()

	// Start a goroutine to encode and write the image to the pipe
	go func() {
		defer pw.Close()
		err := jpeg.Encode(pw, resized, &jpeg.Options{Quality: quality})
		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	// Upload the processed image
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(destinationKey),
		Body:        pr,
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("error uploading processed image: %w", err)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
