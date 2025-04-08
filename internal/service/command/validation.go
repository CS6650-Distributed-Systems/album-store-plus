package command

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/command"
)

// Validator handles validation logic
type Validator struct {
	maxTitleLength  int
	maxArtistLength int
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		maxTitleLength:  100,
		maxArtistLength: 100,
	}
}

// ValidateAlbumCreate validates an album creation command
func (v *Validator) ValidateAlbumCreate(cmd *command.AlbumCreateCommand) error {
	if cmd == nil {
		return errors.New("command cannot be nil")
	}

	// Validate artist
	if err := v.validateArtist(cmd.Artist); err != nil {
		return err
	}

	// Validate title
	if err := v.validateTitle(cmd.Title); err != nil {
		return err
	}

	// Validate year
	if err := v.validateYear(cmd.Year); err != nil {
		return err
	}

	// Validate image data
	if len(cmd.ImageData) == 0 {
		return errors.New("image data is required")
	}

	return nil
}

// validateArtist validates an artist name
func (v *Validator) validateArtist(artist string) error {
	if artist == "" {
		return errors.New("artist is required")
	}

	if len(artist) > v.maxArtistLength {
		return fmt.Errorf("artist name exceeds maximum length of %d characters", v.maxArtistLength)
	}

	return nil
}

// validateTitle validates an album title
func (v *Validator) validateTitle(title string) error {
	if title == "" {
		return errors.New("title is required")
	}

	if len(title) > v.maxTitleLength {
		return fmt.Errorf("title exceeds maximum length of %d characters", v.maxTitleLength)
	}

	return nil
}

// validateYear validates an album year
func (v *Validator) validateYear(year string) error {
	if year == "" {
		return errors.New("year is required")
	}

	// Check if year is a 4-digit number
	yearRegex := regexp.MustCompile(`^\d{4}$`)
	if !yearRegex.MatchString(year) {
		return errors.New("year must be a 4-digit number")
	}

	// Check if year is within reasonable range
	currentYear := time.Now().Year()
	var yearInt int
	_, err := fmt.Sscanf(year, "%d", &yearInt)
	if err != nil {
		return fmt.Errorf("invalid year format: %w", err)
	}

	if yearInt < 1900 || yearInt > currentYear+1 {
		return fmt.Errorf("year must be between 1900 and %d", currentYear+1)
	}

	return nil
}

// ValidateAlbumID validates an album ID
func (v *Validator) ValidateAlbumID(albumID string) error {
	if albumID == "" {
		return errors.New("album ID is required")
	}

	// Add additional validation logic as needed for your ID format
	// For example, if you're using UUIDs:
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(strings.ToLower(albumID)) {
		return errors.New("invalid album ID format")
	}

	return nil
}

// ValidateImageContentType validates an image content type
func (v *Validator) ValidateImageContentType(contentType string) error {
	if contentType == "" {
		return errors.New("content type is required")
	}

	// List of allowed image content types
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
	}

	if !allowedTypes[contentType] {
		return fmt.Errorf("unsupported image format: %s. Supported formats: JPEG, PNG, GIF", contentType)
	}

	return nil
}

// ValidateImageSize validates an image size
func (v *Validator) ValidateImageSize(size int64) error {
	// Maximum image size (10MB)
	const maxSize = 10 * 1024 * 1024

	if size <= 0 {
		return errors.New("image size must be greater than 0")
	}

	if size > maxSize {
		return fmt.Errorf("image size exceeds maximum allowed size of 10MB")
	}

	return nil
}
