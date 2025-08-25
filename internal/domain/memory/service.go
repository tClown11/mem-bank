package memory

import (
	"context"

	"mem_bank/internal/domain/user"
)

// Service defines the business operations for memories
type Service interface {
	// CreateMemory creates a new memory with validation and embedding generation
	CreateMemory(ctx context.Context, req CreateRequest) (*Memory, error)

	// GetMemory retrieves a memory by ID and updates access info
	GetMemory(ctx context.Context, id ID) (*Memory, error)

	// UpdateMemory updates an existing memory
	UpdateMemory(ctx context.Context, id ID, req UpdateRequest) (*Memory, error)

	// DeleteMemory deletes a memory by ID
	DeleteMemory(ctx context.Context, id ID) error

	// ListUserMemories returns a list of memories for a user with pagination
	ListUserMemories(ctx context.Context, userID user.ID, limit, offset int) ([]*Memory, error)

	// SearchMemories searches memories based on various criteria
	SearchMemories(ctx context.Context, req SearchRequest) ([]*Memory, error)

	// SearchSimilarMemories finds similar memories using embedding vectors
	SearchSimilarMemories(ctx context.Context, content string, userID user.ID, limit int, threshold float64) ([]*Memory, error)

	// GetMemoryStats returns memory statistics for a user
	GetMemoryStats(ctx context.Context, userID user.ID) (*Stats, error)
}
