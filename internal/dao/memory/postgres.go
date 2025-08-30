package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
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
	if len(embedding) == 0 {
		return []*memory.Memory{}, nil
	}

	// Convert float32 slice to pgvector.Vector
	vec := pgvector.NewVector(embedding)

	// Query for similar vectors using cosine similarity
	// We use 1 - (embedding <=> ?) as similarity score (higher is more similar)
	// and filter by threshold (similarity >= threshold means 1 - cosine_distance >= threshold)
	var gormMemories []*model.Memory

	query := r.db.WithContext(ctx).
		Where("user_id = ? AND embedding IS NOT NULL", userID.String()).
		Where("1 - (embedding <=> ?) >= ?", vec, threshold).
		Order(gorm.Expr("embedding <=> ?", vec)). // Order by cosine distance (ascending = most similar first)
		Limit(limit)

	err := query.Find(&gormMemories).Error
	if err != nil {
		return nil, fmt.Errorf("searching similar memories: %w", err)
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

func (r *postgresRepository) SearchSimilarWithScores(ctx context.Context, embedding []float32, userID user.ID, limit int, threshold float64) ([]*memory.MemoryWithScore, error) {
	if len(embedding) == 0 {
		return []*memory.MemoryWithScore{}, nil
	}

	// Convert float32 slice to pgvector.Vector
	vec := pgvector.NewVector(embedding)

	// Query for similar vectors with scores
	var results []struct {
		Memory model.Memory
		Score  float64
	}

	query := r.db.WithContext(ctx).
		Select("*, 1 - (embedding <=> ?) AS score", vec).
		Where("user_id = ? AND embedding IS NOT NULL", userID.String()).
		Where("1 - (embedding <=> ?) >= ?", vec, threshold).
		Order("score DESC"). // Order by similarity score descending (most similar first)
		Limit(limit)

	err := query.Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("searching similar memories with scores: %w", err)
	}

	memoriesWithScores := make([]*memory.MemoryWithScore, 0, len(results))
	for _, result := range results {
		m, err := r.toDomain(&result.Memory)
		if err != nil {
			return nil, fmt.Errorf("converting memory: %w", err)
		}
		memoriesWithScores = append(memoriesWithScores, &memory.MemoryWithScore{
			Memory: m,
			Score:  result.Score,
		})
	}

	return memoriesWithScores, nil
}

func (r *postgresRepository) SearchSimilarByMemory(ctx context.Context, memoryID memory.ID, userID user.ID, limit int, threshold float64) ([]*memory.Memory, error) {
	// First, get the source memory's embedding
	sourceMemory, err := r.FindByID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("finding source memory: %w", err)
	}

	if len(sourceMemory.Embedding) == 0 {
		return []*memory.Memory{}, fmt.Errorf("source memory has no embedding")
	}

	// Search for similar memories, excluding the source memory itself
	vec := pgvector.NewVector(sourceMemory.Embedding)
	var gormMemories []*model.Memory

	query := r.db.WithContext(ctx).
		Where("user_id = ? AND embedding IS NOT NULL AND id != ?", userID.String(), memoryID.String()).
		Where("1 - (embedding <=> ?) >= ?", vec, threshold).
		Order(gorm.Expr("embedding <=> ?", vec)). // Order by cosine distance (ascending = most similar first)
		Limit(limit)

	err = query.Find(&gormMemories).Error
	if err != nil {
		return nil, fmt.Errorf("searching similar memories by memory: %w", err)
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

	// Convert embedding to pgvector format if present
	if len(m.Embedding) > 0 {
		gormMemory.Embedding = pgvector.NewVector(m.Embedding)
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

	// Convert pgvector.Vector to []float32 if present
	vectorSlice := gormMemory.Embedding.Slice()
	if len(vectorSlice) > 0 {
		m.Embedding = make([]float32, len(vectorSlice))
		for i, v := range vectorSlice {
			m.Embedding[i] = float32(v)
		}
	}

	return m, nil
}

// BatchStore creates multiple memories in a single transaction for better performance
func (r *postgresRepository) BatchStore(ctx context.Context, memories []*memory.Memory) error {
	if len(memories) == 0 {
		return nil
	}

	models := make([]*model.Memory, 0, len(memories))
	for _, m := range memories {
		model, err := r.toModel(m)
		if err != nil {
			return fmt.Errorf("converting memory to model: %w", err)
		}
		models = append(models, model)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		batchSize := 100 // Process in batches to avoid memory issues
		for i := 0; i < len(models); i += batchSize {
			end := i + batchSize
			if end > len(models) {
				end = len(models)
			}

			if err := tx.Create(models[i:end]).Error; err != nil {
				return fmt.Errorf("batch creating memories (batch %d): %w", i/batchSize+1, err)
			}
		}
		return nil
	})
}

// BatchUpdate updates multiple memories in a single transaction for better performance
func (r *postgresRepository) BatchUpdate(ctx context.Context, memories []*memory.Memory) error {
	if len(memories) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, m := range memories {
			model, err := r.toModel(m)
			if err != nil {
				return fmt.Errorf("converting memory to model: %w", err)
			}

			result := tx.Where("id = ?", m.ID.String()).Updates(model)
			if result.Error != nil {
				return fmt.Errorf("updating memory %s: %w", m.ID.String(), result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("memory %s not found for update", m.ID.String())
			}
		}
		return nil
	})
}

// BatchDelete removes multiple memories by their IDs in a single transaction
func (r *postgresRepository) BatchDelete(ctx context.Context, ids []memory.ID) error {
	if len(ids) == 0 {
		return nil
	}

	stringIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		stringIDs = append(stringIDs, id.String())
	}

	result := r.db.WithContext(ctx).Where("id IN ?", stringIDs).Delete(&model.Memory{})
	if result.Error != nil {
		return fmt.Errorf("batch deleting memories: %w", result.Error)
	}

	if int(result.RowsAffected) != len(ids) {
		return fmt.Errorf("expected to delete %d memories, but deleted %d", len(ids), result.RowsAffected)
	}

	return nil
}

// BatchUpdateEmbeddings updates embeddings for multiple memories efficiently
func (r *postgresRepository) BatchUpdateEmbeddings(ctx context.Context, updates []memory.EmbeddingUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			// Convert embedding to pgvector format
			var embedding pgvector.Vector
			if len(update.Embedding) > 0 {
				embedding = pgvector.NewVector(update.Embedding)
			}

			result := tx.Model(&model.Memory{}).
				Where("id = ?", update.ID.String()).
				Updates(map[string]interface{}{
					"embedding":  embedding,
					"updated_at": time.Now(),
				})

			if result.Error != nil {
				return fmt.Errorf("updating embedding for memory %s: %w", update.ID.String(), result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("memory %s not found for embedding update", update.ID.String())
			}
		}
		return nil
	})
}
