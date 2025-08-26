package embedding

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"mem_bank/pkg/llm"
	"mem_bank/pkg/logger"
)

// Service provides embedding generation with caching and preprocessing
type Service struct {
	provider llm.EmbeddingProvider
	cache    *redis.Client
	logger   logger.Logger
	config   Config
}

// Config holds embedding service configuration
type Config struct {
	// Maximum text length to process
	MaxTextLength int `mapstructure:"max_text_length"`
	
	// Whether to enable caching
	CacheEnabled bool `mapstructure:"cache_enabled"`
	
	// Cache TTL in minutes
	CacheTTLMinutes int `mapstructure:"cache_ttl_minutes"`
	
	// Batch size for processing multiple texts
	BatchSize int `mapstructure:"batch_size"`
	
	// Content preprocessing options
	PreprocessingConfig PreprocessingConfig `mapstructure:"preprocessing"`
}

// PreprocessingConfig holds text preprocessing configuration
type PreprocessingConfig struct {
	// Whether to normalize whitespace
	NormalizeWhitespace bool `mapstructure:"normalize_whitespace"`
	
	// Whether to convert to lowercase
	ToLowercase bool `mapstructure:"to_lowercase"`
	
	// Whether to remove extra punctuation
	RemoveExtraPunctuation bool `mapstructure:"remove_extra_punctuation"`
	
	// Maximum chunk size for long texts
	ChunkSize int `mapstructure:"chunk_size"`
	
	// Overlap between chunks
	ChunkOverlap int `mapstructure:"chunk_overlap"`
}

// EmbeddingResult represents the result of embedding generation
type EmbeddingResult struct {
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding"`
	Model     string    `json:"model"`
	Cached    bool      `json:"cached"`
}

// BatchEmbeddingResult represents results from batch processing
type BatchEmbeddingResult struct {
	Results []EmbeddingResult `json:"results"`
	Usage   llm.Usage         `json:"usage"`
}

// NewService creates a new embedding service
func NewService(provider llm.EmbeddingProvider, cache *redis.Client, logger logger.Logger, config Config) *Service {
	// Set defaults
	if config.MaxTextLength == 0 {
		config.MaxTextLength = 8000
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.CacheTTLMinutes == 0 {
		config.CacheTTLMinutes = 60 * 24 // 24 hours
	}
	if config.PreprocessingConfig.ChunkSize == 0 {
		config.PreprocessingConfig.ChunkSize = 4000
	}
	if config.PreprocessingConfig.ChunkOverlap == 0 {
		config.PreprocessingConfig.ChunkOverlap = 200
	}

	return &Service{
		provider: provider,
		cache:    cache,
		logger:   logger,
		config:   config,
	}
}

// GenerateEmbedding generates an embedding for a single text
func (s *Service) GenerateEmbedding(ctx context.Context, text string) (*EmbeddingResult, error) {
	results, err := s.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	
	if len(results.Results) == 0 {
		return nil, fmt.Errorf("no embedding generated")
	}
	
	return &results.Results[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (s *Service) GenerateEmbeddings(ctx context.Context, texts []string) (*BatchEmbeddingResult, error) {
	if len(texts) == 0 {
		return &BatchEmbeddingResult{Results: []EmbeddingResult{}}, nil
	}

	// Preprocess texts
	processedTexts := make([]string, len(texts))
	for i, text := range texts {
		processedTexts[i] = s.preprocessText(text)
	}

	results := make([]EmbeddingResult, 0, len(processedTexts))
	var totalUsage llm.Usage

	// Check cache first if enabled
	if s.config.CacheEnabled && s.cache != nil {
		cachedResults, uncachedTexts, uncachedIndices := s.checkCache(ctx, processedTexts)
		results = append(results, cachedResults...)
		
		// Generate embeddings for uncached texts
		if len(uncachedTexts) > 0 {
			generatedResults, usage, err := s.generateUncachedEmbeddings(ctx, uncachedTexts)
			if err != nil {
				return nil, err
			}
			
			totalUsage.PromptTokens += usage.PromptTokens
			totalUsage.TotalTokens += usage.TotalTokens
			
			// Cache new results
			s.cacheResults(ctx, generatedResults)
			
			// Merge results in original order
			results = s.mergeResults(results, generatedResults, uncachedIndices)
		}
	} else {
		// Generate all embeddings without caching
		generatedResults, usage, err := s.generateUncachedEmbeddings(ctx, processedTexts)
		if err != nil {
			return nil, err
		}
		
		results = generatedResults
		totalUsage = usage
	}

	return &BatchEmbeddingResult{
		Results: results,
		Usage:   totalUsage,
	}, nil
}

// preprocessText applies preprocessing to text
func (s *Service) preprocessText(text string) string {
	if text == "" {
		return text
	}

	config := s.config.PreprocessingConfig

	// Normalize whitespace
	if config.NormalizeWhitespace {
		text = strings.TrimSpace(text)
		// Replace multiple whitespace with single space
		text = strings.Join(strings.Fields(text), " ")
	}

	// Convert to lowercase
	if config.ToLowercase {
		text = strings.ToLower(text)
	}

	// Remove extra punctuation (simplified implementation)
	if config.RemoveExtraPunctuation {
		// Remove multiple consecutive punctuation marks
		text = strings.ReplaceAll(text, "...", ".")
		text = strings.ReplaceAll(text, "!!!", "!")
		text = strings.ReplaceAll(text, "???", "?")
	}

	// Truncate if too long
	if len(text) > s.config.MaxTextLength {
		text = text[:s.config.MaxTextLength]
	}

	return text
}

// checkCache checks for cached embeddings and returns cached results and uncached texts
func (s *Service) checkCache(ctx context.Context, texts []string) ([]EmbeddingResult, []string, []int) {
	cachedResults := []EmbeddingResult{}
	uncachedTexts := []string{}
	uncachedIndices := []int{}

	for i, text := range texts {
		cacheKey := s.getCacheKey(text)
		
		cachedData, err := s.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			var result EmbeddingResult
			if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
				result.Cached = true
				cachedResults = append(cachedResults, result)
				continue
			}
		}
		
		// Not cached or error unmarshalling
		uncachedTexts = append(uncachedTexts, text)
		uncachedIndices = append(uncachedIndices, i)
	}

	return cachedResults, uncachedTexts, uncachedIndices
}

// generateUncachedEmbeddings generates embeddings for texts not in cache
func (s *Service) generateUncachedEmbeddings(ctx context.Context, texts []string) ([]EmbeddingResult, llm.Usage, error) {
	results := []EmbeddingResult{}
	var totalUsage llm.Usage

	// Process in batches
	for i := 0; i < len(texts); i += s.config.BatchSize {
		end := i + s.config.BatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		
		req := &llm.EmbeddingRequest{
			Input: batch,
			Model: s.provider.GetDefaultModel(),
		}

		resp, err := s.provider.GenerateEmbeddings(ctx, req)
		if err != nil {
			return nil, totalUsage, fmt.Errorf("generating embeddings: %w", err)
		}

		// Create results for this batch
		for j, embedding := range resp.Embeddings {
			results = append(results, EmbeddingResult{
				Text:      batch[j],
				Embedding: embedding,
				Model:     resp.Model,
				Cached:    false,
			})
		}

		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens
	}

	return results, totalUsage, nil
}

// cacheResults stores embedding results in cache
func (s *Service) cacheResults(ctx context.Context, results []EmbeddingResult) {
	if !s.config.CacheEnabled || s.cache == nil {
		return
	}

	ttl := time.Duration(s.config.CacheTTLMinutes) * time.Minute

	for _, result := range results {
		cacheKey := s.getCacheKey(result.Text)
		
		data, err := json.Marshal(result)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to marshal embedding result for cache")
			continue
		}

		if err := s.cache.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
			s.logger.WithError(err).Warn("Failed to cache embedding result")
		}
	}
}

