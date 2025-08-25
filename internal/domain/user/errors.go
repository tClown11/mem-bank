package user

import "errors"

// Domain-specific errors for user operations
var (
	ErrNotFound        = errors.New("user not found")
	ErrAlreadyExists   = errors.New("user already exists")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidID       = errors.New("invalid user ID")
	ErrEmailTaken      = errors.New("email already taken")
	ErrUsernameTaken   = errors.New("username already taken")
	ErrInactive        = errors.New("user is inactive")
)
