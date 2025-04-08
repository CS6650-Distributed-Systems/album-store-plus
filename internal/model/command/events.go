package command

import (
	"time"
)

// EventType represents the type of domain event
type EventType string

const (
	// Event types
	EventTypeAlbumCreated  EventType = "ALBUM_CREATED"
	EventTypeAlbumLiked    EventType = "ALBUM_LIKED"
	EventTypeAlbumDisliked EventType = "ALBUM_DISLIKED"
	EventTypeImageUploaded EventType = "IMAGE_UPLOADED"
)

// DomainEvent represents a domain event that occurred in the system
type DomainEvent struct {
	ID        string            `json:"id"`
	Type      EventType         `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Payload   interface{}       `json:"payload"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewDomainEvent creates a new domain event
func NewDomainEvent(id string, eventType EventType, payload interface{}, metadata map[string]string) *DomainEvent {
	return &DomainEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
		Metadata:  metadata,
	}
}

// AlbumCreatedEvent payload for album creation events
type AlbumCreatedEvent struct {
	AlbumID   string `json:"albumId"`
	Artist    string `json:"artist"`
	Title     string `json:"title"`
	Year      string `json:"year"`
	ImageID   string `json:"imageId"`
	ImageSize int64  `json:"imageSize"`
}

// AlbumReviewEvent payload for album like/dislike events
type AlbumReviewEvent struct {
	AlbumID string `json:"albumId"`
	Liked   bool   `json:"liked"`
}

// ImageUploadedEvent payload for image upload events
type ImageUploadedEvent struct {
	AlbumID            string `json:"albumId"`
	ImageID            string `json:"imageId"`
	ImageSize          int64  `json:"imageSize"`
	RequiresProcessing bool   `json:"requiresProcessing"`
}
