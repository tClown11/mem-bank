package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mem_bank/internal/domain"
	"mem_bank/internal/model"
	"mem_bank/internal/query"
)

type userRepository struct {
	db *gorm.DB
	q  *query.Query
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{
		db: db,
		q:  query.Use(db),
	}
}

func (r *userRepository) Create(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	profileJSON, err := json.Marshal(user.Profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	settingsJSON, err := json.Marshal(user.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Convert to GORM model
	gormUser := &model.User{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Profile:   stringPtr(string(profileJSON)),
		Settings:  stringPtr(string(settingsJSON)),
		CreatedAt: &user.CreatedAt,
		UpdatedAt: &user.UpdatedAt,
		IsActive:  &user.IsActive,
	}

	err = r.q.User.WithContext(ctx).Create(gormUser)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gormUser, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user, err := convertToDomainUser(gormUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gormUser, err := r.q.User.WithContext(ctx).Where(r.q.User.Username.Eq(username)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	user, err := convertToDomainUser(gormUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gormUser, err := r.q.User.WithContext(ctx).Where(r.q.User.Email.Eq(email)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user, err := convertToDomainUser(gormUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()
	gormUser := convertToModelUser(user)

	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(user.ID.String())).Updates(gormUser)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).Delete()
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdateLastLogin(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).Updates(map[string]interface{}{
		"last_login": &now,
		"updated_at": &now,
	})
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdateSettings(id uuid.UUID, settings domain.UserSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	now := time.Now()
	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).Updates(map[string]interface{}{
		"settings":   stringPtr(string(settingsJSON)),
		"updated_at": &now,
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdateProfile(id uuid.UUID, profile domain.UserProfile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	now := time.Now()
	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).Updates(map[string]interface{}{
		"profile":    stringPtr(string(profileJSON)),
		"updated_at": &now,
	})
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) List(limit, offset int) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gormUsers, err := r.q.User.WithContext(ctx).
		Order(r.q.User.CreatedAt.Desc()).
		Limit(limit).
		Offset(offset).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var users []*domain.User
	for _, gormUser := range gormUsers {
		user, err := convertToDomainUser(gormUser)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := r.q.User.WithContext(ctx).Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return int(count), nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func convertToModelUser(user *domain.User) *model.User {
	profileJSON, _ := json.Marshal(user.Profile)
	settingsJSON, _ := json.Marshal(user.Settings)
	
	gormUser := &model.User{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Profile:   stringPtr(string(profileJSON)),
		Settings:  stringPtr(string(settingsJSON)),
		CreatedAt: &user.CreatedAt,
		UpdatedAt: &user.UpdatedAt,
		IsActive:  &user.IsActive,
	}
	
	if !user.LastLogin.IsZero() {
		gormUser.LastLogin = &user.LastLogin
	}
	
	return gormUser
}

func convertToDomainUser(gormUser *model.User) (*domain.User, error) {
	id, err := uuid.Parse(gormUser.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	user := &domain.User{
		ID:       id,
		Username: gormUser.Username,
		Email:    gormUser.Email,
		IsActive: true,
	}
	
	if gormUser.Profile != nil && *gormUser.Profile != "" {
		if err := json.Unmarshal([]byte(*gormUser.Profile), &user.Profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
		}
	}
	
	if gormUser.Settings != nil && *gormUser.Settings != "" {
		if err := json.Unmarshal([]byte(*gormUser.Settings), &user.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}
	
	if gormUser.CreatedAt != nil {
		user.CreatedAt = *gormUser.CreatedAt
	}
	
	if gormUser.UpdatedAt != nil {
		user.UpdatedAt = *gormUser.UpdatedAt
	}
	
	if gormUser.LastLogin != nil {
		user.LastLogin = *gormUser.LastLogin
	}
	
	if gormUser.IsActive != nil {
		user.IsActive = *gormUser.IsActive
	}
	
	return user, nil
}