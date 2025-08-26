package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
)

// QdrantRepository implements memory.Repository using Qdrant vector database
type QdrantRepository struct {
	client          *qdrant.Client
	collectionName  string
	vectorSize      uint64
	postgresRepo    memory.Repository // Fallback for metadata storage
}

// QdrantConfig holds Qdrant configuration
type QdrantConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	CollectionName string `mapstructure:"collection_name"`
	VectorSize     int    `mapstructure:"vector_size"`
	UseHTTPS       bool   `mapstructure:"use_https"`
	APIKey         string `mapstructure:"api_key"`
}

// NewQdrantRepository creates a new Qdrant-based memory repository
func NewQdrantRepository(config QdrantConfig, postgresRepo memory.Repository) (*QdrantRepository, error) {
	// Default values
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 6334
	}
	if config.CollectionName == "" {
		config.CollectionName = "memories"
	}
	if config.VectorSize == 0 {
		config.VectorSize = 1536 // Default to OpenAI ada-002 dimensions
	}

	// Create Qdrant client using the high-level API
	clientConfig := &qdrant.Config{
		Host: config.Host,
		Port: config.Port,
	}
	
	if config.UseHTTPS {
		clientConfig.UseTLS = true
	}
	
	if config.APIKey != "" {
		clientConfig.APIKey = config.APIKey
	}

	client, err := qdrant.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("creating Qdrant client: %w", err)
	}
	
	repo := &QdrantRepository{
		client:          client,
		collectionName:  config.CollectionName,
		vectorSize:      uint64(config.VectorSize),
		postgresRepo:    postgresRepo,
	}

	// Initialize collection if it doesn't exist
	if err := repo.initializeCollection(context.Background()); err != nil {
		return nil, fmt.Errorf("initializing collection: %w", err)
	}

	return repo, nil
}

// Store creates a new memory in both Qdrant and PostgreSQL
func (r *QdrantRepository) Store(ctx context.Context, mem *memory.Memory) error {
	// Store in PostgreSQL first for metadata
	if err := r.postgresRepo.Store(ctx, mem); err != nil {
		return fmt.Errorf("storing in postgres: %w", err)
	}

	// Store vector in Qdrant if embedding exists
	if len(mem.Embedding) > 0 {
		if err := r.upsertVector(ctx, mem); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to store vector in Qdrant for memory %s: %v\n", mem.ID, err)
		}
	}

	return nil
}

// FindByID retrieves a memory by its ID
func (r *QdrantRepository) FindByID(ctx context.Context, id memory.ID) (*memory.Memory, error) {
	return r.postgresRepo.FindByID(ctx, id)
}

// FindByUserID retrieves memories for a specific user with pagination
func (r *QdrantRepository) FindByUserID(ctx context.Context, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	return r.postgresRepo.FindByUserID(ctx, userID, limit, offset)
}

// Update updates an existing memory in both stores
func (r *QdrantRepository) Update(ctx context.Context, mem *memory.Memory) error {
	// Update in PostgreSQL first
	if err := r.postgresRepo.Update(ctx, mem); err != nil {
		return fmt.Errorf("updating in postgres: %w", err)
	}

	// Update vector in Qdrant if embedding exists
	if len(mem.Embedding) > 0 {
		if err := r.upsertVector(ctx, mem); err != nil {
			fmt.Printf("Warning: failed to update vector in Qdrant for memory %s: %v\n", mem.ID, err)
		}
	}

	return nil
}

// Delete removes a memory by its ID from both stores
func (r *QdrantRepository) Delete(ctx context.Context, id memory.ID) error {
	// Delete from PostgreSQL
	if err := r.postgresRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting from postgres: %w", err)
	}

	// Delete vector from Qdrant
	if err := r.deleteVector(ctx, id); err != nil {
		fmt.Printf("Warning: failed to delete vector from Qdrant for memory %s: %v\n", id, err)
	}

	return nil
}

// SearchSimilar finds similar memories using Qdrant vector search
func (r *QdrantRepository) SearchSimilar(ctx context.Context, embedding []float32, userID user.ID, limit int, threshold float64) ([]*memory.Memory, error) {
	if len(embedding) == 0 {
		return []*memory.Memory{}, nil
	}

	// Search in Qdrant using the new Query API
	queryPoints := &qdrant.QueryPoints{
		CollectionName: r.collectionName,
		Query:          qdrant.NewQuery(embedding...),
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatch("user_id", userID.String()),
			},
		},
		Limit:       qdrant.PtrOf(uint64(limit)),
		WithPayload: qdrant.NewWithPayload(true),
	}

	results, err := r.client.Query(ctx, queryPoints)
	if err != nil {
		return nil, fmt.Errorf("searching in Qdrant: %w", err)
	}

	// Filter results by score threshold if provided
	var filteredResults []*qdrant.ScoredPoint
	for _, result := range results {
		if threshold > 0 && result.Score < float32(threshold) {
			continue
		}
		filteredResults = append(filteredResults, result)
	}

	// Get memory IDs from results
	memoryIDs := make([]memory.ID, 0, len(filteredResults))
	for _, result := range filteredResults {
		if payload := result.Payload; payload != nil {
			if memIDValue, ok := payload["memory_id"]; ok {
				if memIDStr, ok := memIDValue.GetKind().(*qdrant.Value_StringValue); ok {
					if memID, err := uuid.Parse(memIDStr.StringValue); err == nil {
						memoryIDs = append(memoryIDs, memory.ID(memID))
					}
				}
			}
		}
	}

	// Retrieve full memory objects from PostgreSQL
	memories := make([]*memory.Memory, 0, len(memoryIDs))
	for _, id := range memoryIDs {
		mem, err := r.postgresRepo.FindByID(ctx, id)
		if err == nil { // Skip errors for individual memories
			memories = append(memories, mem)
		}
	}

	return memories, nil
}

