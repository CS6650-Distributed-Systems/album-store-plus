package query

// FacetedSearch represents a search with facets (aggregations)
type FacetedSearch struct {
	Query  AlbumSearchQuery `json:"query"`
	Facets []Facet          `json:"facets"`
}

// Facet represents a category for filtering and aggregation
type Facet struct {
	Name   string       `json:"name"`
	Type   string       `json:"type"` // range, term, etc.
	Values []FacetValue `json:"values"`
}

// FacetValue represents a value in a facet with document count
type FacetValue struct {
	Value interface{} `json:"value"`
	Count int         `json:"count"`
}

// FacetedSearchResult represents search results with facets
type FacetedSearchResult struct {
	SearchResult
	Facets []Facet `json:"facets"`
}

// Common facet names
const (
	FacetYear   = "year"
	FacetArtist = "artist"
	FacetLikes  = "likes"
)

// YearRange represents a range of years
type YearRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// LikeRange represents a range of like counts
type LikeRange struct {
	From int `json:"from"`
	To   int `json:"to"`
}
