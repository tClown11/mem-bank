package memory

import "mem_bank/internal/domain/user"

// CreateRequest represents a request to create a new memory
type CreateRequest struct {
	UserID     user.ID
	Content    string
	Summary    string
	Importance int
	MemoryType string
	Tags       []string
	Metadata   map[string]interface{}
}

// UpdateRequest represents a request to update an existing memory
type UpdateRequest struct {
	Content    *string
	Summary    *string
	Importance *int
	MemoryType *string
	Tags       []string
	Metadata   map[string]interface{}
}

// SearchRequest represents a request to search memories
type SearchRequest struct {
	UserID     user.ID
	Query      string
	Tags       []string
	MemoryType string
	Limit      int
	Offset     int
	Threshold  float64
}

// Stats represents memory statistics for a user
type Stats struct {
	TotalMemories     int
	MemoryTypes       map[string]int
	RecentMemories    int
	AverageImportance float64
}
