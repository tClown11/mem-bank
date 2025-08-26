package memory

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
	"mem_bank/pkg/database"
	"mem_bank/configs"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Skip integration tests if database is not available
	dbConfig := &configs.DatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("TEST_DB_USER", "test_user"),
		Password: getEnv("TEST_DB_PASSWORD", "test_password"),
		DBName:   getEnv("TEST_DB_NAME", "test_mem_bank"),
		SSLMode:  "disable",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
		MaxLifetime:  5 * time.Minute,
	}

	db, err := database.NewGormConnection(dbConfig)
	if err != nil {
		t.Skip("Test database not available:", err)
	}

	return db.DB
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestPostgresRepository_SearchSimilar_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresRepository(db)

	// Create test user
	testUserID := user.ID(uuid.New())
	
	// Create test memories with embeddings
	memories := []*memory.Memory{
		{
			ID:        memory.NewID(),
			UserID:    testUserID,
			Content:   "I love playing basketball",
			Summary:   "Sports preference",
			Embedding: []float32{0.1, 0.2, 0.3, 0.4},
			Importance: 5,
			MemoryType: "preference",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        memory.NewID(),
			UserID:    testUserID,
			Content:   "Basketball is my favorite sport",
			Summary:   "Sports favorite",
			Embedding: []float32{0.15, 0.22, 0.28, 0.41}, // Similar to first
			Importance: 6,
			MemoryType: "preference",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        memory.NewID(),
			UserID:    testUserID,
			Content:   "I enjoy cooking pasta",
			Summary:   "Cooking interest",
			Embedding: []float32{0.8, 0.1, 0.9, 0.2}, // Different from sports
			Importance: 4,
			MemoryType: "interest",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        memory.NewID(),
			UserID:    user.ID(uuid.New()), // Different user
			Content:   "I also love basketball",
			Summary:   "Sports preference",
			Embedding: []float32{0.12, 0.21, 0.29, 0.42}, // Similar to first but different user
			Importance: 5,
			MemoryType: "preference",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Store test memories
	ctx := context.Background()
	for _, mem := range memories {
		err := repo.Store(ctx, mem)
		require.NoError(t, err)
	}

	// Clean up after test
	defer func() {
		for _, mem := range memories {
			repo.Delete(ctx, mem.ID)
		}
	}()

	t.Run("find_similar_memories", func(t *testing.T) {
		// Search with embedding similar to basketball memories
		queryEmbedding := []float32{0.12, 0.21, 0.31, 0.39}
		
		results, err := repo.SearchSimilar(ctx, queryEmbedding, testUserID, 5, 0.5)
		
		require.NoError(t, err)
		
		// Should find the basketball-related memories for this user
		assert.True(t, len(results) >= 1, "Should find at least one similar memory")
		
		// Verify results are for the correct user
		for _, result := range results {
			assert.Equal(t, testUserID, result.UserID)
		}
		
		// First result should be most similar (basketball related)
		if len(results) > 0 {
			assert.Contains(t, results[0].Content, "basketball")
		}
	})

	t.Run("high_threshold_returns_fewer_results", func(t *testing.T) {
		queryEmbedding := []float32{0.12, 0.21, 0.31, 0.39}
		
		// High threshold should return fewer results
		resultsHighThreshold, err := repo.SearchSimilar(ctx, queryEmbedding, testUserID, 5, 0.95)
		require.NoError(t, err)
		
		// Low threshold should return more results
		resultsLowThreshold, err := repo.SearchSimilar(ctx, queryEmbedding, testUserID, 5, 0.3)
		require.NoError(t, err)
		
		assert.True(t, len(resultsHighThreshold) <= len(resultsLowThreshold),
			"High threshold should return fewer or equal results")
	})

	t.Run("empty_embedding_returns_empty_results", func(t *testing.T) {
		results, err := repo.SearchSimilar(ctx, []float32{}, testUserID, 5, 0.8)
		
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("nonexistent_user_returns_empty_results", func(t *testing.T) {
		nonexistentUser := user.ID(uuid.New())
		queryEmbedding := []float32{0.1, 0.2, 0.3, 0.4}
		
		results, err := repo.SearchSimilar(ctx, queryEmbedding, nonexistentUser, 5, 0.8)
		
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("limit_is_respected", func(t *testing.T) {
		queryEmbedding := []float32{0.12, 0.21, 0.31, 0.39}
		limit := 1
		
		results, err := repo.SearchSimilar(ctx, queryEmbedding, testUserID, limit, 0.3)
		
		require.NoError(t, err)
		assert.True(t, len(results) <= limit, "Should not exceed the specified limit")
	})
}

func TestPostgresRepository_VectorOperations_Integration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresRepository(db)

	testUserID := user.ID(uuid.New())
	ctx := context.Background()

	t.Run("store_and_retrieve_with_embedding", func(t *testing.T) {
		embedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
		
		mem := &memory.Memory{
			ID:        memory.NewID(),
			UserID:    testUserID,
			Content:   "Test content with embedding",
			Summary:   "Test summary",
			Embedding: embedding,
			Importance: 5,
			MemoryType: "test",
			Tags:      []string{"test", "embedding"},
			Metadata:  map[string]interface{}{"source": "test"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Store memory
		err := repo.Store(ctx, mem)
		require.NoError(t, err)

		// Clean up after test
		defer repo.Delete(ctx, mem.ID)

		// Retrieve memory
		retrieved, err := repo.FindByID(ctx, mem.ID)
		require.NoError(t, err)

		// Verify embedding is preserved
		assert.Equal(t, len(embedding), len(retrieved.Embedding))
		for i, v := range embedding {
			assert.InDelta(t, v, retrieved.Embedding[i], 0.0001, "Embedding value should match")
		}
	})

	t.Run("update_embedding", func(t *testing.T) {
		originalEmbedding := []float32{0.1, 0.2, 0.3}
		updatedEmbedding := []float32{0.4, 0.5, 0.6}
		
		mem := &memory.Memory{
			ID:        memory.NewID(),
			UserID:    testUserID,
			Content:   "Test content for embedding update",
			Summary:   "Test summary",
			Embedding: originalEmbedding,
			Importance: 5,
			MemoryType: "test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Store memory
		err := repo.Store(ctx, mem)
		require.NoError(t, err)

		// Clean up after test
		defer repo.Delete(ctx, mem.ID)

		// Update embedding
		mem.UpdateEmbedding(updatedEmbedding)
		err = repo.Update(ctx, mem)
		require.NoError(t, err)

		// Retrieve and verify
		retrieved, err := repo.FindByID(ctx, mem.ID)
		require.NoError(t, err)

		assert.Equal(t, len(updatedEmbedding), len(retrieved.Embedding))
		for i, v := range updatedEmbedding {
			assert.InDelta(t, v, retrieved.Embedding[i], 0.0001, "Updated embedding value should match")
		}
	})

	t.Run("memory_without_embedding", func(t *testing.T) {
		mem := &memory.Memory{
			ID:        memory.NewID(),
			UserID:    testUserID,
			Content:   "Test content without embedding",
			Summary:   "Test summary",
			Embedding: nil, // No embedding
			Importance: 5,
			MemoryType: "test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Store memory
		err := repo.Store(ctx, mem)
		require.NoError(t, err)

		// Clean up after test
		defer repo.Delete(ctx, mem.ID)

		// Retrieve memory
		retrieved, err := repo.FindByID(ctx, mem.ID)
		require.NoError(t, err)

		// Verify no embedding
		assert.Empty(t, retrieved.Embedding)

		// Memory without embedding should not appear in similarity search
		queryEmbedding := []float32{0.1, 0.2, 0.3}
		results, err := repo.SearchSimilar(ctx, queryEmbedding, testUserID, 10, 0.1)
		require.NoError(t, err)

		// Should not find the memory without embedding
		found := false
		for _, result := range results {
			if result.ID == mem.ID {
				found = true
				break
			}
		}
		assert.False(t, found, "Memory without embedding should not appear in similarity search")
	})
}

func BenchmarkPostgresRepository_SearchSimilar(b *testing.B) {
	// Skip benchmark if database is not available
	dbConfig := &configs.DatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("TEST_DB_USER", "test_user"),
		Password: getEnv("TEST_DB_PASSWORD", "test_password"),
		DBName:   getEnv("TEST_DB_NAME", "test_mem_bank"),
		SSLMode:  "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}

	db, err := database.NewGormConnection(dbConfig)
	if err != nil {
		b.Skip("Test database not available:", err)
	}

	repo := NewPostgresRepository(db.DB)
	testUserID := user.ID(uuid.New())
	ctx := context.Background()

	// Create test data
	numMemories := 100
	memories := make([]*memory.Memory, numMemories)
	for i := 0; i < numMemories; i++ {
		embedding := make([]float32, 1536) // OpenAI embedding size
		for j := range embedding {
			embedding[j] = float32(i+j) * 0.001 // Generate some variation
		}

		memories[i] = &memory.Memory{
			ID:        memory.NewID(),
			UserID:    testUserID,
			Content:   fmt.Sprintf("Test memory content %d", i),
			Summary:   fmt.Sprintf("Summary %d", i),
			Embedding: embedding,
			Importance: 5,
			MemoryType: "test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repo.Store(ctx, memories[i])
	}

	// Clean up after benchmark
	defer func() {
		for _, mem := range memories {
			repo.Delete(ctx, mem.ID)
		}
	}()

	queryEmbedding := make([]float32, 1536)
	for i := range queryEmbedding {
		queryEmbedding[i] = float32(i) * 0.001
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.SearchSimilar(ctx, queryEmbedding, testUserID, 10, 0.8)
		if err != nil {
			b.Fatal(err)
		}
	}
}