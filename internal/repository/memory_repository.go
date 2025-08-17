package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mem_bank/internal/domain"
	"mem_bank/internal/model"
	"mem_bank/internal/query"
)

type memoryRepository struct {
	db *gorm.DB
	q  *query.Query
}

func NewMemoryRepository(db *gorm.DB) domain.MemoryRepository {
	return &memoryRepository{
		db: db,
		q:  query.Use(db),
	}
}

func (r *memoryRepository) Create(memory *domain.Memory) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	memory.ID = uuid.New()
	memory.CreatedAt = time.Now()
	memory.UpdatedAt = time.Now()
	memory.LastAccessed = time.Now()

	metadataJSON, err := json.Marshal(memory.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Convert to GORM model - temporarily skip embedding
	gormMemory := &model.Memory{
		ID:           memory.ID.String(),
		UserID:       memory.UserID.String(),
		Content:      memory.Content,
		Summary:      stringPtr(memory.Summary),
		Importance:   intPtr(int32(memory.Importance)),
		MemoryType:   stringPtr(memory.MemoryType),
		Metadata:     stringPtr(string(metadataJSON)),
		CreatedAt:    &memory.CreatedAt,
		UpdatedAt:    &memory.UpdatedAt,
		LastAccessed: &memory.LastAccessed,
		AccessCount:  intPtr(int32(memory.AccessCount)),
	}

	err = r.q.Memory.WithContext(ctx).Create(gormMemory)
	if err != nil {
		return fmt.Errorf("failed to create memory: %w", err)
	}

	return nil
}

// Placeholder implementations for other methods - these need to be implemented properly
func (r *memoryRepository) GetByID(id uuid.UUID) (*domain.Memory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *memoryRepository) Update(memory *domain.Memory) error {
	return fmt.Errorf("not implemented")
}

func (r *memoryRepository) Delete(id uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *memoryRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *memoryRepository) SearchSimilar(embedding []float32, userID uuid.UUID, limit int, threshold float64) ([]*domain.Memory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *memoryRepository) SearchByContent(query string, userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *memoryRepository) GetByTags(tags []string, userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *memoryRepository) GetStatsByUserID(userID uuid.UUID) (*domain.MemoryStats, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *memoryRepository) UpdateAccessInfo(id uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

// Helper functions
func intPtr(i int32) *int32 {
	return &i
}