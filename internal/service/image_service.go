package service

import (
	"context"
)

// ImageService defines operations for processing images
type ImageService interface {
	// ProcessImage takes an original image key and processes it to a destination key
	ProcessImage(ctx context.Context, originalKey, processedKey string) error

	// GetStatus checks the processing status of an image
	GetStatus(ctx context.Context, processID string) (string, error)
}
