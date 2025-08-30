package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mem_bank/internal/domain/user"
	"mem_bank/internal/model"
	"mem_bank/internal/query"
)

// postgresRepository implements user.Repository using PostgreSQL
type postgresRepository struct {
	db *gorm.DB
	q  *query.Query
}

// NewPostgresRepository creates a new PostgreSQL-based user repository
func NewPostgresRepository(db *gorm.DB) user.Repository {
	return &postgresRepository{
		db: db,
		q:  query.Use(db),
	}
}

func (r *postgresRepository) Store(ctx context.Context, u *user.User) error {
	gormUser, err := r.toModel(u)
	if err != nil {
		return fmt.Errorf("converting to model: %w", err)
	}

	if err := r.q.User.WithContext(ctx).Create(gormUser); err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	return nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	gormUser, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("finding user: %w", err)
	}

	return r.toDomain(gormUser)
}

func (r *postgresRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	gormUser, err := r.q.User.WithContext(ctx).Where(r.q.User.Username.Eq(username)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by username: %w", err)
	}

	return r.toDomain(gormUser)
}

func (r *postgresRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	gormUser, err := r.q.User.WithContext(ctx).Where(r.q.User.Email.Eq(email)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}

	return r.toDomain(gormUser)
}

func (r *postgresRepository) Update(ctx context.Context, u *user.User) error {
	gormUser, err := r.toModel(u)
	if err != nil {
		return fmt.Errorf("converting to model: %w", err)
	}

	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(u.ID.String())).Updates(gormUser)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}

	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id user.ID) error {
	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).Delete()
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}

	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (r *postgresRepository) UpdateLastLogin(ctx context.Context, id user.ID) error {
	now := time.Now()
	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).Updates(map[string]interface{}{
		"last_login": &now,
		"updated_at": &now,
	})
	if err != nil {
		return fmt.Errorf("updating last login: %w", err)
	}

	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (r *postgresRepository) UpdateSettings(ctx context.Context, id user.ID, settings user.Settings) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	now := time.Now()
	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).Updates(map[string]interface{}{
		"settings":   stringPtr(string(settingsJSON)),
		"updated_at": &now,
	})
	if err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}

	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (r *postgresRepository) UpdateProfile(ctx context.Context, id user.ID, profile user.Profile) error {
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("marshaling profile: %w", err)
	}

	now := time.Now()
	result, err := r.q.User.WithContext(ctx).Where(r.q.User.ID.Eq(id.String())).Updates(map[string]interface{}{
		"profile":    stringPtr(string(profileJSON)),
		"updated_at": &now,
	})
	if err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}

	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (r *postgresRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	gormUsers, err := r.q.User.WithContext(ctx).
		Order(r.q.User.CreatedAt.Desc()).
		Limit(limit).
		Offset(offset).
		Find()
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	users := make([]*user.User, 0, len(gormUsers))
	for _, gormUser := range gormUsers {
		u, err := r.toDomain(gormUser)
		if err != nil {
			return nil, fmt.Errorf("converting user: %w", err)
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *postgresRepository) Count(ctx context.Context) (int, error) {
	count, err := r.q.User.WithContext(ctx).Count()
	if err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}

	return int(count), nil
}

func (r *postgresRepository) CountActive(ctx context.Context) (int, error) {
	isActive := true
	count, err := r.q.User.WithContext(ctx).Where(r.q.User.IsActive.Is(isActive)).Count()
	if err != nil {
		return 0, fmt.Errorf("counting active users: %w", err)
	}

	return int(count), nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func (r *postgresRepository) toModel(u *user.User) (*model.User, error) {
	profile, err := json.Marshal(u.Profile)
	if err != nil {
		return nil, fmt.Errorf("marshaling profile: %w", err)
	}

	settings, err := json.Marshal(u.Settings)
	if err != nil {
		return nil, fmt.Errorf("marshaling settings: %w", err)
	}

	gormUser := &model.User{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		Profile:   stringPtr(string(profile)),
		Settings:  stringPtr(string(settings)),
		CreatedAt: &u.CreatedAt,
		UpdatedAt: &u.UpdatedAt,
		IsActive:  &u.IsActive,
	}

	if !u.LastLogin.IsZero() {
		gormUser.LastLogin = &u.LastLogin
	}

	return gormUser, nil
}

func (r *postgresRepository) toDomain(gormUser *model.User) (*user.User, error) {
	id, err := uuid.Parse(gormUser.ID)
	if err != nil {
		return nil, fmt.Errorf("parsing user ID: %w", err)
	}

	u := &user.User{
		ID:       user.ID(id),
		Username: gormUser.Username,
		Email:    gormUser.Email,
		IsActive: true,
	}

	if gormUser.Profile != nil && *gormUser.Profile != "" && *gormUser.Profile != "{}" {
		if err := json.Unmarshal([]byte(*gormUser.Profile), &u.Profile); err != nil {
			return nil, fmt.Errorf("unmarshaling profile: %w", err)
		}
	}

	if gormUser.Settings != nil && *gormUser.Settings != "" && *gormUser.Settings != "{}" {
		if err := json.Unmarshal([]byte(*gormUser.Settings), &u.Settings); err != nil {
			return nil, fmt.Errorf("unmarshaling settings: %w", err)
		}
	}

	if gormUser.CreatedAt != nil {
		u.CreatedAt = *gormUser.CreatedAt
	}

	if gormUser.UpdatedAt != nil {
		u.UpdatedAt = *gormUser.UpdatedAt
	}

	if gormUser.LastLogin != nil {
		u.LastLogin = *gormUser.LastLogin
	}

	if gormUser.IsActive != nil {
		u.IsActive = *gormUser.IsActive
	}

	return u, nil
}
