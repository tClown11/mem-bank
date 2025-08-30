package memory

import (
	"errors"
	"fmt"
)

// Domain-specific errors for memory operations
var (
	ErrNotFound          = errors.New("memory not found")
	ErrInvalidID         = errors.New("invalid memory ID")
	ErrInvalidUserID     = errors.New("invalid user ID")
	ErrInvalidContent    = errors.New("invalid memory content")
	ErrInvalidImportance = errors.New("invalid importance level")
	ErrInvalidMemoryType = errors.New("invalid memory type")
	ErrEmbeddingFailed   = errors.New("failed to generate embedding")
)

// ValidationError represents validation errors with specific field information
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// ServiceError represents a service-level error with additional context
type ServiceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

func (e ServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s (%s): %v", e.Message, e.Code, e.Cause)
	}
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

func (e ServiceError) Unwrap() error {
	return e.Cause
}

// NewServiceError creates a new service error
func NewServiceError(code, message string, cause error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Common service error codes
const (
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeExternalService  = "EXTERNAL_SERVICE_ERROR"
)
