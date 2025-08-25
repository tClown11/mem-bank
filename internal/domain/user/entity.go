package user

import (
	"time"

	"github.com/google/uuid"
)

// ID represents a user identifier
type ID uuid.UUID

// User represents a core business entity for users
type User struct {
	ID        ID
	Username  string
	Email     string
	Profile   Profile
	Settings  Settings
	CreatedAt time.Time
	UpdatedAt time.Time
	LastLogin time.Time
	IsActive  bool
}

// Profile represents user profile information
type Profile struct {
	FirstName   string
	LastName    string
	Avatar      string
	Bio         string
	Preferences map[string]interface{}
}

// Settings represents user configuration and preferences
type Settings struct {
	Language             string
	Timezone             string
	MemoryRetention      int
	PrivacyLevel         string
	NotificationSettings map[string]bool
	EmbeddingModel       string
	MaxMemories          int
	AutoSummary          bool
}

// NewID creates a new user ID
func NewID() ID {
	return ID(uuid.New())
}

// String returns the string representation of the user ID
func (id ID) String() string {
	return uuid.UUID(id).String()
}

// IsZero checks if the ID is zero
func (id ID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// NewUser creates a new user with default values
func NewUser(username, email string, profile Profile, settings Settings) *User {
	return &User{
		ID:        NewID(),
		Username:  username,
		Email:     email,
		Profile:   profile,
		Settings:  settings,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}
}

// UpdateLastLogin updates the user's last login time
func (u *User) UpdateLastLogin() {
	u.LastLogin = time.Now()
	u.UpdatedAt = time.Now()
}

// UpdateProfile updates the user's profile
func (u *User) UpdateProfile(profile Profile) {
	u.Profile = profile
	u.UpdatedAt = time.Now()
}

// UpdateSettings updates the user's settings
func (u *User) UpdateSettings(settings Settings) {
	u.Settings = settings
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// Activate activates the user
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}
