package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaProvider implements the Provider interface using Ollama's local API
type OllamaProvider struct {
	client          *http.Client
	config          *Config
	embeddingModel  string
	completionModel string
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider(config *Config) *OllamaProvider {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if config.TimeoutSeconds > 0 {
		client.Timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	embeddingModel := config.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "nomic-embed-text" // Default Ollama embedding model
	}

	completionModel := config.CompletionModel
	if completionModel == "" {
		completionModel = "llama2" // Default Ollama completion model
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama URL
	}

	return &OllamaProvider{
		client:          client,
		config:          config,
		embeddingModel:  embeddingModel,
		completionModel: completionModel,
	}
}

// ollamaEmbeddingRequest represents Ollama's embedding request format
type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaEmbeddingResponse represents Ollama's embedding response format
type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// ollamaCompletionRequest represents Ollama's completion request format
type ollamaCompletionRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream"`
	Tools    []ollamaToolDefinition `json:"tools,omitempty"`
}

// ollamaMessage represents a message in Ollama format
type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaToolDefinition represents a tool definition in Ollama format
type ollamaToolDefinition struct {
	Type     string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"function"`
}

// ollamaCompletionResponse represents Ollama's completion response format
type ollamaCompletionResponse struct {
	Message struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		ToolCalls []struct {
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls,omitempty"`
	} `json:"message"`
	Done bool `json:"done"`
}

// GenerateEmbeddings generates embeddings for the given texts using Ollama
func (p *OllamaProvider) GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	model := req.Model
	if model == "" {
		model = p.embeddingModel
	}

	embeddings := make([][]float32, len(req.Input))

	// Generate embeddings for each input text
	for i, text := range req.Input {
		ollamaReq := ollamaEmbeddingRequest{
			Model:  model,
			Prompt: text,
		}

		reqBytes, err := json.Marshal(ollamaReq)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/embeddings", bytes.NewBuffer(reqBytes))
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(httpReq)
		if err != nil {
			return nil, p.handleError(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
		}

		var ollamaResp ollamaEmbeddingResponse
		if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}

		// Convert float64 to float32
		embedding := make([]float32, len(ollamaResp.Embedding))
		for j, v := range ollamaResp.Embedding {
			embedding[j] = float32(v)
		}
		embeddings[i] = embedding
	}

	return &EmbeddingResponse{
		Embeddings: embeddings,
		Model:      model,
		Usage: Usage{
			PromptTokens: len(req.Input) * 10, // Rough estimate
			TotalTokens:  len(req.Input) * 10,
		},
	}, nil
}

// GenerateCompletion generates text completion using Ollama
func (p *OllamaProvider) GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.completionModel
	}

	// Convert messages to Ollama format
	messages := make([]ollamaMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	ollamaReq := ollamaCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	// Convert tools if provided
	if len(req.Tools) > 0 {
		tools := make([]ollamaToolDefinition, len(req.Tools))
		for i, tool := range req.Tools {
			tools[i] = ollamaToolDefinition{
				Type: tool.Type,
			}
			tools[i].Function.Name = tool.Function.Name
			tools[i].Function.Description = tool.Function.Description
			tools[i].Function.Parameters = tool.Function.Parameters
		}
		ollamaReq.Tools = tools
	}

	reqBytes, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/chat", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, p.handleError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	response := &CompletionResponse{
		Content: ollamaResp.Message.Content,
		Model:   model,
		Usage: Usage{
			PromptTokens:     len(reqBytes) / 4, // Rough estimate
			CompletionTokens: len(ollamaResp.Message.Content) / 4,
			TotalTokens:      (len(reqBytes) + len(ollamaResp.Message.Content)) / 4,
		},
	}

	// Convert tool calls if present
	if len(ollamaResp.Message.ToolCalls) > 0 {
		toolCalls := make([]ToolCall, len(ollamaResp.Message.ToolCalls))
		for i, tc := range ollamaResp.Message.ToolCalls {
			toolCalls[i] = ToolCall{
				ID:   fmt.Sprintf("call_%d", i),
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
		response.ToolCalls = toolCalls
	}

	return response, nil
}

// GetEmbeddingDimension returns the dimension of embeddings for the given model
func (p *OllamaProvider) GetEmbeddingDimension(model string) int {
	switch model {
	case "nomic-embed-text":
		return 768
	case "all-minilm":
		return 384
	default:
		return 768 // Default assumption
	}
}

// GetDefaultModel returns the default embedding model
func (p *OllamaProvider) GetDefaultModel() string {
	return p.embeddingModel
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// IsHealthy checks if the provider is healthy and accessible
func (p *OllamaProvider) IsHealthy(ctx context.Context) error {
	// Test with a simple request to check if Ollama is running
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("creating health check request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// handleError converts HTTP errors to our error format
func (p *OllamaProvider) handleError(err error) error {
	return &Error{
		Type:    "ollama_error",
		Message: fmt.Sprintf("Ollama API error: %v", err),
	}
}