// mergeResults merges cached and generated results in original order
func (s *Service) mergeResults(cachedResults []EmbeddingResult, generatedResults []EmbeddingResult, uncachedIndices []int) []EmbeddingResult {
	totalCount := len(cachedResults) + len(generatedResults)
	results := make([]EmbeddingResult, totalCount)

	cachedIdx := 0
	generatedIdx := 0
	
	// Fill in the results array in the original order
	for i := 0; i < totalCount; i++ {
		// Check if this index was uncached
		isUncached := false
		for _, uncachedIdx := range uncachedIndices {
			if uncachedIdx == i {
				isUncached = true
				break
			}
		}
		
		if isUncached {
			results[i] = generatedResults[generatedIdx]
			generatedIdx++
		} else {
			results[i] = cachedResults[cachedIdx]
			cachedIdx++
		}
	}

	return results
}

// getCacheKey generates a cache key for a text
func (s *Service) getCacheKey(text string) string {
	model := s.provider.GetDefaultModel()
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s", model, text)))
	return fmt.Sprintf("embedding:%x", hash)
}

// GetCacheStats returns cache statistics
func (s *Service) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	if !s.config.CacheEnabled || s.cache == nil {
		return map[string]interface{}{
			"cache_enabled": false,
		}, nil
	}

	info := s.cache.Info(ctx, "memory")
	result, err := info.Result()
	if err != nil {
		return nil, fmt.Errorf("getting cache stats: %w", err)
	}

	return map[string]interface{}{
		"cache_enabled": true,
		"cache_info":    result,
	}, nil
}

// ClearCache clears all embedding cache entries
func (s *Service) ClearCache(ctx context.Context) error {
	if !s.config.CacheEnabled || s.cache == nil {
		return nil
	}

	keys := s.cache.Keys(ctx, "embedding:*")
	if keys.Err() != nil {
		return fmt.Errorf("getting cache keys: %w", keys.Err())
	}

	if len(keys.Val()) > 0 {
		return s.cache.Del(ctx, keys.Val()...).Err()
	}

	return nil
}