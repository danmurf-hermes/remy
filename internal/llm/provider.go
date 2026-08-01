package llm

import (
	"context"
	"fmt"

	"github.com/yourname/remy/internal/config"
)

//go:generate go run go.uber.org/mock/mockgen -destination=mock_llm/mock_llm.go -package=mock_llm . Provider

// Provider is the interface for LLM backends. Implementations must support
// chat, streaming chat, and text embedding.
type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
	Embed(ctx context.Context, text string) ([]float32, error)
}

// NewProvider creates a Provider from the given config. Currently only
// Ollama-compatible endpoints are supported.
func NewProvider(cfg config.ProviderConfig) (Provider, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("provider endpoint is required")
	}
	if cfg.ChatModel == "" {
		return nil, fmt.Errorf("chat model is required")
	}
	if cfg.EmbeddingModel == "" {
		return nil, fmt.Errorf("embedding model is required")
	}

	return &OllamaClient{
		endpoint:       cfg.Endpoint,
		apiKey:         cfg.APIKey,
		chatModel:      cfg.ChatModel,
		embeddingModel: cfg.EmbeddingModel,
		parameters:     cfg.Parameters,
		httpClient:     defaultHTTPClient(),
	}, nil
}
