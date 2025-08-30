package memory

import (
	"context"
	"fmt"
	"strings"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
	"mem_bank/internal/queue"
	"mem_bank/internal/service/embedding"
	"mem_bank/pkg/logger"
)

// AIService extends the basic memory service with AI capabilities
type AIService struct {
	// Embedded basic service
	service

	// AI components
	embeddingService *embedding.Service
	jobQueue         queue.Producer
	jobFactory       *queue.JobFactory
	logger           logger.Logger

	// Configuration
	config AIServiceConfig
}

// AIServiceConfig holds configuration for the AI-enhanced memory service
type AIServiceConfig struct {
	// Whether to enable async embedding generation
	AsyncEmbedding bool `mapstructure:"async_embedding"`

	// Default similarity threshold for searches
	DefaultSimilarityThreshold float64 `mapstructure:"default_similarity_threshold"`

	// Priority for embedding generation jobs
	EmbeddingJobPriority int `mapstructure:"embedding_job_priority"`

	// Batch size for bulk embedding generation
	BatchEmbeddingSize int `mapstructure:"batch_embedding_size"`

	// Whether to auto-generate embeddings on memory creation
	AutoGenerateEmbeddings bool `mapstructure:"auto_generate_embeddings"`
}

// NewAIService creates a new AI-enhanced memory service
func NewAIService(
	repo memory.Repository,
	userRepo user.Repository,
	embeddingService *embedding.Service,
	jobQueue queue.Producer,
	logger logger.Logger,
	config AIServiceConfig,
) *AIService {
	// Set defaults
	if config.DefaultSimilarityThreshold == 0 {
		config.DefaultSimilarityThreshold = 0.8
	}
	if config.EmbeddingJobPriority == 0 {
		config.EmbeddingJobPriority = 5
	}
	if config.BatchEmbeddingSize == 0 {
		config.BatchEmbeddingSize = 100
	}

	return &AIService{
		service: service{
			repo:     repo,
			userRepo: userRepo,
		},
		embeddingService: embeddingService,
		jobQueue:         jobQueue,
		jobFactory:       queue.NewJobFactory(),
		logger:           logger,
		config:           config,
	}
}

// CreateMemory creates a new memory with AI capabilities
func (s *AIService) CreateMemory(ctx context.Context, req memory.CreateRequest) (*memory.Memory, error) {
	// Create memory using base service
	m, err := s.service.CreateMemory(ctx, req)
	if err != nil {
		return nil, err
	}

	// Generate embedding if enabled
	if s.config.AutoGenerateEmbeddings {
		if s.config.AsyncEmbedding {
			// Generate embedding asynchronously
			if err := s.scheduleEmbeddingGeneration(ctx, m.ID); err != nil {
				s.logger.WithError(err).WithField("memory_id", m.ID.String()).Warn("Failed to schedule embedding generation")
			}
		} else {
			// Generate embedding synchronously
			if err := s.generateEmbeddingSync(ctx, m); err != nil {
				s.logger.WithError(err).WithField("memory_id", m.ID.String()).Warn("Failed to generate embedding synchronously")
				// Don't fail the creation if embedding generation fails
			}
		}
	}

	return m, nil
}

// UpdateMemory updates a memory with AI capabilities
func (s *AIService) UpdateMemory(ctx context.Context, id memory.ID, req memory.UpdateRequest) (*memory.Memory, error) {
	// Check if content is being updated
	contentChanged := req.Content != nil

	// Update memory using base service
	m, err := s.service.UpdateMemory(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Regenerate embedding if content changed
	if contentChanged && s.config.AutoGenerateEmbeddings {
		if s.config.AsyncEmbedding {
			// Regenerate embedding asynchronously
			if err := s.scheduleEmbeddingGeneration(ctx, m.ID); err != nil {
				s.logger.WithError(err).WithField("memory_id", m.ID.String()).Warn("Failed to schedule embedding regeneration")
			}
		} else {
			// Regenerate embedding synchronously
			if err := s.generateEmbeddingSync(ctx, m); err != nil {
				s.logger.WithError(err).WithField("memory_id", m.ID.String()).Warn("Failed to regenerate embedding synchronously")
			}
		}
	}

	return m, nil
}

// SearchSimilarMemories searches for similar memories using vector similarity
func (s *AIService) SearchSimilarMemories(ctx context.Context, content string, userID user.ID, limit int, threshold float64) ([]*memory.Memory, error) {
	if userID.IsZero() {
		return nil, memory.ErrInvalidUserID
	}

	if strings.TrimSpace(content) == "" {
		return nil, memory.ErrInvalidContent
	}

	if limit <= 0 {
		limit = 10 // default limit
	}

	if threshold <= 0 {
		threshold = s.config.DefaultSimilarityThreshold
	}

	// Generate embedding for the search content
	embeddingResult, err := s.embeddingService.GenerateEmbedding(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("generating search embedding: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":          userID.String(),
		"search_content":   content[:min(50, len(content))],
		"threshold":        threshold,
		"embedding_cached": embeddingResult.Cached,
	}).Debug("Performing vector similarity search")

	// Search for similar memories
	memories, err := s.repo.SearchSimilar(ctx, embeddingResult.Embedding, userID, limit, threshold)
	if err != nil {
		return nil, fmt.Errorf("searching similar memories: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":       userID.String(),
		"results_count": len(memories),
	}).Info("Vector similarity search completed")

	return memories, nil
}

// GenerateEmbeddingsForUser generates embeddings for all memories of a user that don't have them
func (s *AIService) GenerateEmbeddingsForUser(ctx context.Context, userID user.ID) error {
	if userID.IsZero() {
		return memory.ErrInvalidUserID
	}

	// Schedule batch embedding generation job
	job := s.jobFactory.CreateBatchEmbeddingJob(userID, s.config.BatchEmbeddingSize, s.config.EmbeddingJobPriority)

	if err := s.jobQueue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("scheduling batch embedding generation: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID.String(),
		"job_id":  job.ID,
	}).Info("Batch embedding generation job scheduled")

	return nil
}

