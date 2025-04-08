package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/command"
	commandService "github.com/CS6650-Distributed-Systems/album-store-plus/internal/service/command"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Routes defines the HTTP routes for the command API
type Routes struct {
	albumService *commandService.AlbumService
}

// NewRoutes creates a new routes handler
func NewRoutes(albumService *commandService.AlbumService) *Routes {
	return &Routes{
		albumService: albumService,
	}
}

// RegisterRoutes registers all API routes
func (r *Routes) RegisterRoutes(router chi.Router) {
	// Add middleware
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recovery)
	router.Use(chimiddleware.Timeout(60 * 1000)) // 60 second timeout

	// API routes
	router.Route("/", func(router chi.Router) {
		// Health check
		router.Get("/health", r.handleHealth)

		// v1 API routes
		router.Route("/api/v1", func(router chi.Router) {
			// Album routes
			router.Route("/albums", func(router chi.Router) {
				router.Post("/", r.handleCreateAlbum)
				router.Post("/batch", r.handleBatchCreateAlbums)
			})

			// Review routes
			router.Route("/review", func(router chi.Router) {
				router.Post("/{likeornot}/{albumID}", r.handleReview)
			})
		})
	})

	// Not found handler
	router.NotFound(middleware.NotFound().ServeHTTP)
	router.MethodNotAllowed(middleware.MethodNotAllowed().ServeHTTP)
}

// handleHealth handles health check requests
func (r *Routes) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "command-api",
	})
}

