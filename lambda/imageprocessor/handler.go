package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

// ImageUploadEvent represents an image upload event from SNS/SQS
type ImageUploadEvent struct {
	AlbumID            string `json:"albumId"`
	ImageID            string `json:"imageId"`
	ImageSize          int64  `json:"imageSize"`
	RequiresProcessing bool   `json:"requiresProcessing"`
}

// ProcessingResult represents the result of image processing
type ProcessingResult struct {
	AlbumID       string `json:"albumId"`
	ImageID       string `json:"imageId"`
	ThumbnailID   string `json:"thumbnailId,omitempty"`
	ProcessedSize int64  `json:"processedSize,omitempty"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
}

// HandleRequest is the Lambda handler function
func (p *ImageProcessor) HandleRequest(ctx context.Context, event json.RawMessage) (ProcessingResult, error) {
	log.Printf("Received event: %s", string(event))

	// Parse event
	var uploadEvent ImageUploadEvent
	if err := json.Unmarshal(event, &uploadEvent); err != nil {
		// Try to parse as SQS event
		var sqsEvent events.SQSEvent
		if sqsErr := json.Unmarshal(event, &sqsEvent); sqsErr == nil {
			// It's an SQS event, extract the first record
			if len(sqsEvent.Records) > 0 {
				log.Printf("Processing SQS event with %d records", len(sqsEvent.Records))
				record := sqsEvent.Records[0]
				if err := json.Unmarshal([]byte(record.Body), &uploadEvent); err != nil {
					return ProcessingResult{
						Success: false,
						Error:   fmt.Sprintf("Failed to parse SQS message body: %v", err),
					}, nil
				}
			} else {
				return ProcessingResult{
					Success: false,
					Error:   "SQS event contains no records",
				}, nil
			}
		} else {
			// It's not an SQS event or image upload event
			return ProcessingResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to parse event: %v", err),
			}, nil
		}
	}

	// Validate
	if uploadEvent.AlbumID == "" || uploadEvent.ImageID == "" {
		return ProcessingResult{
			Success: false,
			Error:   "Invalid event: albumId and imageId are required",
		}, nil
	}

	// Process image
	result, err := p.ProcessImage(ctx, &uploadEvent)
	if err != nil {
		log.Printf("Error processing image: %v", err)
		return ProcessingResult{
			AlbumID: uploadEvent.AlbumID,
			ImageID: uploadEvent.ImageID,
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return result, nil
}
