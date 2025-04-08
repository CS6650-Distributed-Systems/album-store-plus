package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/CS6650-Distributed-Systems/album-store-plus/internal/model/query"
	queryService "github.com/CS6650-Distributed-Systems/album-store-plus/internal/service/query"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/logging"
	"github.com/CS6650-Distributed-Systems/album-store-plus/pkg/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Routes defines the HTTP routes for the query API
type Routes struct {
	albumQueryService *queryService.AlbumQueryService
	searchService     *queryService.SearchService
}

// NewRoutes creates a new routes handler
func NewRoutes(albumQueryService *queryService.AlbumQueryService, searchService *queryService.SearchService) *Routes {
	return &Routes{
		albumQueryService: albumQueryService,
		searchService:     searchService,
	}
}

// RegisterRoutes registers all API routes
func (r *Routes) RegisterRoutes(router chi.Router) {
	// Add middleware
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recovery)
	router.Use(chimiddleware.Timeout(30 * 1000)) // 30 second timeout

	// API routes
	router.Route("/", func(router chi.Router) {
		// Health check
		router.Get("/health", r.handleHealth)

		// v1 API routes
		router.Route("/api/v1", func(router chi.Router) {
			// Album routes
			router.Route("/albums", func(router chi.Router) {
				router.Get("/{albumID}", r.handleGetAlbum)
				router.Get("/{albumID}/stats", r.handleGetAlbumStats)
				router.Get("/search", r.handleSearchAlbums)
				router.Get("/{albumID}/image", r.handleGetAlbumImage)
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
		"service": "query-api",
	})
}

// handleGetAlbum handles get album requests
func (r *Routes) handleGetAlbum(w http.ResponseWriter, req *http.Request) {
	// Get album ID from path
	albumID := chi.URLParam(req, "albumID")
	if albumID == "" {
		middleware.HandleError(w, fmt.Errorf("missing album ID"), http.StatusBadRequest, "Missing album ID")
		return
	}

	// Get album
	album, err := r.albumQueryService.GetAlbum(req.Context(), albumID)
	if err != nil {
		logging.GetLogger().Error("Failed to get album",
			zap.Error(err),
			zap.String("albumId", albumID),
		)
		middleware.HandleError(w, err, http.StatusNotFound, "Album not found")
		return
	}

	// Format response as expected by the API
	response := map[string]string{
		"artist": album.Artist,
		"title":  album.Title,
		"year":   album.Year,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleGetAlbumStats handles get album statistics requests
func (r *Routes) handleGetAlbumStats(w http.ResponseWriter, req *http.Request) {
	// Get album ID from path
	albumID := chi.URLParam(req, "albumID")
	if albumID == "" {
		middleware.HandleError(w, fmt.Errorf("missing album ID"), http.StatusBadRequest, "Missing album ID")
		return
	}

	// Get album statistics
	stats, err := r.albumQueryService.GetAlbumStats(req.Context(), albumID)
	if err != nil {
		logging.GetLogger().Error("Failed to get album statistics",
			zap.Error(err),
			zap.String("albumId", albumID),
		)
		middleware.HandleError(w, err, http.StatusNotFound, "Album statistics not found")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// handleSearchAlbums handles album search requests
func (r *Routes) handleSearchAlbums(w http.ResponseWriter, req *http.Request) {
	// Parse query parameters
	query := &query.AlbumSearchQuery{
		Artist: req.URL.Query().Get("artist"),
		Title:  req.URL.Query().Get("title"),
		Year:   req.URL.Query().Get("year"),
		SortBy: req.URL.Query().Get("sortBy"),
	}

	// Parse numeric parameters
	if minLikesStr := req.URL.Query().Get("minLikes"); minLikesStr != "" {
		minLikes, err := strconv.Atoi(minLikesStr)
		if err != nil {
			middleware.HandleError(w, err, http.StatusBadRequest, "Invalid minLikes parameter")
			return
		}
		query.MinLikes = &minLikes
	}

	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			middleware.HandleError(w, err, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
		query.Limit = limit
	} else {
		query.Limit = 10 // Default limit
	}

	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			middleware.HandleError(w, err, http.StatusBadRequest, "Invalid offset parameter")
			return
		}
		query.Offset = offset
	}

	// Check if faceted search is requested
	facetsStr := req.URL.Query().Get("facets")
	if facetsStr != "" {
		facets := strings.Split(facetsStr, ",")
		result, err := r.searchService.FacetedSearch(req.Context(), query, facets)
		if err != nil {
			logging.GetLogger().Error("Failed to perform faceted search", zap.Error(err))
			middleware.HandleError(w, err, http.StatusInternalServerError, "Failed to perform faceted search")
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		return
	}

	// Regular search
	result, err := r.searchService.Search(req.Context(), query)
	if err != nil {
		logging.GetLogger().Error("Failed to search albums", zap.Error(err))
		middleware.HandleError(w, err, http.StatusInternalServerError, "Failed to search albums")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// handleGetAlbumImage handles get album image requests
func (r *Routes) handleGetAlbumImage(w http.ResponseWriter, req *http.Request) {
	// Get album ID from path
	albumID := chi.URLParam(req, "albumID")
	if albumID == "" {
		middleware.HandleError(w, fmt.Errorf("missing album ID"), http.StatusBadRequest, "Missing album ID")
		return
	}

	// Get album
	album, err := r.albumQueryService.GetAlbum(req.Context(), albumID)
	if err != nil {
		logging.GetLogger().Error("Failed to get album for image",
			zap.Error(err),
			zap.String("albumId", albumID),
		)
		middleware.HandleError(w, err, http.StatusNotFound, "Album not found")
		return
	}

	// Get image URL
	imageURL, err := r.albumQueryService.GetImageURL(req.Context(), albumID, album.ImageID)
	if err != nil {
		logging.GetLogger().Error("Failed to get image URL",
			zap.Error(err),
			zap.String("albumId", albumID),
			zap.String("imageId", album.ImageID),
		)
		middleware.HandleError(w, err, http.StatusInternalServerError, "Failed to get image URL")
		return
	}

	// Redirect to image URL
	http.Redirect(w, req, imageURL, http.StatusTemporaryRedirect)
}
