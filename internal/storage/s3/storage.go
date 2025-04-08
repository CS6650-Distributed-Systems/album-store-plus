package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// StorageService handles operations with S3 storage
type StorageService struct {
	client     *s3.S3
	bucketName string
	region     string
}

// Config represents S3 configuration
type Config struct {
	Region     string
	BucketName string
	Endpoint   string // For local development
}

// NewStorageService creates a new S3 storage service
func NewStorageService(config *Config) (*StorageService, error) {
	awsConfig := &aws.Config{
		Region: aws.String(config.Region),
	}

	// Use endpoint for local development
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	// Create session
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create S3 client
	return &StorageService{
		client:     s3.New(sess),
		bucketName: config.BucketName,
		region:     config.Region,
	}, nil
}

// CreateBucketIfNotExists creates the S3 bucket if it doesn't exist
func (s *StorageService) CreateBucketIfNotExists(ctx context.Context) error {
	// Check if bucket exists
	_, err := s.client.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})

	if err == nil {
		// Bucket already exists
		return nil
	}

	// Create bucket
	_, err = s.client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(s.region),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	return nil
}

// UploadImage uploads an image to S3
func (s *StorageService) UploadImage(ctx context.Context, albumID string, data []byte, contentType string) (string, error) {
	// Generate unique image ID
	imageID := uuid.New().String()
	key := path.Join("images", albumID, imageID)

	// Upload to S3
	_, err := s.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"AlbumId": aws.String(albumID),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload image to S3: %w", err)
	}

	return imageID, nil
}

// GetImage retrieves an image from S3
func (s *StorageService) GetImage(ctx context.Context, albumID, imageID string) ([]byte, string, error) {
	key := path.Join("images", albumID, imageID)

	// Get from S3
	result, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to get image from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the body
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return data, contentType, nil
}

// GetImageURL generates a pre-signed URL for an image
func (s *StorageService) GetImageURL(ctx context.Context, albumID, imageID string, expires time.Duration) (string, error) {
	key := path.Join("images", albumID, imageID)

	// Create request for pre-signed URL
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	// Generate pre-signed URL
	urlStr, err := req.Presign(expires)
	if err != nil {
		return "", fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	return urlStr, nil
}

// DeleteImage removes an image from S3
func (s *StorageService) DeleteImage(ctx context.Context, albumID, imageID string) error {
	key := path.Join("images", albumID, imageID)

	// Delete from S3
	_, err := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete image from S3: %w", err)
	}

	return nil
}