// SearchByContent searches memories by content text using PostgreSQL
func (r *QdrantRepository) SearchByContent(ctx context.Context, query string, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	return r.postgresRepo.SearchByContent(ctx, query, userID, limit, offset)
}

// FindByTags retrieves memories by tags using PostgreSQL
func (r *QdrantRepository) FindByTags(ctx context.Context, tags []string, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	return r.postgresRepo.FindByTags(ctx, tags, userID, limit, offset)
}

// UpdateAccessInfo updates the access information using PostgreSQL
func (r *QdrantRepository) UpdateAccessInfo(ctx context.Context, id memory.ID) error {
	return r.postgresRepo.UpdateAccessInfo(ctx, id)
}

// GetStatsByUserID returns memory statistics using PostgreSQL
func (r *QdrantRepository) GetStatsByUserID(ctx context.Context, userID user.ID) (*memory.Stats, error) {
	return r.postgresRepo.GetStatsByUserID(ctx, userID)
}

// CountByUserID returns the total number of memories for a user using PostgreSQL
func (r *QdrantRepository) CountByUserID(ctx context.Context, userID user.ID) (int, error) {
	return r.postgresRepo.CountByUserID(ctx, userID)
}

// Close closes the Qdrant connection
func (r *QdrantRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Private helper methods

// initializeCollection creates the collection if it doesn't exist
func (r *QdrantRepository) initializeCollection(ctx context.Context) error {
	// Check if collection exists
	collections, err := r.client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("listing collections: %w", err)
	}

	// Check if our collection already exists
	for _, collection := range collections {
		if collection == r.collectionName {
			return nil // Collection already exists
		}
	}

	// Create collection
	createRequest := &qdrant.CreateCollection{
		CollectionName: r.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     r.vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	}

	err = r.client.CreateCollection(ctx, createRequest)
	if err != nil {
		return fmt.Errorf("creating collection: %w", err)
	}

	// Create index for user_id for efficient filtering
	fieldType := qdrant.FieldType_FieldTypeKeyword
	indexRequest := &qdrant.CreateFieldIndexCollection{
		CollectionName: r.collectionName,
		FieldName:      "user_id",
		FieldType:      &fieldType,
	}

	_, err = r.client.CreateFieldIndex(ctx, indexRequest)
	if err != nil {
		fmt.Printf("Warning: failed to create user_id index: %v\n", err)
	}

	return nil
}

// upsertVector inserts or updates a vector in Qdrant
func (r *QdrantRepository) upsertVector(ctx context.Context, mem *memory.Memory) error {
	if len(mem.Embedding) == 0 {
		return fmt.Errorf("memory has no embedding")
	}

	// Create payload
	payload := qdrant.NewValueMap(map[string]any{
		"memory_id":   mem.ID.String(),
		"user_id":     mem.UserID.String(),
		"content":     mem.Content,
		"importance":  int64(mem.Importance),
		"memory_type": mem.MemoryType,
		"created_at":  mem.CreatedAt.Format(time.RFC3339),
	})

	// Create point
	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDUUID(mem.ID.String()),
		Vectors: qdrant.NewVectors(mem.Embedding...),
		Payload: payload,
	}

	// Upsert point
	upsertRequest := &qdrant.UpsertPoints{
		CollectionName: r.collectionName,
		Points:         []*qdrant.PointStruct{point},
	}

	_, err := r.client.Upsert(ctx, upsertRequest)
	if err != nil {
		return fmt.Errorf("upserting vector: %w", err)
	}

	return nil
}

// deleteVector removes a vector from Qdrant
func (r *QdrantRepository) deleteVector(ctx context.Context, id memory.ID) error {
	deleteRequest := &qdrant.DeletePoints{
		CollectionName: r.collectionName,
		Points:         qdrant.NewPointsSelector(qdrant.NewIDUUID(id.String())),
	}

	_, err := r.client.Delete(ctx, deleteRequest)
	if err != nil {
		return fmt.Errorf("deleting vector: %w", err)
	}

	return nil
}

// GetQdrantCollectionInfo returns information about the Qdrant collection
func (r *QdrantRepository) GetQdrantCollectionInfo(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.GetCollectionInfo(ctx, r.collectionName)
	if err != nil {
		return nil, fmt.Errorf("getting collection info: %w", err)
	}

	return map[string]interface{}{
		"collection_name":  r.collectionName,
		"vectors_count":    info.VectorsCount,
		"indexed_vectors":  info.IndexedVectorsCount,
		"points_count":     info.PointsCount,
		"status":          info.Status.String(),
	}, nil
}