package storage

import (
	"context"
	"io"
)

// Repository defines operations for object storage
type Repository interface {
	// UploadObject uploads an object to storage
	UploadObject(ctx context.Context, key string, data io.Reader, contentType string, contentLength int64) error

	// DownloadObject downloads an object from storage
	DownloadObject(ctx context.Context, key string) (io.ReadCloser, error)

	// DeleteObject deletes an object from storage
	DeleteObject(ctx context.Context, key string) error

	// ObjectExists checks if an object exists in storage
	ObjectExists(ctx context.Context, key string) (bool, error)

	// GetObjectURL returns a URL for an object
	GetObjectURL(ctx context.Context, key string) (string, error)
}
