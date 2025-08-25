package user

import "context"

// Service defines the business operations for users
type Service interface {
	// CreateUser creates a new user with validation
	CreateUser(ctx context.Context, req CreateRequest) (*User, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id ID) (*User, error)

	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id ID, req UpdateRequest) (*User, error)

	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id ID) error

	// ListUsers returns a list of users with pagination
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)

	// UpdateLastLogin updates the user's last login time
	UpdateLastLogin(ctx context.Context, id ID) error

	// GetUserStats returns user statistics
	GetUserStats(ctx context.Context) (*Stats, error)
}
