package memory

import "errors"

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
