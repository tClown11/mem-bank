package domain

import "errors"

var (
	ErrMemoryNotFound     = errors.New("memory not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidMemoryType  = errors.New("invalid memory type")
	ErrInvalidImportance  = errors.New("importance must be between 1 and 10")
	ErrEmptyContent       = errors.New("memory content cannot be empty")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrMemoryLimitExceeded = errors.New("memory limit exceeded")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInvalidEmbedding   = errors.New("invalid embedding format")
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrInvalidInput       = errors.New("invalid input parameters")
)

type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e DomainError) Error() string {
	return e.Message
}

func NewDomainError(code, message string, details map[string]interface{}) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Details: details,
	}
}