package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"mem_bank/internal/domain"
)

type MockMemoryRepository struct {
	mock.Mock
}

func (m *MockMemoryRepository) Create(memory *domain.Memory) error {
	args := m.Called(memory)
	return args.Error(0)
}

func (m *MockMemoryRepository) GetByID(id uuid.UUID) (*domain.Memory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Memory), args.Error(1)
}

func (m *MockMemoryRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]*domain.Memory), args.Error(1)
}

func (m *MockMemoryRepository) Update(memory *domain.Memory) error {
	args := m.Called(memory)
	return args.Error(0)
}

func (m *MockMemoryRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMemoryRepository) SearchSimilar(embedding []float32, userID uuid.UUID, limit int, threshold float64) ([]*domain.Memory, error) {
	args := m.Called(embedding, userID, limit, threshold)
	return args.Get(0).([]*domain.Memory), args.Error(1)
}

func (m *MockMemoryRepository) SearchByContent(query string, userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	args := m.Called(query, userID, limit, offset)
	return args.Get(0).([]*domain.Memory), args.Error(1)
}

func (m *MockMemoryRepository) GetByTags(tags []string, userID uuid.UUID, limit, offset int) ([]*domain.Memory, error) {
	args := m.Called(tags, userID, limit, offset)
	return args.Get(0).([]*domain.Memory), args.Error(1)
}

func (m *MockMemoryRepository) UpdateAccessInfo(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMemoryRepository) GetStatsByUserID(userID uuid.UUID) (*domain.MemoryStats, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MemoryStats), args.Error(1)
}