package llm

import (
	"context"
	"fmt"

	"github.com/yourname/remy/internal/config"
)

type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
	Embed(ctx context.Context, text string) ([]float32, error)
}

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
