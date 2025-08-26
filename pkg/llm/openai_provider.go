package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface using OpenAI's API
type OpenAIProvider struct {
	client          *openai.Client
	config          *Config
	embeddingModel  string
	completionModel string
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(config *Config) *OpenAIProvider {
	client := openai.NewClient(config.APIKey)
	
	// Set custom base URL if provided
	if config.BaseURL != "" {
		clientConfig := openai.DefaultConfig(config.APIKey)
		clientConfig.BaseURL = config.BaseURL
		client = openai.NewClientWithConfig(clientConfig)
	}

	embeddingModel := config.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = string(openai.AdaEmbeddingV2) // text-embedding-ada-002
	}

	completionModel := config.CompletionModel
	if completionModel == "" {
		completionModel = openai.GPT3Dot5Turbo
	}

	return &OpenAIProvider{
		client:          client,
		config:          config,
		embeddingModel:  embeddingModel,
		completionModel: completionModel,
	}
}

// GenerateEmbeddings generates embeddings for the given texts
func (p *OpenAIProvider) GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	model := req.Model
	if model == "" {
		model = p.embeddingModel
	}

	// Apply timeout if configured
	if p.config.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(p.config.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	openaiReq := openai.EmbeddingRequest{
		Input: req.Input,
		Model: openai.EmbeddingModel(model),
	}

	resp, err := p.client.CreateEmbeddings(ctx, openaiReq)
	if err != nil {
		return nil, p.handleError(err)
	}

	// Convert OpenAI response to our format
	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return &EmbeddingResponse{
		Embeddings: embeddings,
		Model:      model,
		Usage: Usage{
			PromptTokens: resp.Usage.PromptTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}, nil
}

// GenerateCompletion generates text completion for the given messages
func (p *OpenAIProvider) GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.completionModel
	}

	// Apply timeout if configured
	if p.config.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(p.config.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	// Convert our messages to OpenAI format
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}

	// Convert tools if provided
	if len(req.Tools) > 0 {
		tools := make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			tools[i] = openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		openaiReq.Tools = tools
	}

	resp, err := p.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, p.handleError(err)
	}

	if len(resp.Choices) == 0 {
		return nil, &Error{Type: "no_choices", Message: "no completion choices returned"}
	}

	choice := resp.Choices[0]
	response := &CompletionResponse{
		Content: choice.Message.Content,
		Model:   model,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// Convert tool calls if present
	if len(choice.Message.ToolCalls) > 0 {
		toolCalls := make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			toolCalls[i] = ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
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
func (p *OpenAIProvider) GetEmbeddingDimension(model string) int {
	switch model {
	case string(openai.AdaEmbeddingV2):
		return 1536
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-3-large":
		return 3072
	default:
		return 1536 // Default to ada-002 dimensions
	}
}

// GetDefaultModel returns the default embedding model
func (p *OpenAIProvider) GetDefaultModel() string {
	return p.embeddingModel
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// IsHealthy checks if the provider is healthy and accessible
func (p *OpenAIProvider) IsHealthy(ctx context.Context) error {
	// Test with a simple embedding request
	req := &EmbeddingRequest{
		Input: []string{"test"},
		Model: p.embeddingModel,
	}

	_, err := p.GenerateEmbeddings(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// handleError converts OpenAI errors to our error format
func (p *OpenAIProvider) handleError(err error) error {
	switch e := err.(type) {
	case *openai.APIError:
		switch e.HTTPStatusCode {
		case 401:
			return ErrInvalidAPIKey
		case 429:
			return ErrRateLimitExceeded
		case 404:
			return ErrModelNotFound
		case 400:
			return ErrInvalidRequest
		case 500, 502, 503, 504:
			return ErrServiceUnavailable
		default:
			return &Error{
				Type:    "api_error",
				Message: e.Message,
				Code:    e.HTTPStatusCode,
			}
		}
	case *openai.RequestError:
		return &Error{
			Type:    "request_error",
			Message: e.Error(),
		}
	default:
		return &Error{
			Type:    "unknown_error",
			Message: err.Error(),
		}
	}
}