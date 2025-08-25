package memory

import (
	"context"
	"fmt"
	"strings"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
)

// service implements memory.Service interface
type service struct {
	repo     memory.Repository
	userRepo user.Repository
}

// NewService creates a new memory service
func NewService(repo memory.Repository, userRepo user.Repository) memory.Service {
	return &service{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *service) CreateMemory(ctx context.Context, req memory.CreateRequest) (*memory.Memory, error) {
	// Validation
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Verify user exists
	_, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		if err == user.ErrNotFound {
			return nil, memory.ErrInvalidUserID
		}
		return nil, fmt.Errorf("verifying user: %w", err)
	}

	// Create new memory
	m := memory.NewMemory(req.UserID, req.Content, req.Summary, req.Importance, req.MemoryType)

	// Set optional fields
	if len(req.Tags) > 0 {
		m.Tags = req.Tags
	}
	if req.Metadata != nil {
		m.Metadata = req.Metadata
	}

	// TODO: Generate embedding for the content
	// For now, we'll leave it empty until we implement the embedding service

	// Store memory
	if err := s.repo.Store(ctx, m); err != nil {
		return nil, fmt.Errorf("storing memory: %w", err)
	}

	return m, nil
}

func (s *service) GetMemory(ctx context.Context, id memory.ID) (*memory.Memory, error) {
	if id.IsZero() {
		return nil, memory.ErrInvalidID
	}

	// Get memory
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update access info
	m.Access()
	if err := s.repo.UpdateAccessInfo(ctx, id); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to update access info for memory %s: %v\n", id, err)
	}

	return m, nil
}

func (s *service) UpdateMemory(ctx context.Context, id memory.ID, req memory.UpdateRequest) (*memory.Memory, error) {
	if id.IsZero() {
		return nil, memory.ErrInvalidID
	}

	// Get existing memory
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Content != nil || req.Summary != nil || req.Importance != nil {
		content := m.Content
		if req.Content != nil {
			content = *req.Content
		}

		summary := m.Summary
		if req.Summary != nil {
			summary = *req.Summary
		}

		importance := m.Importance
		if req.Importance != nil {
			importance = *req.Importance
		}

		tags := m.Tags
		if len(req.Tags) > 0 {
			tags = req.Tags
		}

		metadata := m.Metadata
		if req.Metadata != nil {
			metadata = req.Metadata
		}

		m.Update(content, summary, importance, tags, metadata)

		// TODO: Regenerate embedding if content changed
	}

	if req.MemoryType != nil {
		m.MemoryType = *req.MemoryType
	}

	// Update memory
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("updating memory: %w", err)
	}

	return m, nil
}

func (s *service) DeleteMemory(ctx context.Context, id memory.ID) error {
	if id.IsZero() {
		return memory.ErrInvalidID
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) ListUserMemories(ctx context.Context, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	if userID.IsZero() {
		return nil, memory.ErrInvalidUserID
	}

	if limit <= 0 {
		limit = 20 // default limit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindByUserID(ctx, userID, limit, offset)
}

func (s *service) SearchMemories(ctx context.Context, req memory.SearchRequest) ([]*memory.Memory, error) {
	if req.UserID.IsZero() {
		return nil, memory.ErrInvalidUserID
	}

	if req.Limit <= 0 {
		req.Limit = 20 // default limit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Search by content if query is provided
	if strings.TrimSpace(req.Query) != "" {
		return s.repo.SearchByContent(ctx, req.Query, req.UserID, req.Limit, req.Offset)
	}

	// Search by tags if provided
	if len(req.Tags) > 0 {
		return s.repo.FindByTags(ctx, req.Tags, req.UserID, req.Limit, req.Offset)
	}

	// Default to listing user memories
	return s.repo.FindByUserID(ctx, req.UserID, req.Limit, req.Offset)
}

func (s *service) SearchSimilarMemories(ctx context.Context, content string, userID user.ID, limit int, threshold float64) ([]*memory.Memory, error) {
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
		threshold = 0.8 // default threshold
	}

	// TODO: Generate embedding for content and search
	// For now, return empty results until embedding service is implemented
	return []*memory.Memory{}, nil
}

func (s *service) GetMemoryStats(ctx context.Context, userID user.ID) (*memory.Stats, error) {
	if userID.IsZero() {
		return nil, memory.ErrInvalidUserID
	}

	return s.repo.GetStatsByUserID(ctx, userID)
}

// Validation helpers
func (s *service) validateCreateRequest(req memory.CreateRequest) error {
	if req.UserID.IsZero() {
		return memory.ErrInvalidUserID
	}

	if strings.TrimSpace(req.Content) == "" {
		return memory.ErrInvalidContent
	}

	if req.Importance < 1 || req.Importance > 10 {
		return memory.ErrInvalidImportance
	}

	if strings.TrimSpace(req.MemoryType) == "" {
		return memory.ErrInvalidMemoryType
	}

	return nil
}