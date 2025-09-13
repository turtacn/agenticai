// pkg/types/storage.go
package types

// Vector represents a high-dimensional vector for similarity search.
type Vector struct {
	ID     string    `json:"id"`
	Vector []float32 `json:"vector"`
	// Metadata stores arbitrary data associated with the vector.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResult represents a single item in a vector search result set.
type SearchResult struct {
	Vector Vector  `json:"vector"`
	Score  float32 `json:"score"`
}
