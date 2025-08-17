package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"mem_bank/internal/domain"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*domain.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateSettings(id uuid.UUID, settings domain.UserSettings) error {
	args := m.Called(id, settings)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateProfile(id uuid.UUID, profile domain.UserProfile) error {
	args := m.Called(id, profile)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]*domain.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) Count() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}