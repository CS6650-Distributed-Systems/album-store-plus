package s3

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// FileManager handles higher-level file operations including image processing
type FileManager struct {
	storageService *StorageService
}

// ImageInfo contains information about a processed image
type ImageInfo struct {
	ID          string
	Size        int64
	ContentType string
	Width       int
	Height      int
}

// NewFileManager creates a new file manager
func NewFileManager(storageService *StorageService) *FileManager {
	return &FileManager{
		storageService: storageService,
	}
}

// ProcessAndStoreImage processes and stores an image
func (fm *FileManager) ProcessAndStoreImage(ctx context.Context, albumID string, data []byte, contentType string) (*ImageInfo, error) {
	// Extract image information
	img, format, err := decodeImage(data, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Store original image
	imageID, err := fm.storageService.UploadImage(ctx, albumID, data, contentType)
	if err != nil {
		return nil, err
	}

	return &ImageInfo{
		ID:          imageID,
		Size:        int64(len(data)),
		ContentType: contentType,
		Width:       width,
		Height:      height,
	}, nil
}

// CreateThumbnail creates and stores a thumbnail of an image
func (fm *FileManager) CreateThumbnail(ctx context.Context, albumID string, imageID string, maxWidth, maxHeight int) (string, error) {
	// Get original image
	data, contentType, err := fm.storageService.GetImage(ctx, albumID, imageID)
	if err != nil {
		return "", err
	}

	// Decode image
	img, _, err := decodeImage(data, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Create thumbnail (not actually resizing here - would use an imaging library in production)
	// This is a simplified version for demonstration
	thumbData := data

	// Upload thumbnail
	thumbID := fmt.Sprintf("%s-thumb", imageID)
	thumbKey := fmt.Sprintf("images/%s/%s", albumID, thumbID)

	_, err = fm.storageService.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(fm.storageService.bucketName),
		Key:         aws.String(thumbKey),
		Body:        bytes.NewReader(thumbData),
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"AlbumId":       aws.String(albumID),
			"Thumbnail":     aws.String("true"),
			"OriginalImage": aws.String(imageID),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	return thumbID, nil
}

// BatchProcessImages processes multiple images in batch
func (fm *FileManager) BatchProcessImages(ctx context.Context, albumID string, images [][]byte, contentTypes []string) ([]ImageInfo, error) {
	if len(images) != len(contentTypes) {
		return nil, fmt.Errorf("number of images and content types must match")
	}

	var results []ImageInfo

	for i, data := range images {
		info, err := fm.ProcessAndStoreImage(ctx, albumID, data, contentTypes[i])
		if err != nil {
			return nil, fmt.Errorf("failed to process image %d: %w", i, err)
		}
		results = append(results, *info)
	}

	return results, nil
}

// decodeImage decodes image data based on content type
func decodeImage(data []byte, contentType string) (image.Image, string, error) {
	reader := bytes.NewReader(data)

	// Determine image format from content type
	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		img, err := jpeg.Decode(reader)
		return img, "jpeg", err
	} else if strings.Contains(contentType, "png") {
		img, err := png.Decode(reader)
		return img, "png", err
	} else {
		// Try to decode based on image data
		reader.Seek(0, 0) // Reset reader
		img, format, err := image.Decode(reader)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode image: %w", err)
		}
		return img, format, nil
	}
}
