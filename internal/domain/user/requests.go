package user

// CreateRequest represents a request to create a new user
type CreateRequest struct {
	Username string
	Email    string
	Profile  Profile
	Settings Settings
}

// UpdateRequest represents a request to update an existing user
type UpdateRequest struct {
	Username *string
	Email    *string
	Profile  *Profile
	Settings *Settings
	IsActive *bool
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string
	Password string
}

// Stats represents user statistics
type Stats struct {
	TotalUsers    int
	ActiveUsers   int
	NewUsers      int
	TotalMemories int
}
