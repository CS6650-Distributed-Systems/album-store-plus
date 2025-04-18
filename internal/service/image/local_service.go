package image

import (
	"context"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/storage"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"io"
)

// LocalService implements ImageService with local processing
type LocalService struct {
	storageRepo storage.Repository
	maxWidth    uint
	maxHeight   uint
	quality     int
}

// NewLocalService creates a new local image service
func NewLocalService(storageRepo storage.Repository, maxWidth, maxHeight uint, quality int) *LocalService {
	return &LocalService{
		storageRepo: storageRepo,
		maxWidth:    maxWidth,
		maxHeight:   maxHeight,
		quality:     quality,
	}
}

// ProcessImage processes an image using local Go libraries
func (s *LocalService) ProcessImage(ctx context.Context, originalKey, processedKey string) error {
	// Download the original image
	reader, err := s.storageRepo.DownloadObject(ctx, originalKey)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Decode the image
	img, _, err := image.Decode(reader)
	if err != nil {
		return err
	}

	// Resize the image
	resized := resize.Thumbnail(s.maxWidth, s.maxHeight, img, resize.Lanczos3)

	// Create a buffer to store the processed image
	pr, pw := io.Pipe()

	// Write the processed image in a goroutine
	go func() {
		defer pw.Close()
		err := jpeg.Encode(pw, resized, &jpeg.Options{Quality: s.quality})
		if err != nil {
			// We can't return the error directly from the goroutine
			// In a production app, you'd want to use a channel to communicate this error
			pw.CloseWithError(err)
		}
	}()

	// Upload the processed image
	err = s.storageRepo.UploadObject(ctx, processedKey, pr, "image/jpeg")
	if err != nil {
		return err
	}

	return nil
}

// GetStatus always returns "completed" for local processing since it's synchronous
func (s *LocalService) GetStatus(ctx context.Context, processID string) (string, error) {
	return "completed", nil
}
