package usecase

import (
	"fmt"

	"github.com/google/uuid"
	"mem_bank/internal/domain"
)

type MemoryUsecase interface {
	CreateMemory(req *domain.MemoryCreateRequest) (*domain.Memory, error)
	GetMemoryByID(id uuid.UUID) (*domain.Memory, error)
	GetMemoriesByUserID(userID uuid.UUID, limit, offset int) ([]*domain.Memory, error)
	UpdateMemory(id uuid.UUID, req *domain.MemoryUpdateRequest) (*domain.Memory, error)
	DeleteMemory(id uuid.UUID) error
	SearchSimilarMemories(embedding []float32, userID uuid.UUID, limit int, threshold float64) ([]*domain.Memory, error)
	SearchMemoriesByContent(query string, userID uuid.UUID, limit, offset int) ([]*domain.Memory, error)
	SearchMemoriesByTags(tags []string, userID uuid.UUID, limit, offset int) ([]*domain.Memory, error)
	GetMemoryStats(userID uuid.UUID) (*domain.MemoryStats, error)
}

type memoryUsecase struct {
	memoryRepo domain.MemoryRepository
	userRepo   domain.UserRepository
}

func NewMemoryUsecase(memoryRepo domain.MemoryRepository, userRepo domain.UserRepository) MemoryUsecase {
	return &memoryUsecase{
		memoryRepo: memoryRepo,
		userRepo:   userRepo,
	}
}

func (u *memoryUsecase) CreateMemory(req *domain.MemoryCreateRequest) (*domain.Memory, error) {
	if err := u.validateCreateRequest(req); err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	if req.Importance < 1 || req.Importance > 10 {
		return nil, domain.ErrInvalidImportance
	}

	memory := &domain.Memory{
		UserID:     req.UserID,
		Content:    req.Content,
		Summary:    req.Summary,
		Importance: req.Importance,
		MemoryType: req.MemoryType,
		Tags:       req.Tags,
		Metadata:   req.Metadata,
	}

	if memory.Metadata == nil {
		memory.Metadata = make(map[string]interface{})
	}

	err = u.memoryRepo.Create(memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory: %w", err)
	}

	return memory, nil
}

func (u *memoryUsecase) GetMemoryByID(id uuid.UUID) (*domain.Memory, error) {
	memory, err := u.memoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := u.memoryRepo.UpdateAccessInfo(id); err != nil {
		return nil, fmt.Errorf("failed to update access info: %w", err)
	}

	return memory, nil
}

func (u *memoryUsecase) GetMemoriesByUserID(userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	_, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	memories, err := u.memoryRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories: %w", err)
	}

	return memories, nil
}

func (u *memoryUsecase) UpdateMemory(id uuid.UUID, req *domain.MemoryUpdateRequest) (*domain.Memory, error) {
	memory, err := u.memoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Content != nil {
		if *req.Content == "" {
			return nil, domain.ErrEmptyContent
		}
		memory.Content = *req.Content
	}

	if req.Summary != nil {
		memory.Summary = *req.Summary
	}

	if req.Importance != nil {
		if *req.Importance < 1 || *req.Importance > 10 {
			return nil, domain.ErrInvalidImportance
		}
		memory.Importance = *req.Importance
	}

	if req.MemoryType != nil {
		memory.MemoryType = *req.MemoryType
	}

	if req.Tags != nil {
		memory.Tags = req.Tags
	}

	if req.Metadata != nil {
		memory.Metadata = req.Metadata
	}

	err = u.memoryRepo.Update(memory)
	if err != nil {
		return nil, fmt.Errorf("failed to update memory: %w", err)
	}

	return memory, nil
}

func (u *memoryUsecase) DeleteMemory(id uuid.UUID) error {
	_, err := u.memoryRepo.GetByID(id)
	if err != nil {
		return err
	}

	err = u.memoryRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	return nil
}

func (u *memoryUsecase) SearchSimilarMemories(embedding []float32, userID uuid.UUID, limit int, threshold float64) ([]*domain.Memory, error) {
	if len(embedding) == 0 {
		return nil, domain.ErrInvalidEmbedding
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	if threshold <= 0 || threshold > 1 {
		threshold = 0.8
	}

	_, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	memories, err := u.memoryRepo.SearchSimilar(embedding, userID, limit, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar memories: %w", err)
	}

	return memories, nil
}

func (u *memoryUsecase) SearchMemoriesByContent(query string, userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	if query == "" {
		return nil, domain.ErrInvalidInput
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	_, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	memories, err := u.memoryRepo.SearchByContent(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories by content: %w", err)
	}

	return memories, nil
}

func (u *memoryUsecase) SearchMemoriesByTags(tags []string, userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	if len(tags) == 0 {
		return nil, domain.ErrInvalidInput
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	_, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	memories, err := u.memoryRepo.GetByTags(tags, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories by tags: %w", err)
	}

	return memories, nil
}

func (u *memoryUsecase) GetMemoryStats(userID uuid.UUID) (*domain.MemoryStats, error) {
	_, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	stats, err := u.memoryRepo.GetStatsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	return stats, nil
}

func (u *memoryUsecase) validateCreateRequest(req *domain.MemoryCreateRequest) error {
	if req.UserID == uuid.Nil {
		return domain.ErrInvalidInput
	}

	if req.Content == "" {
		return domain.ErrEmptyContent
	}

	if req.MemoryType == "" {
		req.MemoryType = "general"
	}

	if req.Importance < 1 || req.Importance > 10 {
		return domain.ErrInvalidImportance
	}

	return nil
}