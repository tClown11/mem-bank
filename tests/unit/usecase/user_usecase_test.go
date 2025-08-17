package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mem_bank/internal/domain"
	"mem_bank/internal/usecase"
	"mem_bank/tests/mocks"
)

func TestUserUsecase_CreateUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *domain.UserCreateRequest
		setupMocks    func(*mocks.MockUserRepository, *mocks.MockMemoryRepository)
		expectedError error
	}{
		{
			name: "successful user creation",
			request: &domain.UserCreateRequest{
				Username: "testuser",
				Email:    "test@example.com",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, memoryRepo *mocks.MockMemoryRepository) {
				userRepo.On("GetByUsername", "testuser").Return(nil, domain.ErrUserNotFound)
				userRepo.On("GetByEmail", "test@example.com").Return(nil, domain.ErrUserNotFound)
				userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "invalid username",
			request: &domain.UserCreateRequest{
				Username: "ab",
				Email:    "test@example.com",
			},
			setupMocks:    func(userRepo *mocks.MockUserRepository, memoryRepo *mocks.MockMemoryRepository) {},
			expectedError: domain.ErrInvalidUsername,
		},
		{
			name: "invalid email",
			request: &domain.UserCreateRequest{
				Username: "testuser",
				Email:    "invalid-email",
			},
			setupMocks:    func(userRepo *mocks.MockUserRepository, memoryRepo *mocks.MockMemoryRepository) {},
			expectedError: domain.ErrInvalidEmail,
		},
		{
			name: "duplicate username",
			request: &domain.UserCreateRequest{
				Username: "testuser",
				Email:    "test@example.com",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, memoryRepo *mocks.MockMemoryRepository) {
				existingUser := &domain.User{ID: uuid.New(), Username: "testuser"}
				userRepo.On("GetByUsername", "testuser").Return(existingUser, nil)
			},
			expectedError: domain.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(mocks.MockUserRepository)
			memoryRepo := new(mocks.MockMemoryRepository)
			userUsecase := usecase.NewUserUsecase(userRepo, memoryRepo)

			tt.setupMocks(userRepo, memoryRepo)

			user, err := userUsecase.CreateUser(tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Username, user.Username)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.True(t, user.IsActive)
			}

			userRepo.AssertExpectations(t)
			memoryRepo.AssertExpectations(t)
		})
	}
}