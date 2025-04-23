package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/service"
	"github.com/gorilla/mux"
)

// ReviewHandler handles HTTP requests for album reviews
type ReviewHandler struct {
	service service.ReviewService
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(service service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		service: service,
	}
}

// GetReview handles GET requests to retrieve a review
func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["id"]

	if albumID == "" {
		http.Error(w, "Album ID is required", http.StatusBadRequest)
		return
	}

	review, err := h.service.GetReview(r.Context(), albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving review: %v", err), http.StatusInternalServerError)
		return
	}

	if review == nil {
		// Return empty review with zero counts if none exists
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"album_id":      albumID,
			"like_count":    0,
			"dislike_count": 0,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(review); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleReviewAction is a helper method to process common review actions (like/dislike)
func (h *ReviewHandler) handleReviewAction(w http.ResponseWriter, r *http.Request, actionType string, actionFunc func(context.Context, string) error) {
	vars := mux.Vars(r)
	albumID := vars["id"]

	if albumID == "" {
		http.Error(w, "Album ID is required", http.StatusBadRequest)
		return
	}

	// Queue the operation
	if err := actionFunc(r.Context(), albumID); err != nil {
		http.Error(w, fmt.Sprintf("Error processing %s: %v", actionType, err), http.StatusInternalServerError)
		return
	}

	// Return immediate success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("%s operation queued", actionType),
		"status":  "accepted",
	})
}

// LikeAlbum handles POST requests to add a like to an album
func (h *ReviewHandler) LikeAlbum(w http.ResponseWriter, r *http.Request) {
	h.handleReviewAction(w, r, "Like", h.service.LikeAlbum)
}

// DislikeAlbum handles POST requests to add a dislike to an album
func (h *ReviewHandler) DislikeAlbum(w http.ResponseWriter, r *http.Request) {
	h.handleReviewAction(w, r, "Dislike", h.service.DislikeAlbum)
}

// DeleteReview handles DELETE requests to remove a review
func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	albumID := vars["id"]

	if albumID == "" {
		http.Error(w, "Album ID is required", http.StatusBadRequest)
		return
	}

	// Queue the delete operation
	if err := h.service.DeleteReview(r.Context(), albumID); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting review: %v", err), http.StatusInternalServerError)
		return
	}

	// Return immediate success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Delete operation queued",
		"status":  "accepted",
	})
}
