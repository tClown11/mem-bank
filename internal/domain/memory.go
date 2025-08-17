package domain

import (
	"time"

	"github.com/google/uuid"
)

type Memory struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Content     string    `json:"content" db:"content"`
	Summary     string    `json:"summary" db:"summary"`
	Embedding   []float32 `json:"embedding" db:"embedding"`
	Importance  int       `json:"importance" db:"importance"`
	MemoryType  string    `json:"memory_type" db:"memory_type"`
	Tags        []string  `json:"tags" db:"tags"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	LastAccessed time.Time `json:"last_accessed" db:"last_accessed"`
	AccessCount int       `json:"access_count" db:"access_count"`
}

type MemoryRepository interface {
	Create(memory *Memory) error
	GetByID(id uuid.UUID) (*Memory, error)
	GetByUserID(userID uuid.UUID, limit, offset int) ([]*Memory, error)
	Update(memory *Memory) error
	Delete(id uuid.UUID) error
	SearchSimilar(embedding []float32, userID uuid.UUID, limit int, threshold float64) ([]*Memory, error)
	SearchByContent(query string, userID uuid.UUID, limit, offset int) ([]*Memory, error)
	GetByTags(tags []string, userID uuid.UUID, limit, offset int) ([]*Memory, error)
	UpdateAccessInfo(id uuid.UUID) error
	GetStatsByUserID(userID uuid.UUID) (*MemoryStats, error)
}

type MemoryStats struct {
	TotalMemories   int `json:"total_memories"`
	MemoryTypes     map[string]int `json:"memory_types"`
	RecentMemories  int `json:"recent_memories"`
	AverageImportance float64 `json:"average_importance"`
}

type MemorySearchRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	Query     string    `json:"query"`
	Tags      []string  `json:"tags,omitempty"`
	MemoryType string   `json:"memory_type,omitempty"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
	Threshold float64   `json:"threshold,omitempty"`
}

type MemoryCreateRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	Content    string    `json:"content"`
	Summary    string    `json:"summary,omitempty"`
	Importance int       `json:"importance"`
	MemoryType string    `json:"memory_type"`
	Tags       []string  `json:"tags,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type MemoryUpdateRequest struct {
	Content    *string   `json:"content,omitempty"`
	Summary    *string   `json:"summary,omitempty"`
	Importance *int      `json:"importance,omitempty"`
	MemoryType *string   `json:"memory_type,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}