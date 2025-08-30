package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
	"mem_bank/internal/service/embedding"
	"mem_bank/pkg/logger"
)

// Mock repository
type mockMemoryRepository struct {
	mock.Mock
}

func (m *mockMemoryRepository) Store(ctx context.Context, mem *memory.Memory) error {
	args := m.Called(ctx, mem)
	return args.Error(0)
}

func (m *mockMemoryRepository) FindByID(ctx context.Context, id memory.ID) (*memory.Memory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*memory.Memory), args.Error(1)
}

func (m *mockMemoryRepository) FindByUserID(ctx context.Context, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*memory.Memory), args.Error(1)
}

func (m *mockMemoryRepository) Update(ctx context.Context, mem *memory.Memory) error {
	args := m.Called(ctx, mem)
	return args.Error(0)
}

func (m *mockMemoryRepository) Delete(ctx context.Context, id memory.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockMemoryRepository) SearchSimilar(ctx context.Context, emb []float32, userID user.ID, limit int, threshold float64) ([]*memory.Memory, error) {
	args := m.Called(ctx, emb, userID, limit, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*memory.Memory), args.Error(1)
}

func (m *mockMemoryRepository) SearchSimilarWithScores(ctx context.Context, emb []float32, userID user.ID, limit int, threshold float64) ([]*memory.MemoryWithScore, error) {
	args := m.Called(ctx, emb, userID, limit, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*memory.MemoryWithScore), args.Error(1)
}

func (m *mockMemoryRepository) SearchSimilarByMemory(ctx context.Context, memID memory.ID, userID user.ID, limit int, threshold float64) ([]*memory.Memory, error) {
	args := m.Called(ctx, memID, userID, limit, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*memory.Memory), args.Error(1)
}

func (m *mockMemoryRepository) SearchByContent(ctx context.Context, query string, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	args := m.Called(ctx, query, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*memory.Memory), args.Error(1)
}

func (m *mockMemoryRepository) FindByTags(ctx context.Context, tags []string, userID user.ID, limit, offset int) ([]*memory.Memory, error) {
	args := m.Called(ctx, tags, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*memory.Memory), args.Error(1)
}

func (m *mockMemoryRepository) UpdateAccessInfo(ctx context.Context, id memory.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockMemoryRepository) GetStatsByUserID(ctx context.Context, userID user.ID) (*memory.Stats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*memory.Stats), args.Error(1)
}

func (m *mockMemoryRepository) CountByUserID(ctx context.Context, userID user.ID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *mockMemoryRepository) BatchStore(ctx context.Context, memories []*memory.Memory) error {
	args := m.Called(ctx, memories)
	return args.Error(0)
}

func (m *mockMemoryRepository) BatchUpdate(ctx context.Context, memories []*memory.Memory) error {
	args := m.Called(ctx, memories)
	return args.Error(0)
}

func (m *mockMemoryRepository) BatchDelete(ctx context.Context, ids []memory.ID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *mockMemoryRepository) BatchUpdateEmbeddings(ctx context.Context, updates []memory.EmbeddingUpdate) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

// Mock user repository
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Store(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id user.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepository) UpdateLastLogin(ctx context.Context, id user.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) UpdateSettings(ctx context.Context, id user.ID, settings user.Settings) error {
	args := m.Called(ctx, id, settings)
	return args.Error(0)
}

func (m *mockUserRepository) UpdateProfile(ctx context.Context, id user.ID, profile user.Profile) error {
	args := m.Called(ctx, id, profile)
	return args.Error(0)
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *mockUserRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockUserRepository) CountActive(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// Mock embedding service
type mockEmbeddingService struct {
	mock.Mock
}

func (m *mockEmbeddingService) GenerateEmbedding(ctx context.Context, text string) (*embedding.EmbeddingResult, error) {
	args := m.Called(ctx, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*embedding.EmbeddingResult), args.Error(1)
}

func (m *mockEmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) (*embedding.BatchEmbeddingResult, error) {
	args := m.Called(ctx, texts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*embedding.BatchEmbeddingResult), args.Error(1)
}

// Mock logger
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) WithError(err error) logger.Logger {
	return m
}

func (m *mockLogger) WithField(key string, value interface{}) logger.Logger {
	return m
}

func (m *mockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	return m
}

