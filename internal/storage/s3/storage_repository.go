package s3

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Repository implements the storage Repository interface for S3
type Repository struct {
	client     *s3.Client
	bucketName string
	region     string
}

// NewRepository creates a new S3 storage repository
func NewRepository(client *s3.Client, bucketName, region string) *Repository {
	return &Repository{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}
}

// UploadObject uploads an object to S3
func (r *Repository) UploadObject(ctx context.Context, key string, data io.Reader, contentType string, contentLength int64) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(key),
		Body:          data,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	})
	return err
}

// DownloadObject downloads an object from S3
func (r *Repository) DownloadObject(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

// DeleteObject deletes an object from S3
func (r *Repository) DeleteObject(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// ObjectExists checks if an object exists in S3
func (r *Repository) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if the error is because the object doesn't exist
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetObjectURL returns a pre-signed URL for an object
func (r *Repository) GetObjectURL(ctx context.Context, key string) (string, error) {
	// Check if object exists
	exists, err := r.ObjectExists(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("object does not exist")
	}

	// For a real implementation, you'd use the presign client to create a URL
	// For now, we'll create a direct S3 URL
	return "https://" + r.bucketName + ".storage." + r.region + ".amazonaws.com/" + key, nil
}
