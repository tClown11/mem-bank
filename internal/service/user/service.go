package user

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"mem_bank/internal/domain/user"
)

// service implements user.Service interface
type service struct {
	repo user.Repository
}

// NewService creates a new user service
func NewService(repo user.Repository) user.Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateUser(ctx context.Context, req user.CreateRequest) (*user.User, error) {
	// Validation
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check for existing username
	existingUser, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil && err != user.ErrNotFound {
		return nil, fmt.Errorf("checking existing username: %w", err)
	}
	if existingUser != nil {
		return nil, user.ErrUsernameTaken
	}

	// Check for existing email
	existingUser, err = s.repo.FindByEmail(ctx, req.Email)
	if err != nil && err != user.ErrNotFound {
		return nil, fmt.Errorf("checking existing email: %w", err)
	}
	if existingUser != nil {
		return nil, user.ErrEmailTaken
	}

	// Create new user
	u := user.NewUser(req.Username, req.Email, req.Profile, req.Settings)

	// Set default settings
	s.setDefaultSettings(&u.Settings)

	// Store user
	if err := s.repo.Store(ctx, u); err != nil {
		return nil, fmt.Errorf("storing user: %w", err)
	}

	return u, nil
}

func (s *service) GetUser(ctx context.Context, id user.ID) (*user.User, error) {
	if id.IsZero() {
		return nil, user.ErrInvalidID
	}

	return s.repo.FindByID(ctx, id)
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, user.ErrInvalidUsername
	}

	return s.repo.FindByUsername(ctx, username)
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, user.ErrInvalidEmail
	}

	return s.repo.FindByEmail(ctx, email)
}

func (s *service) UpdateUser(ctx context.Context, id user.ID, req user.UpdateRequest) (*user.User, error) {
	if id.IsZero() {
		return nil, user.ErrInvalidID
	}

	// Get existing user
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Username != nil && *req.Username != u.Username {
		// Check username availability
		existing, err := s.repo.FindByUsername(ctx, *req.Username)
		if err != nil && err != user.ErrNotFound {
			return nil, fmt.Errorf("checking username availability: %w", err)
		}
		if existing != nil {
			return nil, user.ErrUsernameTaken
		}
		u.Username = *req.Username
	}

	if req.Email != nil && *req.Email != u.Email {
		// Check email availability
		existing, err := s.repo.FindByEmail(ctx, *req.Email)
		if err != nil && err != user.ErrNotFound {
			return nil, fmt.Errorf("checking email availability: %w", err)
		}
		if existing != nil {
			return nil, user.ErrEmailTaken
		}
		u.Email = *req.Email
	}

	if req.Profile != nil {
		u.UpdateProfile(*req.Profile)
	}

	if req.Settings != nil {
		u.UpdateSettings(*req.Settings)
	}

	if req.IsActive != nil {
		if *req.IsActive {
			u.Activate()
		} else {
			u.Deactivate()
		}
	}

	// Update user
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	return u, nil
}

func (s *service) DeleteUser(ctx context.Context, id user.ID) error {
	if id.IsZero() {
		return user.ErrInvalidID
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) ListUsers(ctx context.Context, limit, offset int) ([]*user.User, error) {
	if limit <= 0 {
		limit = 20 // default limit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *service) UpdateLastLogin(ctx context.Context, id user.ID) error {
	if id.IsZero() {
		return user.ErrInvalidID
	}

	return s.repo.UpdateLastLogin(ctx, id)
}

func (s *service) GetUserStats(ctx context.Context) (*user.Stats, error) {
	totalUsers, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting total users: %w", err)
	}

	activeUsers, err := s.repo.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting active users: %w", err)
	}

	// For simplicity, new users and total memories are set to 0
	// These would require additional queries or be passed from another service
	return &user.Stats{
		TotalUsers:    totalUsers,
		ActiveUsers:   activeUsers,
		NewUsers:      0,
		TotalMemories: 0,
	}, nil
}

// Validation helpers
func (s *service) validateCreateRequest(req user.CreateRequest) error {
	if strings.TrimSpace(req.Username) == "" {
		return user.ErrInvalidUsername
	}

	if len(req.Username) < 3 || len(req.Username) > 50 {
		return user.ErrInvalidUsername
	}

	if !isValidEmail(req.Email) {
		return user.ErrInvalidEmail
	}

	return nil
}

func (s *service) setDefaultSettings(settings *user.Settings) {
	if settings.Language == "" {
		settings.Language = "en"
	}
	if settings.Timezone == "" {
		settings.Timezone = "UTC"
	}
	if settings.MemoryRetention == 0 {
		settings.MemoryRetention = 365 // days
	}
	if settings.PrivacyLevel == "" {
		settings.PrivacyLevel = "private"
	}
	if settings.NotificationSettings == nil {
		settings.NotificationSettings = map[string]bool{
			"email":   true,
			"browser": true,
			"desktop": false,
			"mobile":  true,
		}
	}
	if settings.EmbeddingModel == "" {
		settings.EmbeddingModel = "text-embedding-ada-002"
	}
	if settings.MaxMemories == 0 {
		settings.MaxMemories = 10000
	}
	// AutoSummary defaults to false (zero value)
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}