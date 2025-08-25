package user

import "context"

// Repository defines the interface for user data access operations
type Repository interface {
	// Store creates a new user in the repository
	Store(ctx context.Context, user *User) error

	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id ID) (*User, error)

	// FindByUsername retrieves a user by their username
	FindByUsername(ctx context.Context, username string) (*User, error)

	// FindByEmail retrieves a user by their email
	FindByEmail(ctx context.Context, email string) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// Delete removes a user by their ID
	Delete(ctx context.Context, id ID) error

	// UpdateLastLogin updates the last login time for a user
	UpdateLastLogin(ctx context.Context, id ID) error

	// UpdateSettings updates user settings
	UpdateSettings(ctx context.Context, id ID, settings Settings) error

	// UpdateProfile updates user profile
	UpdateProfile(ctx context.Context, id ID, profile Profile) error

	// List retrieves users with pagination
	List(ctx context.Context, limit, offset int) ([]*User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int, error)

	// CountActive returns the total number of active users
	CountActive(ctx context.Context) (int, error)
}
