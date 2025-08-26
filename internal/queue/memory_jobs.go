package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
	"mem_bank/internal/service/embedding"
	"mem_bank/pkg/logger"
)

// Job types for memory processing
const (
	JobTypeGenerateEmbedding = "generate_embedding"
	JobTypeUpdateMemory      = "update_memory"
	JobTypeBatchEmbedding    = "batch_embedding"
)

// GenerateEmbeddingHandler handles embedding generation jobs
type GenerateEmbeddingHandler struct {
	embeddingService *embedding.Service
	memoryRepo       memory.Repository
	logger           logger.Logger
}

// NewGenerateEmbeddingHandler creates a new embedding generation handler
func NewGenerateEmbeddingHandler(embeddingService *embedding.Service, memoryRepo memory.Repository, logger logger.Logger) *GenerateEmbeddingHandler {
	return &GenerateEmbeddingHandler{
		embeddingService: embeddingService,
		memoryRepo:       memoryRepo,
		logger:           logger,
	}
}

// Handle processes an embedding generation job
func (h *GenerateEmbeddingHandler) Handle(ctx context.Context, job *Job) (*JobResult, error) {
	// Extract memory ID from payload
	memoryIDStr, ok := job.Payload["memory_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid memory_id in job payload")
	}

	memoryID, err := parseMemoryID(memoryIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid memory ID: %w", err)
	}

	// Retrieve memory from repository
	mem, err := h.memoryRepo.FindByID(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("retrieving memory: %w", err)
	}

	// Generate embedding for memory content
	embeddingResult, err := h.embeddingService.GenerateEmbedding(ctx, mem.Content)
	if err != nil {
		return nil, fmt.Errorf("generating embedding: %w", err)
	}

	// Update memory with embedding
	mem.UpdateEmbedding(embeddingResult.Embedding)

	// Save updated memory
	if err := h.memoryRepo.Update(ctx, mem); err != nil {
		return nil, fmt.Errorf("updating memory with embedding: %w", err)
	}

	h.logger.WithFields(map[string]interface{}{
		"memory_id":    mem.ID.String(),
		"embedding_dim": len(embeddingResult.Embedding),
		"model":        embeddingResult.Model,
		"cached":       embeddingResult.Cached,
	}).Info("Memory embedding generated and updated")

	return &JobResult{
		Result: map[string]interface{}{
			"memory_id":      mem.ID.String(),
			"embedding_dim":  len(embeddingResult.Embedding),
			"model":          embeddingResult.Model,
			"cached":         embeddingResult.Cached,
			"token_usage":    embeddingResult,
		},
	}, nil
}

// Name returns the handler name
func (h *GenerateEmbeddingHandler) Name() string {
	return "GenerateEmbeddingHandler"
}

// JobType returns the job type this handler processes
func (h *GenerateEmbeddingHandler) JobType() string {
	return JobTypeGenerateEmbedding
}

// BatchEmbeddingHandler handles batch embedding generation jobs
type BatchEmbeddingHandler struct {
	embeddingService *embedding.Service
	memoryRepo       memory.Repository
	logger           logger.Logger
}

// NewBatchEmbeddingHandler creates a new batch embedding handler
func NewBatchEmbeddingHandler(embeddingService *embedding.Service, memoryRepo memory.Repository, logger logger.Logger) *BatchEmbeddingHandler {
	return &BatchEmbeddingHandler{
		embeddingService: embeddingService,
		memoryRepo:       memoryRepo,
		logger:           logger,
	}
}

