package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mem_bank/configs"
	memoryDao "mem_bank/internal/dao/memory"
	userDao "mem_bank/internal/dao/user"
	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
	"mem_bank/internal/queue"
	embeddingService "mem_bank/internal/service/embedding"
	memoryService "mem_bank/internal/service/memory"
	"mem_bank/pkg/database"
	"mem_bank/pkg/llm"
	"mem_bank/pkg/logger"
)

func TestAIMemoryIntegration(t *testing.T) {
	// Skip integration test if OpenAI API key is not provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping AI integration test")
	}

	// Setup test database
	dbConfig := &configs.DatabaseConfig{
		Host:         getEnv("TEST_DB_HOST", "localhost"),
		Port:         5432,
		User:         getEnv("TEST_DB_USER", "test_user"),
		Password:     getEnv("TEST_DB_PASSWORD", "test_password"),
		DBName:       getEnv("TEST_DB_NAME", "test_mem_bank"),
		SSLMode:      "disable",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
		MaxLifetime:  5 * time.Minute,
	}

	db, err := database.NewGormConnection(dbConfig)
	if err != nil {
		t.Skip("Test database not available:", err)
	}

	// Setup Redis
	redisClient, err := database.NewRedisClientWithOptions(&redis.Options{
		Addr: getEnv("TEST_REDIS_ADDR", "localhost:6379"),
		DB:   1, // Use different DB for testing
	}, time.Second)
	if err != nil {
		t.Skip("Redis not available, skipping integration test:", err)
	}

	// Clean up test data
	defer redisClient.FlushDB(context.Background())

	// Setup logger
	appLogger, err := logger.NewLogger(&configs.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})
	require.NoError(t, err)

	// Setup LLM provider
	llmConfig := &llm.Config{
		APIKey:         apiKey,
		EmbeddingModel: "text-embedding-ada-002",
		TimeoutSeconds: 30,
	}
	llmProvider := llm.NewOpenAIProvider(llmConfig)

	// Setup embedding service
	embeddingConfig := embeddingService.Config{
		MaxTextLength:   5000,
		CacheEnabled:    true,
		CacheTTLMinutes: 60,
		BatchSize:       10,
		PreprocessingConfig: embeddingService.PreprocessingConfig{
			NormalizeWhitespace: true,
			ToLowercase:         false,
		},
	}
	embedSvc := embeddingService.NewService(llmProvider, redisClient, appLogger, embeddingConfig)

	// Setup queue (simplified without actual processing for this test)
	queueConfig := queue.Config{
		QueueName:  "test_queue",
		MaxRetries: 3,
	}
	jobQueue := queue.NewRedisQueue(redisClient, appLogger, queueConfig)

	// Setup repositories
	memoryRepo := memoryDao.NewPostgresRepository(db.DB)
	userRepo := userDao.NewPostgresRepository(db.DB)

	// Setup AI-enhanced memory service
	aiConfig := memoryService.AIServiceConfig{
		AsyncEmbedding:             false, // Sync for testing
		DefaultSimilarityThreshold: 0.8,
		AutoGenerateEmbeddings:     true,
	}
	aiMemoryService := memoryService.NewAIService(memoryRepo, userRepo, embedSvc, jobQueue, appLogger, aiConfig)

	// Run integration tests
	t.Run("end_to_end_memory_with_ai", func(t *testing.T) {
		// Create test user
		testUser := &user.User{
			ID:       user.NewID(),
			Username: "test_ai_user",
			Email:    "test@example.com",
			Profile:  user.Profile{FirstName: "Test", LastName: "User"},
			IsActive: true,
		}

		err := userRepo.Store(context.Background(), testUser)
		require.NoError(t, err)

		// Clean up user after test
		defer userRepo.Delete(context.Background(), testUser.ID)

		// Create memories with AI processing
		memoryContents := []string{
			"I love playing basketball and watching NBA games",
			"Basketball is my favorite sport to play and watch",
			"I enjoy cooking Italian pasta dishes",
			"Machine learning and AI fascinate me",
			"Deep learning models are powerful tools",
		}

		createdMemories := make([]*memory.Memory, 0, len(memoryContents))

		// Create memories (should auto-generate embeddings)
		for i, content := range memoryContents {
			req := memory.CreateRequest{
				UserID:     testUser.ID,
				Content:    content,
				Summary:    fmt.Sprintf("Test memory %d", i+1),
				Importance: 5,
				MemoryType: "test",
				Tags:       []string{"test", "integration"},
			}

			mem, err := aiMemoryService.CreateMemory(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, mem)

			createdMemories = append(createdMemories, mem)

			// Verify embedding was generated
			assert.NotEmpty(t, mem.Embedding, "Embedding should be generated automatically")
			assert.Equal(t, 1536, len(mem.Embedding), "Should use OpenAI ada-002 embedding dimensions")
		}

		// Clean up memories after test
		defer func() {
			for _, mem := range createdMemories {
				memoryRepo.Delete(context.Background(), mem.ID)
			}
		}()

		// Test semantic similarity search
		t.Run("semantic_similarity_search", func(t *testing.T) {
			// Search for basketball-related content
			results, err := aiMemoryService.SearchSimilarMemories(
				context.Background(),
				"sports and basketball games",
				testUser.ID,
				3,
				0.7,
			)

			require.NoError(t, err)
			assert.True(t, len(results) >= 1, "Should find basketball-related memories")

			// Verify results contain basketball-related content
			foundBasketball := false
			for _, result := range results {
				if strings.Contains(strings.ToLower(result.Content), "basketball") {
					foundBasketball = true
					break
				}
			}
			assert.True(t, foundBasketball, "Should find basketball-related memories")
		})

		t.Run("hybrid_search_with_semantic_ranking", func(t *testing.T) {
			searchReq := memory.SearchRequest{
				UserID: testUser.ID,
				Query:  "artificial intelligence",
				Limit:  5,
			}

			results, err := aiMemoryService.SearchWithSemanticRanking(
				context.Background(),
				searchReq,
				0.6, // Give more weight to semantic similarity
			)

			require.NoError(t, err)

			// Should find AI/ML related content
			foundAI := false
			for _, result := range results {
				content := strings.ToLower(result.Content)
				if strings.Contains(content, "machine learning") ||
					strings.Contains(content, "deep learning") ||
					strings.Contains(content, "ai") {
					foundAI = true
					break
				}
			}
			assert.True(t, foundAI, "Should find AI/ML related memories")
		})

		t.Run("embedding_stats", func(t *testing.T) {
			stats, err := aiMemoryService.GetEmbeddingStats(context.Background(), testUser.ID)

			require.NoError(t, err)
			assert.Equal(t, len(createdMemories), int(stats["total_memories"].(int)))
			assert.Equal(t, len(createdMemories), int(stats["memories_with_embeddings"].(int)))
			assert.Equal(t, 0, int(stats["memories_without_embeddings"].(int)))
			assert.Equal(t, 100.0, stats["embedding_coverage_percent"].(float64))
		})

		t.Run("update_memory_regenerates_embedding", func(t *testing.T) {
			if len(createdMemories) == 0 {
				t.Skip("No memories created")
			}

			originalMemory := createdMemories[0]
			originalEmbedding := make([]float32, len(originalMemory.Embedding))
			copy(originalEmbedding, originalMemory.Embedding)

			// Update memory content
			newContent := "Updated content about quantum computing and physics"
			updateReq := memory.UpdateRequest{
				Content: &newContent,
			}

			updatedMemory, err := aiMemoryService.UpdateMemory(
				context.Background(),
				originalMemory.ID,
				updateReq,
			)

			require.NoError(t, err)
			assert.Equal(t, newContent, updatedMemory.Content)

			// Embedding should be different (regenerated)
			embeddingsEqual := true
			if len(originalEmbedding) == len(updatedMemory.Embedding) {
				for i := range originalEmbedding {
					if abs(originalEmbedding[i]-updatedMemory.Embedding[i]) > 0.001 {
						embeddingsEqual = false
						break
					}
				}
			} else {
				embeddingsEqual = false
			}

			assert.False(t, embeddingsEqual, "Embedding should be regenerated when content changes")
		})
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
