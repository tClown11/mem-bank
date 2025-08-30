package embedding

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"mem_bank/configs"
	"mem_bank/pkg/database"
	"mem_bank/pkg/llm"
	"mem_bank/pkg/logger"
)

// MockEmbeddingProvider implements the llm.EmbeddingProvider interface for testing
type MockEmbeddingProvider struct {
	mock.Mock
}

func (m *MockEmbeddingProvider) GenerateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*llm.EmbeddingResponse), args.Error(1)
}

func (m *MockEmbeddingProvider) GetEmbeddingDimension(model string) int {
	args := m.Called(model)
	return args.Int(0)
}

func (m *MockEmbeddingProvider) GetDefaultModel() string {
	args := m.Called()
	return args.String(0)
}

func TestNewService(t *testing.T) {
	provider := &MockEmbeddingProvider{}
	logger, _ := logger.NewLogger(&configs.LoggingConfig{Level: "info"})

	config := Config{
		MaxTextLength: 5000,
		CacheEnabled:  true,
		BatchSize:     50,
	}

	service := NewService(provider, nil, logger, config)

	assert.NotNil(t, service)
	assert.Equal(t, 5000, service.config.MaxTextLength)
	assert.True(t, service.config.CacheEnabled)
	assert.Equal(t, 50, service.config.BatchSize)
}

func TestService_GenerateEmbedding(t *testing.T) {
	provider := &MockEmbeddingProvider{}
	logger, _ := logger.NewLogger(&configs.LoggingConfig{Level: "info"})

	config := Config{
		CacheEnabled: false, // Disable cache for simplicity
	}

	service := NewService(provider, nil, logger, config)

	t.Run("success", func(t *testing.T) {
		expectedEmbedding := []float32{0.1, 0.2, 0.3}

		provider.On("GetDefaultModel").Return("test-model")
		provider.On("GenerateEmbeddings", mock.Anything, mock.MatchedBy(func(req *llm.EmbeddingRequest) bool {
			return len(req.Input) == 1 && req.Input[0] == "test content"
		})).Return(&llm.EmbeddingResponse{
			Embeddings: [][]float32{expectedEmbedding},
			Model:      "test-model",
		}, nil)

		result, err := service.GenerateEmbedding(context.Background(), "test content")

		require.NoError(t, err)
		assert.Equal(t, "test content", result.Text)
		assert.Equal(t, expectedEmbedding, result.Embedding)
		assert.Equal(t, "test-model", result.Model)
		assert.False(t, result.Cached)

		provider.AssertExpectations(t)
	})

	t.Run("empty_text", func(t *testing.T) {
		// Setup expectation for empty text processing
		provider.On("GetDefaultModel").Return("test-model")
		provider.On("GenerateEmbeddings", mock.Anything, mock.MatchedBy(func(req *llm.EmbeddingRequest) bool {
			return len(req.Input) == 1 && req.Input[0] == ""
		})).Return(&llm.EmbeddingResponse{
			Embeddings: [][]float32{},
			Model:      "test-model",
		}, nil)

		results, err := service.GenerateEmbeddings(context.Background(), []string{""})

		require.NoError(t, err)
		assert.Empty(t, results.Results)

		provider.AssertExpectations(t)
	})
}

func TestService_GenerateEmbeddings_Batch(t *testing.T) {
	provider := &MockEmbeddingProvider{}
	logger, _ := logger.NewLogger(&configs.LoggingConfig{Level: "info"})

	config := Config{
		CacheEnabled: false,
		BatchSize:    2,
	}

	service := NewService(provider, nil, logger, config)

	texts := []string{"text1", "text2", "text3"}
	expectedEmbeddings := [][]float32{
		{0.1, 0.2},
		{0.3, 0.4},
		{0.5, 0.6},
	}

	provider.On("GetDefaultModel").Return("test-model")

	// First batch (2 items)
	provider.On("GenerateEmbeddings", mock.Anything, mock.MatchedBy(func(req *llm.EmbeddingRequest) bool {
		return len(req.Input) == 2
	})).Return(&llm.EmbeddingResponse{
		Embeddings: expectedEmbeddings[:2],
		Model:      "test-model",
		Usage:      llm.Usage{TotalTokens: 10},
	}, nil).Once()

	// Second batch (1 item)
	provider.On("GenerateEmbeddings", mock.Anything, mock.MatchedBy(func(req *llm.EmbeddingRequest) bool {
		return len(req.Input) == 1
	})).Return(&llm.EmbeddingResponse{
		Embeddings: expectedEmbeddings[2:],
		Model:      "test-model",
		Usage:      llm.Usage{TotalTokens: 5},
	}, nil).Once()

	result, err := service.GenerateEmbeddings(context.Background(), texts)

	require.NoError(t, err)
	assert.Len(t, result.Results, 3)
	assert.Equal(t, 15, result.Usage.TotalTokens) // 10 + 5

	for i, res := range result.Results {
		assert.Equal(t, texts[i], res.Text)
		assert.Equal(t, expectedEmbeddings[i], res.Embedding)
		assert.False(t, res.Cached)
	}

	provider.AssertExpectations(t)
}

