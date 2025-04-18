package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all the routes for the API
func SetupRoutes(
	albumHandler *AlbumHandler,
	//artistHandler *ArtistHandler,
	reviewHandler *ReviewHandler,
) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Album routes
	router.HandleFunc("/albums/{id}", albumHandler.GetAlbum).Methods("GET")
	router.HandleFunc("/artists/{artistId}/albums", albumHandler.GetAlbumsByArtist).Methods("GET")
	router.HandleFunc("/albums", albumHandler.CreateAlbum).Methods("POST")
	router.HandleFunc("/albums/{id}", albumHandler.UpdateAlbum).Methods("PUT")
	router.HandleFunc("/albums/{id}", albumHandler.DeleteAlbum).Methods("DELETE")
	router.HandleFunc("/albums/{id}/cover", albumHandler.UploadAlbumCover).Methods("POST")
	router.HandleFunc("/albums/{id}/cover", albumHandler.GetAlbumCover).Methods("GET")

	// Artist routes
	//router.HandleFunc("/artists/{id}", artistHandler.GetArtist).Methods("GET")
	//router.HandleFunc("/artists", artistHandler.CreateArtist).Methods("POST")
	//router.HandleFunc("/artists/{id}", artistHandler.UpdateArtist).Methods("PUT")
	//router.HandleFunc("/artists/{id}", artistHandler.DeleteArtist).Methods("DELETE")

	// Review routes
	router.HandleFunc("/albums/{id}/review", reviewHandler.GetReview).Methods("GET")
	router.HandleFunc("/albums/{id}/like", reviewHandler.LikeAlbum).Methods("POST")
	router.HandleFunc("/albums/{id}/dislike", reviewHandler.DislikeAlbum).Methods("POST")
	router.HandleFunc("/albums/{id}/review", reviewHandler.DeleteReview).Methods("DELETE")

	return router
}