// Handle processes a batch embedding generation job
func (h *BatchEmbeddingHandler) Handle(ctx context.Context, job *Job) (*JobResult, error) {
	// Extract user ID and limit from payload
	userIDStr, ok := job.Payload["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid user_id in job payload")
	}

	userID, err := parseUserID(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	limit := 100 // Default batch size
	if limitFloat, ok := job.Payload["limit"].(float64); ok {
		limit = int(limitFloat)
	}

	// Find memories without embeddings
	memories, err := h.findMemoriesWithoutEmbeddings(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("finding memories without embeddings: %w", err)
	}

	if len(memories) == 0 {
		return &JobResult{
			Result: map[string]interface{}{
				"processed_count": 0,
				"message":        "No memories found without embeddings",
			},
		}, nil
	}

	// Extract content for batch embedding generation
	texts := make([]string, len(memories))
	for i, mem := range memories {
		texts[i] = mem.Content
	}

	// Generate embeddings in batch
	batchResult, err := h.embeddingService.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("generating batch embeddings: %w", err)
	}

	// Update memories with embeddings
	updatedCount := 0
	for i, mem := range memories {
		if i < len(batchResult.Results) {
			mem.UpdateEmbedding(batchResult.Results[i].Embedding)
			if err := h.memoryRepo.Update(ctx, mem); err != nil {
				h.logger.WithError(err).WithField("memory_id", mem.ID.String()).Error("Failed to update memory with embedding")
				continue
			}
			updatedCount++
		}
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id":        userID.String(),
		"processed_count": updatedCount,
		"total_usage":    batchResult.Usage,
	}).Info("Batch embeddings generated and updated")

	return &JobResult{
		Result: map[string]interface{}{
			"user_id":         userID.String(),
			"processed_count": updatedCount,
			"total_memories":  len(memories),
			"token_usage":     batchResult.Usage,
		},
	}, nil
}

// Name returns the handler name
func (h *BatchEmbeddingHandler) Name() string {
	return "BatchEmbeddingHandler"
}

// JobType returns the job type this handler processes
func (h *BatchEmbeddingHandler) JobType() string {
	return JobTypeBatchEmbedding
}

// findMemoriesWithoutEmbeddings finds memories that don't have embeddings yet
func (h *BatchEmbeddingHandler) findMemoriesWithoutEmbeddings(ctx context.Context, userID user.ID, limit int) ([]*memory.Memory, error) {
	// This is a simplified implementation. In practice, you might want to add
	// a specific method to the repository to efficiently find memories without embeddings
	memories, err := h.memoryRepo.FindByUserID(ctx, userID, limit*2, 0) // Get more than we need
	if err != nil {
		return nil, err
	}

	// Filter out memories that already have embeddings
	result := make([]*memory.Memory, 0, limit)
	for _, mem := range memories {
		if len(mem.Embedding) == 0 {
			result = append(result, mem)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// Helper functions
func parseMemoryID(idStr string) (memory.ID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return memory.ID{}, fmt.Errorf("parsing memory ID: %w", err)
	}
	return memory.ID(id), nil
}

func parseUserID(idStr string) (user.ID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return user.ID{}, fmt.Errorf("parsing user ID: %w", err)
	}
	return user.ID(id), nil
}

// JobFactory provides convenience methods for creating common job types
type JobFactory struct{}

// NewJobFactory creates a new job factory
func NewJobFactory() *JobFactory {
	return &JobFactory{}
}

// CreateGenerateEmbeddingJob creates a job for generating embedding for a single memory
func (f *JobFactory) CreateGenerateEmbeddingJob(memoryID memory.ID, priority int) *Job {
	return &Job{
		Type:     JobTypeGenerateEmbedding,
		Priority: priority,
		Payload: map[string]interface{}{
			"memory_id": memoryID.String(),
		},
		CreatedAt: time.Now(),
	}
}

// CreateBatchEmbeddingJob creates a job for generating embeddings for multiple memories
func (f *JobFactory) CreateBatchEmbeddingJob(userID user.ID, limit int, priority int) *Job {
	return &Job{
		Type:     JobTypeBatchEmbedding,
		Priority: priority,
		Payload: map[string]interface{}{
			"user_id": userID.String(),
			"limit":   limit,
		},
		CreatedAt: time.Now(),
	}
}