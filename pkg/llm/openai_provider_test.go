package llm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIProvider_New(t *testing.T) {
	config := &Config{
		APIKey:          "test-key",
		EmbeddingModel:  "text-embedding-ada-002",
		CompletionModel: "gpt-3.5-turbo",
		TimeoutSeconds:  30,
		MaxRetries:      3,
	}

	provider := NewOpenAIProvider(config)
	
	assert.NotNil(t, provider)
	assert.Equal(t, "openai", provider.Name())
	assert.Equal(t, "text-embedding-ada-002", provider.GetDefaultModel())
}

func TestOpenAIProvider_GetEmbeddingDimension(t *testing.T) {
	provider := NewOpenAIProvider(&Config{})
	
	testCases := []struct {
		model    string
		expected int
	}{
		{"text-embedding-ada-002", 1536},
		{"text-embedding-3-small", 1536},
		{"text-embedding-3-large", 3072},
		{"unknown-model", 1536}, // default
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			dim := provider.GetEmbeddingDimension(tc.model)
			assert.Equal(t, tc.expected, dim)
		})
	}
}

func TestOpenAIProvider_GenerateEmbeddings_Integration(t *testing.T) {
	// Skip integration test if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	config := &Config{
		APIKey:         apiKey,
		EmbeddingModel: "text-embedding-ada-002",
		TimeoutSeconds: 30,
	}

	provider := NewOpenAIProvider(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("single_text", func(t *testing.T) {
		req := &EmbeddingRequest{
			Input: []string{"Hello, world!"},
		}

		resp, err := provider.GenerateEmbeddings(ctx, req)
		
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Embeddings, 1)
		assert.Len(t, resp.Embeddings[0], 1536) // ada-002 dimensions
		assert.NotEmpty(t, resp.Model)
		assert.Greater(t, resp.Usage.TotalTokens, 0)
	})

	t.Run("multiple_texts", func(t *testing.T) {
		req := &EmbeddingRequest{
			Input: []string{
				"Hello, world!",
				"This is a test.",
				"AI memory systems are fascinating.",
			},
		}

		resp, err := provider.GenerateEmbeddings(ctx, req)
		
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Embeddings, 3)
		
		for _, embedding := range resp.Embeddings {
			assert.Len(t, embedding, 1536)
		}
		
		assert.NotEmpty(t, resp.Model)
		assert.Greater(t, resp.Usage.TotalTokens, 0)
	})

	t.Run("empty_input", func(t *testing.T) {
		req := &EmbeddingRequest{
			Input: []string{},
		}

		resp, err := provider.GenerateEmbeddings(ctx, req)
		
		// OpenAI API should return an error for empty input
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestOpenAIProvider_GenerateCompletion_Integration(t *testing.T) {
	// Skip integration test if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	config := &Config{
		APIKey:          apiKey,
		CompletionModel: "gpt-3.5-turbo",
		TimeoutSeconds:  30,
	}

	provider := NewOpenAIProvider(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("simple_completion", func(t *testing.T) {
		req := &CompletionRequest{
			Messages: []Message{
				{Role: "user", Content: "Say hello in a friendly way."},
			},
		}

		resp, err := provider.GenerateCompletion(ctx, req)
		
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Content)
		assert.NotEmpty(t, resp.Model)
		assert.Greater(t, resp.Usage.TotalTokens, 0)
		assert.Greater(t, resp.Usage.CompletionTokens, 0)
	})

	t.Run("system_and_user_messages", func(t *testing.T) {
		req := &CompletionRequest{
			Messages: []Message{
				{Role: "system", Content: "You are a helpful assistant that responds concisely."},
				{Role: "user", Content: "What is 2+2?"},
			},
		}

		resp, err := provider.GenerateCompletion(ctx, req)
		
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Content)
		assert.Contains(t, resp.Content, "4") // Should contain the answer
	})
}

func TestOpenAIProvider_IsHealthy_Integration(t *testing.T) {
	// Skip integration test if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	config := &Config{
		APIKey:         apiKey,
		EmbeddingModel: "text-embedding-ada-002",
		TimeoutSeconds: 10,
	}

	provider := NewOpenAIProvider(config)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := provider.IsHealthy(ctx)
	assert.NoError(t, err)
}

func TestOpenAIProvider_IsHealthy_InvalidKey(t *testing.T) {
	config := &Config{
		APIKey:         "invalid-key",
		EmbeddingModel: "text-embedding-ada-002",
		TimeoutSeconds: 5,
	}

	provider := NewOpenAIProvider(config)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := provider.IsHealthy(ctx)
	assert.Error(t, err)
}

func BenchmarkOpenAIProvider_GenerateEmbeddings(b *testing.B) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		b.Skip("OPENAI_API_KEY not set, skipping benchmark")
	}

	config := &Config{
		APIKey:         apiKey,
		EmbeddingModel: "text-embedding-ada-002",
		TimeoutSeconds: 30,
	}

	provider := NewOpenAIProvider(config)
	ctx := context.Background()

	req := &EmbeddingRequest{
		Input: []string{"This is a test sentence for benchmarking embedding generation."},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GenerateEmbeddings(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}