// GenerateEmbeddingForMemory generates an embedding for a specific memory
func (s *AIService) GenerateEmbeddingForMemory(ctx context.Context, memoryID memory.ID) error {
	if memoryID.IsZero() {
		return memory.ErrInvalidID
	}

	if s.config.AsyncEmbedding {
		return s.scheduleEmbeddingGeneration(ctx, memoryID)
	} else {
		// Get the memory
		m, err := s.repo.FindByID(ctx, memoryID)
		if err != nil {
			return fmt.Errorf("finding memory: %w", err)
		}

		return s.generateEmbeddingSync(ctx, m)
	}
}

// SearchWithSemanticRanking performs hybrid search combining text and semantic similarity
func (s *AIService) SearchWithSemanticRanking(ctx context.Context, req memory.SearchRequest, semanticWeight float64) ([]*memory.Memory, error) {
	// Validate request
	if req.UserID.IsZero() {
		return nil, memory.ErrInvalidUserID
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	// If no query, fall back to regular search
	if strings.TrimSpace(req.Query) == "" {
		return s.SearchMemories(ctx, req)
	}

	// Get text-based results
	textResults, err := s.repo.SearchByContent(ctx, req.Query, req.UserID, req.Limit*2, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("text search failed: %w", err)
	}

	// Get semantic results
	semanticResults, err := s.SearchSimilarMemories(ctx, req.Query, req.UserID, req.Limit*2, s.config.DefaultSimilarityThreshold*0.7)
	if err != nil {
		s.logger.WithError(err).Warn("Semantic search failed, falling back to text search")
		return textResults[:min(req.Limit, len(textResults))], nil
	}

	// Combine and rank results (simplified ranking algorithm)
	combined := s.combineSearchResults(textResults, semanticResults, semanticWeight)

	// Limit results
	if len(combined) > req.Limit {
		combined = combined[:req.Limit]
	}

	return combined, nil
}

// GetEmbeddingStats returns statistics about embeddings for a user
func (s *AIService) GetEmbeddingStats(ctx context.Context, userID user.ID) (map[string]interface{}, error) {
	if userID.IsZero() {
		return nil, memory.ErrInvalidUserID
	}

	// Get all user memories
	memories, err := s.repo.FindByUserID(ctx, userID, 1000, 0) // Get a large batch
	if err != nil {
		return nil, fmt.Errorf("getting user memories: %w", err)
	}

	totalMemories := len(memories)
	memoriesWithEmbeddings := 0

	for _, mem := range memories {
		if len(mem.Embedding) > 0 {
			memoriesWithEmbeddings++
		}
	}

	embeddingsCacheStats, err := s.embeddingService.GetCacheStats(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get cache stats")
		embeddingsCacheStats = map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{
		"total_memories":              totalMemories,
		"memories_with_embeddings":    memoriesWithEmbeddings,
		"memories_without_embeddings": totalMemories - memoriesWithEmbeddings,
		"embedding_coverage_percent":  float64(memoriesWithEmbeddings) / float64(totalMemories) * 100,
		"cache_stats":                 embeddingsCacheStats,
	}, nil
}

// Private helper methods

func (s *AIService) scheduleEmbeddingGeneration(ctx context.Context, memoryID memory.ID) error {
	job := s.jobFactory.CreateGenerateEmbeddingJob(memoryID, s.config.EmbeddingJobPriority)

	if err := s.jobQueue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("enqueuing embedding generation job: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"memory_id": memoryID.String(),
		"job_id":    job.ID,
	}).Debug("Embedding generation job scheduled")

	return nil
}

func (s *AIService) generateEmbeddingSync(ctx context.Context, m *memory.Memory) error {
	// Generate embedding
	embeddingResult, err := s.embeddingService.GenerateEmbedding(ctx, m.Content)
	if err != nil {
		return fmt.Errorf("generating embedding: %w", err)
	}

	// Update memory with embedding
	m.UpdateEmbedding(embeddingResult.Embedding)

	// Save updated memory
	if err := s.repo.Update(ctx, m); err != nil {
		return fmt.Errorf("updating memory with embedding: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"memory_id":        m.ID.String(),
		"embedding_dim":    len(embeddingResult.Embedding),
		"embedding_cached": embeddingResult.Cached,
	}).Debug("Memory embedding generated synchronously")

	return nil
}

func (s *AIService) combineSearchResults(textResults, semanticResults []*memory.Memory, semanticWeight float64) []*memory.Memory {
	// Create a map to avoid duplicates and combine scores
	resultMap := make(map[string]*memory.Memory)

	// Add text results
	for _, mem := range textResults {
		resultMap[mem.ID.String()] = mem
	}

	// Add semantic results (giving preference to semantic results if semanticWeight > 0.5)
	for _, mem := range semanticResults {
		resultMap[mem.ID.String()] = mem
	}

	// Convert back to slice
	combined := make([]*memory.Memory, 0, len(resultMap))
	for _, mem := range resultMap {
		combined = append(combined, mem)
	}

	return combined
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
