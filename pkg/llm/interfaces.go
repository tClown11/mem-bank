package llm

import (
	"context"
)

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model,omitempty"`
}

// EmbeddingResponse represents the response from an embedding generation
type EmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Model      string      `json:"model"`
	Usage      Usage       `json:"usage"`
}

// CompletionRequest represents a request for text completion
type CompletionRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model,omitempty"`
	Tools    []Tool    `json:"tools,omitempty"`
}

// CompletionResponse represents the response from text completion
type CompletionResponse struct {
	Content   string    `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Model     string    `json:"model"`
	Usage     Usage     `json:"usage"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// Tool represents a tool/function that can be called by the LLM
type Tool struct {
	Type     string   `json:"type"`     // "function"
	Function Function `json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a tool call made by the LLM
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Usage represents API usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingProvider defines the interface for embedding generation
type EmbeddingProvider interface {
	// GenerateEmbeddings generates embeddings for the given texts
	GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
	
	// GetEmbeddingDimension returns the dimension of embeddings for the given model
	GetEmbeddingDimension(model string) int
	
	// GetDefaultModel returns the default embedding model
	GetDefaultModel() string
}

// CompletionProvider defines the interface for text completion
type CompletionProvider interface {
	// GenerateCompletion generates text completion for the given messages
	GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	
	// GetDefaultModel returns the default completion model
	GetDefaultModel() string
}

// Provider combines both embedding and completion capabilities
type Provider interface {
	EmbeddingProvider
	CompletionProvider
	
	// Name returns the provider name
	Name() string
	
	// IsHealthy checks if the provider is healthy and accessible
	IsHealthy(ctx context.Context) error
}

// Config holds LLM provider configuration
type Config struct {
	// Provider type (e.g., "openai", "azure", "local")
	Provider string `mapstructure:"provider"`
	
	// API key for authenticated providers
	APIKey string `mapstructure:"api_key"`
	
	// Base URL for the API (optional, uses default if not set)
	BaseURL string `mapstructure:"base_url"`
	
	// Default embedding model
	EmbeddingModel string `mapstructure:"embedding_model"`
	
	// Default completion model
	CompletionModel string `mapstructure:"completion_model"`
	
	// Request timeout in seconds
	TimeoutSeconds int `mapstructure:"timeout_seconds"`
	
	// Maximum retries for failed requests
	MaxRetries int `mapstructure:"max_retries"`
	
	// Rate limiting - requests per minute
	RateLimit int `mapstructure:"rate_limit"`
}

// Error types for LLM operations
type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// Common error types
var (
	ErrInvalidAPIKey    = &Error{Type: "invalid_api_key", Message: "invalid API key"}
	ErrRateLimitExceeded = &Error{Type: "rate_limit_exceeded", Message: "rate limit exceeded"}
	ErrModelNotFound    = &Error{Type: "model_not_found", Message: "model not found"}
	ErrInvalidRequest   = &Error{Type: "invalid_request", Message: "invalid request"}
	ErrServiceUnavailable = &Error{Type: "service_unavailable", Message: "service unavailable"}
)