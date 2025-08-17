package usecase

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"mem_bank/internal/domain"
)

type UserUsecase interface {
	CreateUser(req *domain.UserCreateRequest) (*domain.User, error)
	GetUserByID(id uuid.UUID) (*domain.User, error)
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	UpdateUser(id uuid.UUID, req *domain.UserUpdateRequest) (*domain.User, error)
	DeleteUser(id uuid.UUID) error
	UpdateLastLogin(id uuid.UUID) error
	UpdateUserSettings(id uuid.UUID, settings domain.UserSettings) error
	UpdateUserProfile(id uuid.UUID, profile domain.UserProfile) error
	ListUsers(limit, offset int) ([]*domain.User, error)
	GetUserStats() (*domain.UserStats, error)
}

type userUsecase struct {
	userRepo   domain.UserRepository
	memoryRepo domain.MemoryRepository
}

func NewUserUsecase(userRepo domain.UserRepository, memoryRepo domain.MemoryRepository) UserUsecase {
	return &userUsecase{
		userRepo:   userRepo,
		memoryRepo: memoryRepo,
	}
}

func (u *userUsecase) CreateUser(req *domain.UserCreateRequest) (*domain.User, error) {
	if err := u.validateCreateRequest(req); err != nil {
		return nil, err
	}

	existingUser, _ := u.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	existingUser, _ = u.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Profile:  req.Profile,
		Settings: req.Settings,
		IsActive: true,
	}

	u.setDefaultSettings(&user.Settings)

	err := u.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (u *userUsecase) GetUserByID(id uuid.UUID) (*domain.User, error) {
	user, err := u.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) GetUserByUsername(username string) (*domain.User, error) {
	if username == "" {
		return nil, domain.ErrInvalidUsername
	}

	user, err := u.userRepo.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) GetUserByEmail(email string) (*domain.User, error) {
	if !u.isValidEmail(email) {
		return nil, domain.ErrInvalidEmail
	}

	user, err := u.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) UpdateUser(id uuid.UUID, req *domain.UserUpdateRequest) (*domain.User, error) {
	user, err := u.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Username != nil {
		if *req.Username == "" {
			return nil, domain.ErrInvalidUsername
		}
		existingUser, _ := u.userRepo.GetByUsername(*req.Username)
		if existingUser != nil && existingUser.ID != id {
			return nil, domain.ErrUserAlreadyExists
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		if !u.isValidEmail(*req.Email) {
			return nil, domain.ErrInvalidEmail
		}
		existingUser, _ := u.userRepo.GetByEmail(*req.Email)
		if existingUser != nil && existingUser.ID != id {
			return nil, domain.ErrUserAlreadyExists
		}
		user.Email = *req.Email
	}

	if req.Profile != nil {
		user.Profile = *req.Profile
	}

	if req.Settings != nil {
		user.Settings = *req.Settings
		u.setDefaultSettings(&user.Settings)
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	err = u.userRepo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (u *userUsecase) DeleteUser(id uuid.UUID) error {
	_, err := u.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	err = u.userRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (u *userUsecase) UpdateLastLogin(id uuid.UUID) error {
	_, err := u.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	err = u.userRepo.UpdateLastLogin(id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

func (u *userUsecase) UpdateUserSettings(id uuid.UUID, settings domain.UserSettings) error {
	_, err := u.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	u.setDefaultSettings(&settings)

	err = u.userRepo.UpdateSettings(id, settings)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

func (u *userUsecase) UpdateUserProfile(id uuid.UUID, profile domain.UserProfile) error {
	_, err := u.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	err = u.userRepo.UpdateProfile(id, profile)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (u *userUsecase) ListUsers(limit, offset int) ([]*domain.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	users, err := u.userRepo.List(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

func (u *userUsecase) GetUserStats() (*domain.UserStats, error) {
	totalUsers, err := u.userRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	stats := &domain.UserStats{
		TotalUsers: totalUsers,
	}

	return stats, nil
}

func (u *userUsecase) validateCreateRequest(req *domain.UserCreateRequest) error {
	if req.Username == "" {
		return domain.ErrInvalidUsername
	}

	if len(req.Username) < 3 || len(req.Username) > 50 {
		return domain.ErrInvalidUsername
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(req.Username) {
		return domain.ErrInvalidUsername
	}

	if !u.isValidEmail(req.Email) {
		return domain.ErrInvalidEmail
	}

	return nil
}

func (u *userUsecase) isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (u *userUsecase) setDefaultSettings(settings *domain.UserSettings) {
	if settings.Language == "" {
		settings.Language = "en"
	}
	if settings.Timezone == "" {
		settings.Timezone = "UTC"
	}
	if settings.MemoryRetention == 0 {
		settings.MemoryRetention = 365
	}
	if settings.PrivacyLevel == "" {
		settings.PrivacyLevel = "standard"
	}
	if settings.NotificationSettings == nil {
		settings.NotificationSettings = map[string]bool{
			"email":    true,
			"push":     false,
			"browser":  true,
		}
	}
	if settings.EmbeddingModel == "" {
		settings.EmbeddingModel = "text-embedding-ada-002"
	}
	if settings.MaxMemories == 0 {
		settings.MaxMemories = 10000
	}
}