package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/domain"
	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/service"
	"github.com/gorilla/mux"
)

// AlbumHandler handles HTTP requests for albums
type AlbumHandler struct {
	service service.AlbumService
}

// NewAlbumHandler creates a new album handler
func NewAlbumHandler(service service.AlbumService) *AlbumHandler {
	return &AlbumHandler{
		service: service,
	}
}

// GetAlbum handles GET requests to retrieve an album
func (h *AlbumHandler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["id"]

	// Parse query parameters
	includeArtist := r.URL.Query().Get("include_artist") == "true"
	includeReview := r.URL.Query().Get("include_review") == "true"

	album, err := h.service.GetAlbum(r.Context(), albumID, includeArtist, includeReview)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving album: %v", err), http.StatusInternalServerError)
		return
	}

	if album == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(album); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetAlbumsByArtist handles GET requests to retrieve albums by artist
func (h *AlbumHandler) GetAlbumsByArtist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	artistID := vars["artistId"]

	albums, err := h.service.GetAlbumsByArtist(r.Context(), artistID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving albums: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(albums); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// CreateAlbum handles POST requests to create an album
func (h *AlbumHandler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	// First, check if this is a multipart form request (with image)
	isMultipart := strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data")

	var album domain.Album

	var file multipart.File
	var fileHeader *multipart.FileHeader
	var err error

	if isMultipart {
		// Parse multipart form with 10MB limit
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}

		// Get album data from form field
		albumData := r.FormValue("album")
		if albumData == "" {
			http.Error(w, "Missing album data", http.StatusBadRequest)
			return
		}

		// Parse album JSON
		if err := json.Unmarshal([]byte(albumData), &album); err != nil {
			http.Error(w, "Invalid album data format", http.StatusBadRequest)
			return
		}

		// Get image file
		file, fileHeader, err = r.FormFile("image")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()
	} else {
		// Regular JSON request without image
		if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	// Create the album
	if err := h.service.CreateAlbum(r.Context(), &album); err != nil {
		http.Error(w, fmt.Sprintf("Error creating album: %v", err), http.StatusInternalServerError)
		return
	}

	// Process image if it was uploaded
	if isMultipart && file != nil {
		if err := h.service.UploadAlbumCover(r.Context(), album.ID, file, fileHeader.Filename); err != nil {
			// Don't fail the whole request if image upload fails
			log.Printf("Error uploading cover for new album %s: %v", album.ID, err)
		}

		// Refresh album to get updated image info
		updatedAlbum, err := h.service.GetAlbum(r.Context(), album.ID, false, false)
		if err == nil && updatedAlbum != nil {
			album = *updatedAlbum
		}
	}

	// Return the created album
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(album); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// UpdateAlbum handles PUT requests to update an album
func (h *AlbumHandler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["id"]

	var album domain.Album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	album.ID = albumID

	if err := h.service.UpdateAlbum(r.Context(), &album); err != nil {
		http.Error(w, fmt.Sprintf("Error updating album: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(album); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// DeleteAlbum handles DELETE requests to delete an album
func (h *AlbumHandler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["id"]

	if err := h.service.DeleteAlbum(r.Context(), albumID); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting album: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadAlbumCover handles POST requests to upload an album cover
func (h *AlbumHandler) UploadAlbumCover(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["id"]

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Process and upload the image
	if err := h.service.UploadAlbumCover(r.Context(), albumID, file, fileHeader.Filename); err != nil {
		http.Error(w, fmt.Sprintf("Error uploading cover: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the album with updated image info
	album, err := h.service.GetAlbum(r.Context(), albumID, false, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving updated album: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(album)
}

// GetAlbumCover handles GET requests for an album cover
func (h *AlbumHandler) GetAlbumCover(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["id"]

	url, err := h.service.GetAlbumCoverURL(r.Context(), albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving cover: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to the actual image URL
	http.Redirect(w, r, url, http.StatusFound)
}
