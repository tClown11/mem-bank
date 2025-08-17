package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Profile   UserProfile `json:"profile" db:"profile"`
	Settings  UserSettings `json:"settings" db:"settings"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	LastLogin time.Time `json:"last_login" db:"last_login"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

type UserProfile struct {
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	Avatar      string            `json:"avatar"`
	Bio         string            `json:"bio"`
	Preferences map[string]interface{} `json:"preferences"`
}

type UserSettings struct {
	Language         string  `json:"language"`
	Timezone         string  `json:"timezone"`
	MemoryRetention  int     `json:"memory_retention"`
	PrivacyLevel     string  `json:"privacy_level"`
	NotificationSettings map[string]bool `json:"notification_settings"`
	EmbeddingModel   string  `json:"embedding_model"`
	MaxMemories      int     `json:"max_memories"`
	AutoSummary      bool    `json:"auto_summary"`
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	UpdateLastLogin(id uuid.UUID) error
	UpdateSettings(id uuid.UUID, settings UserSettings) error
	UpdateProfile(id uuid.UUID, profile UserProfile) error
	List(limit, offset int) ([]*User, error)
	Count() (int, error)
}

type UserCreateRequest struct {
	Username string      `json:"username" binding:"required"`
	Email    string      `json:"email" binding:"required,email"`
	Profile  UserProfile `json:"profile"`
	Settings UserSettings `json:"settings"`
}

type UserUpdateRequest struct {
	Username *string      `json:"username,omitempty"`
	Email    *string      `json:"email,omitempty"`
	Profile  *UserProfile `json:"profile,omitempty"`
	Settings *UserSettings `json:"settings,omitempty"`
	IsActive *bool        `json:"is_active,omitempty"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserStats struct {
	TotalUsers   int `json:"total_users"`
	ActiveUsers  int `json:"active_users"`
	NewUsers     int `json:"new_users"`
	TotalMemories int `json:"total_memories"`
}