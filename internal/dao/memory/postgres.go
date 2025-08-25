package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
	"mem_bank/internal/model"
	"mem_bank/internal/query"
)

// postgresRepository implements memory.Repository using PostgreSQL
type postgresRepository struct {
	db *gorm.DB
	q  *query.Query
}

// NewPostgresRepository creates a new PostgreSQL-based memory repository
func NewPostgresRepository(db *gorm.DB) memory.Repository {
	return &postgresRepository{
		db: db,
		q:  query.Use(db),
	}
}

func (r *postgresRepository) Store(ctx context.Context, m *memory.Memory) error {
	gormMemory, err := r.toModel(m)
	if err != nil {
		return fmt.Errorf("converting to model: %w", err)
	}

	if err := r.q.Memory.WithContext(ctx).Create(gormMemory); err != nil {
		return fmt.Errorf("creating memory: %w", err)
	}

	return nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id memory.ID) (*memory.Memory, error) {
	gormMemory, err := r.q.Memory.WithContext(ctx).Where(r.q.Memory.ID.Eq(id.String())).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, memory.ErrNotFound
		}
		return nil, fmt.Errorf("finding memory: %w", err)
	}

	return r.toDomain(gormMemory)
}

func (r *postgresRepository) FindByUserID(ctx context.Context, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	gormMemories, err := r.q.Memory.WithContext(ctx).
		Where(r.q.Memory.UserID.Eq(userID.String())).
		Order(r.q.Memory.CreatedAt.Desc()).
		Limit(limit).
		Offset(offset).
		Find()
	if err != nil {
		return nil, fmt.Errorf("finding memories by user ID: %w", err)
	}

	memories := make([]*memory.Memory, 0, len(gormMemories))
	for _, gormMemory := range gormMemories {
		m, err := r.toDomain(gormMemory)
		if err != nil {
			return nil, fmt.Errorf("converting memory: %w", err)
		}
		memories = append(memories, m)
	}

	return memories, nil
}