func TestService_PreprocessText(t *testing.T) {
	provider := &MockEmbeddingProvider{}
	logger, _ := logger.NewLogger(&configs.LoggingConfig{Level: "info"})

	config := Config{
		MaxTextLength: 50,
		PreprocessingConfig: PreprocessingConfig{
			NormalizeWhitespace:    true,
			ToLowercase:            true,
			RemoveExtraPunctuation: true,
		},
	}

	service := NewService(provider, nil, logger, config)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normalize_whitespace",
			input:    "Hello    world\n\n\ttest",
			expected: "hello world test",
		},
		{
			name:     "remove_extra_punctuation",
			input:    "Hello!!! World??? Test...",
			expected: "hello! world? test.",
		},
		{
			name:     "truncate_long_text",
			input:    "This is a very long text that should be truncated because it exceeds the maximum length",
			expected: "this is a very long text that should be trunca",
		},
		{
			name:     "empty_text",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.preprocessText(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestService_WithCache(t *testing.T) {
	// Skip if Redis is not available
	redisClient, err := database.NewRedisClientWithOptions(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use different DB for testing
	}, time.Second)
	if err != nil {
		t.Skip("Redis not available, skipping cache tests:", err)
	}

	// Clean up test data
	defer redisClient.FlushDB(context.Background())

	provider := &MockEmbeddingProvider{}
	logger, _ := logger.NewLogger(&configs.LoggingConfig{Level: "info"})

	config := Config{
		CacheEnabled:    true,
		CacheTTLMinutes: 1,
	}

	service := NewService(provider, redisClient, logger, config)

	expectedEmbedding := []float32{0.1, 0.2, 0.3}

	provider.On("GetDefaultModel").Return("test-model")
	provider.On("GenerateEmbeddings", mock.Anything, mock.Anything).Return(&llm.EmbeddingResponse{
		Embeddings: [][]float32{expectedEmbedding},
		Model:      "test-model",
	}, nil).Once() // Should only be called once due to caching

	// First call - should hit the provider
	result1, err := service.GenerateEmbedding(context.Background(), "test content")
	require.NoError(t, err)
	assert.False(t, result1.Cached)

	// Second call - should hit the cache
	result2, err := service.GenerateEmbedding(context.Background(), "test content")
	require.NoError(t, err)
	assert.True(t, result2.Cached)
	assert.Equal(t, result1.Embedding, result2.Embedding)

	provider.AssertExpectations(t)
}

func TestService_ClearCache(t *testing.T) {
	// Skip if Redis is not available
	redisClient, err := database.NewRedisClientWithOptions(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use different DB for testing
	}, time.Second)
	if err != nil {
		t.Skip("Redis not available, skipping cache tests:", err)
	}

	// Clean up test data
	defer redisClient.FlushDB(context.Background())

	provider := &MockEmbeddingProvider{}
	logger, _ := logger.NewLogger(&configs.LoggingConfig{Level: "info"})

	config := Config{
		CacheEnabled: true,
	}

	service := NewService(provider, redisClient, logger, config)

	// Add some data to cache
	expectedEmbedding := []float32{0.1, 0.2, 0.3}
	provider.On("GetDefaultModel").Return("test-model")
	provider.On("GenerateEmbeddings", mock.Anything, mock.Anything).Return(&llm.EmbeddingResponse{
		Embeddings: [][]float32{expectedEmbedding},
		Model:      "test-model",
	}, nil)

	_, err = service.GenerateEmbedding(context.Background(), "test content")
	require.NoError(t, err)

	// Clear cache
	err = service.ClearCache(context.Background())
	require.NoError(t, err)

	// Verify cache is cleared by checking that the next call hits the provider again
	provider.On("GenerateEmbeddings", mock.Anything, mock.Anything).Return(&llm.EmbeddingResponse{
		Embeddings: [][]float32{expectedEmbedding},
		Model:      "test-model",
	}, nil).Once()

	result, err := service.GenerateEmbedding(context.Background(), "test content")
	require.NoError(t, err)
	assert.False(t, result.Cached) // Should not be cached after clearing
}

func BenchmarkService_GenerateEmbedding(b *testing.B) {
	provider := &MockEmbeddingProvider{}
	logger, _ := logger.NewLogger(&configs.LoggingConfig{Level: "error"}) // Reduce log noise

	config := Config{
		CacheEnabled: false, // Disable cache for benchmark
	}

	service := NewService(provider, nil, logger, config)

	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001
	}

	provider.On("GetDefaultModel").Return("test-model")
	provider.On("GenerateEmbeddings", mock.Anything, mock.Anything).Return(&llm.EmbeddingResponse{
		Embeddings: [][]float32{embedding},
		Model:      "test-model",
	}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateEmbedding(context.Background(), "benchmark text")
		if err != nil {
			b.Fatal(err)
		}
	}
}