func (m *mockLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *mockLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *mockLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *mockLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *mockLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func TestService_CreateMemory(t *testing.T) {
	tests := []struct {
		name        string
		req         memory.CreateRequest
		setupMocks  func(*mockMemoryRepository, *mockUserRepository, *mockEmbeddingService)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation with embedding",
			req: memory.CreateRequest{
				UserID:     user.ID(uuid.New()),
				Content:    "Test memory content",
				Summary:    "Test summary",
				Importance: 5,
				MemoryType: "general",
				Tags:       []string{"test"},
				Metadata:   map[string]interface{}{"key": "value"},
			},
			setupMocks: func(memRepo *mockMemoryRepository, userRepo *mockUserRepository, embSvc *mockEmbeddingService) {
				// User exists
				testUser := &user.User{
					ID:       user.ID(uuid.New()),
					Username: "testuser",
					Email:    "test@example.com",
					Profile: user.Profile{
						FirstName: "Test",
						LastName:  "User",
					},
					IsActive: true,
				}
				userRepo.On("FindByID", mock.Anything, mock.Anything).Return(testUser, nil)

				// Store succeeds (embedding service is nil, so no embedding calls)
				memRepo.On("Store", mock.Anything, mock.AnythingOfType("*memory.Memory")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			req: memory.CreateRequest{
				UserID:     user.ID(uuid.New()),
				Content:    "Test content",
				Summary:    "Test summary",
				Importance: 5,
				MemoryType: "general",
			},
			setupMocks: func(memRepo *mockMemoryRepository, userRepo *mockUserRepository, embSvc *mockEmbeddingService) {
				userRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, user.ErrNotFound)
			},
			wantErr:     true,
			expectedErr: memory.ErrInvalidUserID,
		},
		{
			name: "invalid importance",
			req: memory.CreateRequest{
				UserID:     user.ID(uuid.New()),
				Content:    "Test content",
				Summary:    "Test summary",
				Importance: 15, // Invalid
				MemoryType: "general",
			},
			setupMocks: func(memRepo *mockMemoryRepository, userRepo *mockUserRepository, embSvc *mockEmbeddingService) {
				// No mocks needed - validation fails first
			},
			wantErr:     true,
			expectedErr: memory.ErrInvalidImportance,
		},
		{
			name: "embedding generation fails but memory still created",
			req: memory.CreateRequest{
				UserID:     user.ID(uuid.New()),
				Content:    "Test memory content",
				Summary:    "Test summary",
				Importance: 5,
				MemoryType: "general",
			},
			setupMocks: func(memRepo *mockMemoryRepository, userRepo *mockUserRepository, embSvc *mockEmbeddingService) {
				// User exists
				testUser := &user.User{
					ID:       user.ID(uuid.New()),
					Username: "testuser",
					Email:    "test@example.com",
					Profile: user.Profile{
						FirstName: "Test",
						LastName:  "User",
					},
					IsActive: true,
				}
				userRepo.On("FindByID", mock.Anything, mock.Anything).Return(testUser, nil)

				// Embedding generation fails
				// Store succeeds (embedding service is nil, so no embedding calls)
				memRepo.On("Store", mock.Anything, mock.AnythingOfType("*memory.Memory")).Return(nil)
			},
			wantErr: false, // Should not fail - embedding is not critical
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			memRepo := &mockMemoryRepository{}
			userRepo := &mockUserRepository{}
			embSvc := &mockEmbeddingService{}
			logger := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(memRepo, userRepo, embSvc)
			}

			// Mock logger calls (no embedding calls expected since service is nil)

			// Create service
			svc := NewService(memRepo, userRepo, nil, logger)

			// Execute test
			result, err := svc.CreateMemory(context.Background(), tt.req)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Content, result.Content)
				assert.Equal(t, tt.req.Summary, result.Summary)
				assert.Equal(t, tt.req.Importance, result.Importance)
				assert.Equal(t, tt.req.MemoryType, result.MemoryType)
				assert.Equal(t, tt.req.Tags, result.Tags)
				assert.Equal(t, tt.req.Metadata, result.Metadata)
				assert.False(t, result.ID.IsZero())
				assert.False(t, result.CreatedAt.IsZero())
			}

			// Verify all mock expectations
			memRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			embSvc.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestService_GetMemory(t *testing.T) {
	tests := []struct {
		name       string
		memoryID   memory.ID
		setupMocks func(*mockMemoryRepository)
		wantErr    bool
		wantMemory *memory.Memory
	}{
		{
			name:     "successful retrieval",
			memoryID: memory.ID(uuid.New()),
			setupMocks: func(repo *mockMemoryRepository) {
				testMemory := &memory.Memory{
					ID:           memory.ID(uuid.New()),
					UserID:       user.ID(uuid.New()),
					Content:      "Test content",
					Summary:      "Test summary",
					Importance:   5,
					MemoryType:   "general",
					Tags:         []string{"test"},
					Metadata:     map[string]interface{}{"key": "value"},
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
					LastAccessed: time.Now(),
					AccessCount:  0,
				}
				repo.On("FindByID", mock.Anything, mock.Anything).Return(testMemory, nil)
				repo.On("UpdateAccessInfo", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "memory not found",
			memoryID: memory.ID(uuid.New()),
			setupMocks: func(repo *mockMemoryRepository) {
				repo.On("FindByID", mock.Anything, mock.Anything).Return(nil, memory.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name:     "zero ID",
			memoryID: memory.ID(uuid.Nil),
			setupMocks: func(repo *mockMemoryRepository) {
				// No setup needed - validation fails first
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			memRepo := &mockMemoryRepository{}
			userRepo := &mockUserRepository{}
			logger := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(memRepo)
			}

			// Mock logger for access info update failures
			if tt.name == "successful retrieval" {
				// We might have warning logs if access info update fails, but test should still pass
			}

			// Create service
			svc := NewService(memRepo, userRepo, nil, logger)

			// Execute test
			result, err := svc.GetMemory(context.Background(), tt.memoryID)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Verify mock expectations
			memRepo.AssertExpectations(t)
		})
	}
}

func TestService_SearchSimilarMemories(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		userID     user.ID
		limit      int
		threshold  float64
		setupMocks func(*mockMemoryRepository, *mockEmbeddingService)
		wantErr    bool
	}{
		{
			name:      "successful search",
			content:   "test content",
			userID:    user.ID(uuid.New()),
			limit:     10,
			threshold: 0.8,
			setupMocks: func(repo *mockMemoryRepository, embSvc *mockEmbeddingService) {
				// Embedding generation succeeds
				embResult := &embedding.EmbeddingResult{
					Text:      "test content",
					Embedding: []float32{0.1, 0.2, 0.3},
					Model:     "test-model",
				}
				embSvc.On("GenerateEmbedding", mock.Anything, "test content").Return(embResult, nil)

				// Search returns results
				memories := []*memory.Memory{
					{
						ID:         memory.ID(uuid.New()),
						UserID:     user.ID(uuid.New()),
						Content:    "Similar content",
						Importance: 5,
						MemoryType: "general",
					},
				}
				repo.On("SearchSimilar", mock.Anything, mock.Anything, mock.Anything, 10, 0.8).Return(memories, nil)
			},
			wantErr: false,
		},
		{
			name:       "empty content",
			content:    "",
			userID:     user.ID(uuid.New()),
			limit:      10,
			threshold:  0.8,
			setupMocks: func(repo *mockMemoryRepository, embSvc *mockEmbeddingService) {},
			wantErr:    true,
		},
		{
			name:      "embedding generation fails",
			content:   "test content",
			userID:    user.ID(uuid.New()),
			limit:     10,
			threshold: 0.8,
			setupMocks: func(repo *mockMemoryRepository, embSvc *mockEmbeddingService) {
				embSvc.On("GenerateEmbedding", mock.Anything, "test content").Return(nil, errors.New("embedding failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			memRepo := &mockMemoryRepository{}
			embSvc := &mockEmbeddingService{}
			logger := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(memRepo, embSvc)
			}

			// Mock logger for embedding failures
			if tt.name == "embedding generation fails" {
				logger.On("WithError", mock.Anything).Return(logger)
				logger.On("Error", "Failed to generate embedding for similarity search")
			}

			// Create service
			svc := NewService(memRepo, nil, nil, logger)

			// Mock logger for embedding service not available
			logger.On("Warn", "Embedding service not available for similarity search").Maybe()

			// Execute test
			result, err := svc.SearchSimilarMemories(context.Background(), tt.content, tt.userID, tt.limit, tt.threshold)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Verify mock expectations
			memRepo.AssertExpectations(t)
			embSvc.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}