func (r *postgresRepository) Update(ctx context.Context, m *memory.Memory) error {
	gormMemory, err := r.toModel(m)
	if err != nil {
		return fmt.Errorf("converting to model: %w", err)
	}

	result, err := r.q.Memory.WithContext(ctx).Where(r.q.Memory.ID.Eq(m.ID.String())).Updates(gormMemory)
	if err != nil {
		return fmt.Errorf("updating memory: %w", err)
	}

	if result.RowsAffected == 0 {
		return memory.ErrNotFound
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id memory.ID) error {
	result, err := r.q.Memory.WithContext(ctx).Where(r.q.Memory.ID.Eq(id.String())).Delete()
	if err != nil {
		return fmt.Errorf("deleting memory: %w", err)
	}

	if result.RowsAffected == 0 {
		return memory.ErrNotFound
	}

	return nil
}

func (r *postgresRepository) SearchSimilar(ctx context.Context, embedding []float32, userID user.ID, limit int, threshold float64) ([]*memory.Memory, error) {
	// TODO: Implement vector similarity search once embedding support is added
	// For now, return empty results
	return []*memory.Memory{}, nil
}

func (r *postgresRepository) SearchByContent(ctx context.Context, query string, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	searchTerm := fmt.Sprintf("%%%s%%", query)
	gormMemories, err := r.q.Memory.WithContext(ctx).
		Where(r.q.Memory.UserID.Eq(userID.String())).
		Where(r.q.Memory.Content.Like(searchTerm)).
		Order(r.q.Memory.CreatedAt.Desc()).
		Limit(limit).
		Offset(offset).
		Find()
	if err != nil {
		return nil, fmt.Errorf("searching memories by content: %w", err)
	}

	memories := make([]*memory.Memory, 0, len(gormMemories))
	for _, gormMemory := range gormMemories {
		m, err := r.toDomain(gormMemory)
		if err != nil {
			return nil, fmt.Errorf("converting memory: %w", err)
		}
		memories = append(memories, m)
	}

	return memories, nil
}

func (r *postgresRepository) FindByTags(ctx context.Context, tags []string, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	if len(tags) == 0 {
		return []*memory.Memory{}, nil
	}

	// Use PostgreSQL array overlap operator (&&) to find memories with any of the specified tags
	// This is much more efficient than LIKE queries and works properly with PostgreSQL arrays
	var gormMemories []*model.Memory
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND tags && ?", userID.String(), pq.Array(tags)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&gormMemories).Error
	if err != nil {
		return nil, fmt.Errorf("finding memories by tags: %w", err)
	}

	memories := make([]*memory.Memory, 0, len(gormMemories))
	for _, gormMemory := range gormMemories {
		m, err := r.toDomain(gormMemory)
		if err != nil {
			return nil, fmt.Errorf("converting memory: %w", err)
		}
		memories = append(memories, m)
	}

	return memories, nil
}

func (r *postgresRepository) UpdateAccessInfo(ctx context.Context, id memory.ID) error {
	now := time.Now()
	result, err := r.q.Memory.WithContext(ctx).
		Where(r.q.Memory.ID.Eq(id.String())).
		Updates(map[string]interface{}{
			"last_accessed": &now,
			"access_count":  gorm.Expr("access_count + 1"),
		})
	if err != nil {
		return fmt.Errorf("updating access info: %w", err)
	}

	if result.RowsAffected == 0 {
		return memory.ErrNotFound
	}

	return nil
}

func (r *postgresRepository) GetStatsByUserID(ctx context.Context, userID user.ID) (*memory.Stats, error) {
	// Get total count
	totalCount, err := r.q.Memory.WithContext(ctx).Where(r.q.Memory.UserID.Eq(userID.String())).Count()
	if err != nil {
		return nil, fmt.Errorf("counting memories: %w", err)
	}

	// Get recent count (last 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7)
	recentCount, err := r.q.Memory.WithContext(ctx).
		Where(r.q.Memory.UserID.Eq(userID.String())).
		Where(r.q.Memory.CreatedAt.Gte(weekAgo)).
		Count()
	if err != nil {
		return nil, fmt.Errorf("counting recent memories: %w", err)
	}

	// Get average importance
	var avgImportance float64
	err = r.db.WithContext(ctx).
		Model(&model.Memory{}).
		Where("user_id = ?", userID.String()).
		Select("AVG(COALESCE(importance, 5))").
		Scan(&avgImportance).Error
	if err != nil {
		return nil, fmt.Errorf("calculating average importance: %w", err)
	}

	// For now, return basic stats without memory type breakdown
	// TODO: Implement proper memory type counting
	return &memory.Stats{
		TotalMemories:     int(totalCount),
		MemoryTypes:       make(map[string]int),
		RecentMemories:    int(recentCount),
		AverageImportance: avgImportance,
	}, nil
}

func (r *postgresRepository) CountByUserID(ctx context.Context, userID user.ID) (int, error) {
	count, err := r.q.Memory.WithContext(ctx).Where(r.q.Memory.UserID.Eq(userID.String())).Count()
	if err != nil {
		return 0, fmt.Errorf("counting user memories: %w", err)
	}

	return int(count), nil
}

// Helper functions
func intPtr(i int32) *int32 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func (r *postgresRepository) toModel(m *memory.Memory) (*model.Memory, error) {
	metadata, err := json.Marshal(m.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshaling metadata: %w", err)
	}

	// Convert Go slice to pq.StringArray - this handles the PostgreSQL array conversion properly
	tags := pq.StringArray(m.Tags)

	gormMemory := &model.Memory{
		ID:           m.ID.String(),
		UserID:       m.UserID.String(),
		Content:      m.Content,
		Summary:      stringPtr(m.Summary),
		Importance:   intPtr(int32(m.Importance)),
		MemoryType:   stringPtr(m.MemoryType),
		Tags:         tags,
		Metadata:     stringPtr(string(metadata)),
		CreatedAt:    &m.CreatedAt,
		UpdatedAt:    &m.UpdatedAt,
		LastAccessed: &m.LastAccessed,
		AccessCount:  intPtr(int32(m.AccessCount)),
	}

	return gormMemory, nil
}

func (r *postgresRepository) toDomain(gormMemory *model.Memory) (*memory.Memory, error) {
	id, err := uuid.Parse(gormMemory.ID)
	if err != nil {
		return nil, fmt.Errorf("parsing memory ID: %w", err)
	}

	userID, err := uuid.Parse(gormMemory.UserID)
	if err != nil {
		return nil, fmt.Errorf("parsing user ID: %w", err)
	}

	m := &memory.Memory{
		ID:          memory.ID(id),
		UserID:      user.ID(userID),
		Content:     gormMemory.Content,
		Tags:        make([]string, 0),
		Metadata:    make(map[string]interface{}),
		AccessCount: 0,
	}

	if gormMemory.Summary != nil {
		m.Summary = *gormMemory.Summary
	}

	if gormMemory.Importance != nil {
		m.Importance = int(*gormMemory.Importance)
	} else {
		m.Importance = 5 // default
	}

	if gormMemory.MemoryType != nil {
		m.MemoryType = *gormMemory.MemoryType
	} else {
		m.MemoryType = "general" // default
	}

	// Convert pq.StringArray to Go slice
	m.Tags = []string(gormMemory.Tags)

	if gormMemory.Metadata != nil && *gormMemory.Metadata != "" && *gormMemory.Metadata != "{}" {
		if err := json.Unmarshal([]byte(*gormMemory.Metadata), &m.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshaling metadata: %w", err)
		}
	}

	if gormMemory.CreatedAt != nil {
		m.CreatedAt = *gormMemory.CreatedAt
	}

	if gormMemory.UpdatedAt != nil {
		m.UpdatedAt = *gormMemory.UpdatedAt
	}

	if gormMemory.LastAccessed != nil {
		m.LastAccessed = *gormMemory.LastAccessed
	}

	if gormMemory.AccessCount != nil {
		m.AccessCount = int(*gormMemory.AccessCount)
	}

	return m, nil
}