// handleCreateAlbum handles album creation requests
func (r *Routes) handleCreateAlbum(w http.ResponseWriter, req *http.Request) {
	// Maximum file size (10MB)
	const maxFileSize = 10 << 20
	req.Body = http.MaxBytesReader(w, req.Body, maxFileSize)

	// Parse multipart form
	if err := req.ParseMultipartForm(maxFileSize); err != nil {
		middleware.HandleError(w, err, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get album metadata from form
	profileJSON := req.FormValue("profile")
	if profileJSON == "" {
		middleware.HandleError(w, fmt.Errorf("missing profile"), http.StatusBadRequest, "Missing album profile")
		return
	}

	// Parse profile JSON
	var profile struct {
		Artist string `json:"artist"`
		Title  string `json:"title"`
		Year   string `json:"year"`
	}
	if err := json.Unmarshal([]byte(profileJSON), &profile); err != nil {
		middleware.HandleError(w, err, http.StatusBadRequest, "Invalid profile format")
		return
	}

	// Get image file
	file, header, err := req.FormFile("image")
	if err != nil {
		middleware.HandleError(w, err, http.StatusBadRequest, "Failed to get image file")
		return
	}
	defer file.Close()

	// Read image data
	imageData, err := io.ReadAll(file)
	if err != nil {
		middleware.HandleError(w, err, http.StatusInternalServerError, "Failed to read image data")
		return
	}

	// Create album command
	cmd := &command.AlbumCreateCommand{
		Artist:    profile.Artist,
		Title:     profile.Title,
		Year:      profile.Year,
		ImageData: imageData,
	}

	// Create album
	album, err := r.albumService.CreateAlbum(req.Context(), cmd, imageData, header.Header.Get("Content-Type"))
	if err != nil {
		logging.GetLogger().Error("Failed to create album", zap.Error(err))
		middleware.HandleError(w, err, http.StatusInternalServerError, "Failed to create album")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"albumID":   album.ID,
		"imageSize": strconv.FormatInt(album.ImageSize, 10),
	})
}

// handleReview handles album review requests (like/dislike)
func (r *Routes) handleReview(w http.ResponseWriter, req *http.Request) {
	// Get path parameters
	albumID := chi.URLParam(req, "albumID")
	likeOrNot := chi.URLParam(req, "likeornot")

	// Validate parameters
	if albumID == "" {
		middleware.HandleError(w, fmt.Errorf("missing album ID"), http.StatusBadRequest, "Missing album ID")
		return
	}

	// Determine if it's a like or dislike
	var err error
	if likeOrNot == "like" {
		err = r.albumService.LikeAlbum(req.Context(), albumID)
	} else if likeOrNot == "dislike" {
		err = r.albumService.DislikeAlbum(req.Context(), albumID)
	} else {
		middleware.HandleError(w, fmt.Errorf("invalid review type"), http.StatusBadRequest, "Invalid review type, must be 'like' or 'dislike'")
		return
	}

	// Handle errors
	if err != nil {
		logging.GetLogger().Error("Failed to process review",
			zap.Error(err),
			zap.String("albumId", albumID),
			zap.String("reviewType", likeOrNot),
		)
		middleware.HandleError(w, err, http.StatusInternalServerError, "Failed to process review")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// BatchImageRequest represents an image in a batch request
type BatchImageRequest struct {
	Data        []byte
	ContentType string
	Profile     command.AlbumCreateCommand
}

// BatchAlbumResult represents the result of creating a single album in a batch
type BatchAlbumResult struct {
	AlbumID   string `json:"albumId"`
	ImageSize string `json:"imageSize"`
	Error     string `json:"error,omitempty"`
}

// handleBatchCreateAlbums handles batch album creation requests
func (r *Routes) handleBatchCreateAlbums(w http.ResponseWriter, req *http.Request) {
	// Maximum file size (50MB for batch)
	const maxFileSize = 50 << 20
	req.Body = http.MaxBytesReader(w, req.Body, maxFileSize)

	// Parse multipart form
	if err := req.ParseMultipartForm(maxFileSize); err != nil {
		middleware.HandleError(w, err, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Extract the form values
	form := req.MultipartForm
	imageFiles := form.File["images[]"]
	profiles := form.Value["profiles[]"]

	// Validate input
	if len(imageFiles) == 0 {
		middleware.HandleError(w, fmt.Errorf("no images provided"), http.StatusBadRequest, "No images provided")
		return
	}

	if len(imageFiles) != len(profiles) {
		middleware.HandleError(w, fmt.Errorf("mismatch between images and profiles count"), http.StatusBadRequest,
			"Number of images must match number of profiles")
		return
	}

	// Process each image
	commands := make([]*command.AlbumCreateCommand, len(imageFiles))
	imageDataArray := make([][]byte, len(imageFiles))
	contentTypes := make([]string, len(imageFiles))

	for i, fileHeader := range imageFiles {
		// Open the image file
		file, err := fileHeader.Open()
		if err != nil {
			middleware.HandleError(w, err, http.StatusBadRequest, fmt.Sprintf("Failed to open image %d", i))
			return
		}
		defer file.Close()

		// Read image data
		imageData, err := io.ReadAll(file)
		if err != nil {
			middleware.HandleError(w, err, http.StatusInternalServerError, fmt.Sprintf("Failed to read image %d", i))
			return
		}

		// Parse profile
		var profile struct {
			Artist string `json:"artist"`
			Title  string `json:"title"`
			Year   string `json:"year"`
		}
		if err := json.Unmarshal([]byte(profiles[i]), &profile); err != nil {
			middleware.HandleError(w, err, http.StatusBadRequest, fmt.Sprintf("Invalid profile format for image %d", i))
			return
		}

		// Create command
		commands[i] = &command.AlbumCreateCommand{
			Artist:    profile.Artist,
			Title:     profile.Title,
			Year:      profile.Year,
			ImageData: imageData,
		}
		imageDataArray[i] = imageData
		contentTypes[i] = fileHeader.Header.Get("Content-Type")
	}

	// Create albums in batch
	albums, err := r.albumService.BatchCreateAlbums(req.Context(), commands, imageDataArray, contentTypes)
	if err != nil {
		logging.GetLogger().Error("Failed to create albums in batch", zap.Error(err))
		middleware.HandleError(w, err, http.StatusInternalServerError, "Failed to create albums in batch")
		return
	}

	// Create response
	results := make([]map[string]string, len(albums))
	for i, album := range albums {
		results[i] = map[string]string{
			"albumID":   album.ID,
			"imageSize": strconv.FormatInt(album.ImageSize, 10),
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}
