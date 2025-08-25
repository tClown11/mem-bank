package memory

import (
	"context"

	"mem_bank/internal/domain/user"
)

// Repository defines the interface for memory data access operations
type Repository interface {
	// Store creates a new memory in the repository
	Store(ctx context.Context, memory *Memory) error

	// FindByID retrieves a memory by its ID
	FindByID(ctx context.Context, id ID) (*Memory, error)

	// FindByUserID retrieves memories for a specific user with pagination
	FindByUserID(ctx context.Context, userID user.ID, limit, offset int) ([]*Memory, error)

	// Update updates an existing memory
	Update(ctx context.Context, memory *Memory) error

	// Delete removes a memory by its ID
	Delete(ctx context.Context, id ID) error

	// SearchSimilar finds similar memories based on embedding vector
	SearchSimilar(ctx context.Context, embedding []float32, userID user.ID, limit int, threshold float64) ([]*Memory, error)

	// SearchByContent searches memories by content text
	SearchByContent(ctx context.Context, query string, userID user.ID, limit, offset int) ([]*Memory, error)

	// FindByTags retrieves memories by tags for a specific user
	FindByTags(ctx context.Context, tags []string, userID user.ID, limit, offset int) ([]*Memory, error)

	// UpdateAccessInfo updates the access information (last accessed time and count)
	UpdateAccessInfo(ctx context.Context, id ID) error

	// GetStatsByUserID returns memory statistics for a user
	GetStatsByUserID(ctx context.Context, userID user.ID) (*Stats, error)

	// CountByUserID returns the total number of memories for a user
	CountByUserID(ctx context.Context, userID user.ID) (int, error)
}
