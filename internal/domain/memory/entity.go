package memory

import (
	"time"

	"github.com/google/uuid"

	"mem_bank/internal/domain/user"
)

// ID represents a memory identifier
type ID uuid.UUID

// Memory represents a core business entity for memories
type Memory struct {
	ID           ID
	UserID       user.ID
	Content      string
	Summary      string
	Embedding    []float32
	Importance   int
	MemoryType   string
	Tags         []string
	Metadata     map[string]interface{}
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastAccessed time.Time
	AccessCount  int
}

// NewID creates a new memory ID
func NewID() ID {
	return ID(uuid.New())
}

// String returns the string representation of the memory ID
func (id ID) String() string {
	return uuid.UUID(id).String()
}

// IsZero checks if the ID is zero
func (id ID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// NewMemory creates a new memory with default values
func NewMemory(userID user.ID, content, summary string, importance int, memoryType string) *Memory {
	return &Memory{
		ID:           NewID(),
		UserID:       userID,
		Content:      content,
		Summary:      summary,
		Importance:   importance,
		MemoryType:   memoryType,
		Tags:         make([]string, 0),
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  0,
	}
}

// Access records an access to this memory
func (m *Memory) Access() {
	m.LastAccessed = time.Now()
	m.AccessCount++
}

// Update updates the memory content and metadata
func (m *Memory) Update(content, summary string, importance int, tags []string, metadata map[string]interface{}) {
	m.Content = content
	m.Summary = summary
	m.Importance = importance
	m.Tags = tags
	m.Metadata = metadata
	m.UpdatedAt = time.Now()
}

// UpdateEmbedding updates the embedding vector
func (m *Memory) UpdateEmbedding(embedding []float32) {
	m.Embedding = embedding
	m.UpdatedAt = time.Now()
}

// AddTag adds a tag if it doesn't exist
func (m *Memory) AddTag(tag string) {
	for _, t := range m.Tags {
		if t == tag {
			return
		}
	}
	m.Tags = append(m.Tags, tag)
	m.UpdatedAt = time.Now()
}

// RemoveTag removes a tag if it exists
func (m *Memory) RemoveTag(tag string) {
	for i, t := range m.Tags {
		if t == tag {
			m.Tags = append(m.Tags[:i], m.Tags[i+1:]...)
			m.UpdatedAt = time.Now()
			break
		}
	}
}

// HasTag checks if the memory has a specific tag
func (m *Memory) HasTag(tag string) bool {
	for _, t := range m.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// EmbeddingUpdate represents an embedding update operation for batch processing
type EmbeddingUpdate struct {
	ID        ID        `json:"id"`
	Embedding []float32 `json:"embedding"`
}

// MemoryWithScore represents a memory with its similarity score
type MemoryWithScore struct {
	Memory *Memory `json:"memory"`
	Score  float64 `json:"score"`
}
