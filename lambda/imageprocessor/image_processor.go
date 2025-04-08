package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
)

// ImageProcessor processes images stored in S3
type ImageProcessor struct {
	s3Client       *s3.S3
	bucketName     string
	maxWidth       uint
	maxHeight      uint
	thumbnailWidth uint
}

// NewImageProcessor creates a new image processor
func NewImageProcessor(s3Client *s3.S3) *ImageProcessor {
	return &ImageProcessor{
		s3Client:       s3Client,
		bucketName:     getBucketName(),
		maxWidth:       2000,
		maxHeight:      2000,
		thumbnailWidth: 300,
	}
}

// ProcessImage processes an image stored in S3
func (p *ImageProcessor) ProcessImage(ctx context.Context, event *ImageUploadEvent) (ProcessingResult, error) {
	log.Printf("Processing image for album %s, image %s", event.AlbumID, event.ImageID)

	// Get image from S3
	key := fmt.Sprintf("images/%s/%s", event.AlbumID, event.ImageID)
	result, err := p.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return ProcessingResult{}, fmt.Errorf("failed to get image from S3: %w", err)
	}
	defer result.Body.Close()

	// Read image data
	imageData, err := io.ReadAll(result.Body)
	if err != nil {
		return ProcessingResult{}, fmt.Errorf("failed to read image data: %w", err)
	}

	// Get content type
	contentType := aws.StringValue(result.ContentType)
	if contentType == "" {
		contentType = "image/jpeg" // Default if not specified
	}

	// Create temporary file to store original image
	tempFile, err := os.CreateTemp("", "image-*"+getExtensionFromContentType(contentType))
	if err != nil {
		return ProcessingResult{}, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write image data to temporary file
	if _, err := tempFile.Write(imageData); err != nil {
		return ProcessingResult{}, fmt.Errorf("failed to write image data to temp file: %w", err)
	}

	// Create thumbnail
	thumbnailID, err := p.createThumbnail(ctx, event.AlbumID, event.ImageID, imageData, contentType)
	if err != nil {
		log.Printf("Warning: failed to create thumbnail: %v", err)
		// Continue without thumbnail
	}

	return ProcessingResult{
		AlbumID:       event.AlbumID,
		ImageID:       event.ImageID,
		ThumbnailID:   thumbnailID,
		ProcessedSize: event.ImageSize,
		Success:       true,
	}, nil
}

// createThumbnail creates a thumbnail for the image
func (p *ImageProcessor) createThumbnail(ctx context.Context, albumID, imageID string, imageData []byte, contentType string) (string, error) {
	// Decode image
	img, format, err := decodeImage(imageData, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image to create thumbnail
	thumbnail := resize.Resize(p.thumbnailWidth, 0, img, resize.Lanczos3)

	// Encode thumbnail
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
			return "", fmt.Errorf("failed to encode JPEG thumbnail: %w", err)
		}
	case "png":
		if err := png.Encode(&buf, thumbnail); err != nil {
			return "", fmt.Errorf("failed to encode PNG thumbnail: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}

	// Create thumbnail ID
	thumbnailID := fmt.Sprintf("%s-thumb", imageID)

	// Upload thumbnail to S3
	thumbnailKey := fmt.Sprintf("images/%s/%s", albumID, thumbnailID)
	_, err = p.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucketName),
		Key:         aws.String(thumbnailKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"AlbumId":       aws.String(albumID),
			"Thumbnail":     aws.String("true"),
			"OriginalImage": aws.String(imageID),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload thumbnail to S3: %w", err)
	}

	return thumbnailID, nil
}

// decodeImage decodes image data based on content type
func decodeImage(data []byte, contentType string) (image.Image, string, error) {
	reader := bytes.NewReader(data)

	// Try to decode based on content type
	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		img, err := jpeg.Decode(reader)
		return img, "jpeg", err
	} else if strings.Contains(contentType, "png") {
		img, err := png.Decode(reader)
		return img, "png", err
	}

	// If content type didn't work, try to decode based on image data
	reader.Seek(0, 0) // Reset reader
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}
	return img, format, nil
}

// getExtensionFromContentType returns the file extension for a content type
func getExtensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	default:
		return ".jpg" // Default
	}
}

// getBucketName gets the S3 bucket name from environment variable or uses a default
func getBucketName() string {
	if bucket := os.Getenv("S3_BUCKET_NAME"); bucket != "" {
		return bucket
	}
	return "album-store-images"
}
