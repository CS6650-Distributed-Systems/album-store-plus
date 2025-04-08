package query

import (
	"time"
)

// AlbumView represents an album entity for query operations (read model)
type AlbumView struct {
	ID           string    `json:"id"`
	Artist       string    `json:"artist"`
	Title        string    `json:"title"`
	Year         string    `json:"year"`
	ImageID      string    `json:"imageId"`
	ImageSize    int64     `json:"imageSize"`
	LikeCount    int       `json:"likeCount"`
	DislikeCount int       `json:"dislikeCount"`
	Popularity   float64   `json:"popularity"` // Calculated field based on likes and other metrics
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// AlbumStats represents statistics for an album
type AlbumStats struct {
	AlbumID        string  `json:"albumId"`
	ViewCount      int64   `json:"viewCount"`
	LikeCount      int     `json:"likeCount"`
	DislikeCount   int     `json:"dislikeCount"`
	LikeRatio      float64 `json:"likeRatio"`      // Calculated as likes / (likes + dislikes)
	PopularityRank int     `json:"popularityRank"` // Rank among all albums
}

// AlbumSearchQuery represents search parameters
type AlbumSearchQuery struct {
	Artist    string `json:"artist,omitempty"`
	Title     string `json:"title,omitempty"`
	Year      string `json:"year,omitempty"`
	MinLikes  *int   `json:"minLikes,omitempty"`
	SortBy    string `json:"sortBy,omitempty"`    // popularity, date, title
	SortOrder string `json:"sortOrder,omitempty"` // asc, desc
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// SearchResult represents search results with pagination info
type SearchResult struct {
	Albums     []AlbumView `json:"albums"`
	TotalCount int         `json:"totalCount"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}
