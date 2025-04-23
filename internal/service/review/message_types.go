package review

// ReviewMessageType represents the type of review operation
type ReviewMessageType string

const (
	AddLikeType    ReviewMessageType = "add_like"
	AddDislikeType ReviewMessageType = "add_dislike"
	DeleteType     ReviewMessageType = "delete"
)

// ReviewMessage represents a message for review operations
type ReviewMessage struct {
	Type    ReviewMessageType `json:"type"`
	AlbumID string            `json:"album_id"`
}
