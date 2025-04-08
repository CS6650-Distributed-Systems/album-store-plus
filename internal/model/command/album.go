package command

import (
	"time"
)

// Album represents an album entity for command operations (write model)
type Album struct {
	ID        string    `json:"id"`
	Artist    string    `json:"artist"`
	Title     string    `json:"title"`
	Year      string    `json:"year"`
	ImageID   string    `json:"imageId"`
	ImageSize int64     `json:"imageSize"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewAlbum creates a new Album instance
func NewAlbum(id, artist, title, year, imageID string, imageSize int64) *Album {
	now := time.Now()
	return &Album{
		ID:        id,
		Artist:    artist,
		Title:     title,
		Year:      year,
		ImageID:   imageID,
		ImageSize: imageSize,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AlbumCreateCommand represents a command to create a new album
type AlbumCreateCommand struct {
	Artist    string `json:"artist"`
	Title     string `json:"title"`
	Year      string `json:"year"`
	ImageData []byte `json:"-"` // Not serialized to JSON
}