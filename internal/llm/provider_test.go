package llm_test

import (
	"testing"

	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/llm"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.ProviderConfig
		wantErr bool
	}{
		{
			name: "valid",
			cfg: config.ProviderConfig{
				Endpoint:       "http://localhost:11434/v1",
				ChatModel:      "llama3.1:8b",
				EmbeddingModel: "nomic-embed-text",
			},
		},
		{
			name: "with parameters",
			cfg: config.ProviderConfig{
				Endpoint:       "http://localhost:11434/v1",
				ChatModel:      "llama3.1:8b",
				EmbeddingModel: "nomic-embed-text",
				Parameters:     map[string]any{"temperature": 0.9, "max_tokens": 2048},
			},
		},
		{
			name: "with API key",
			cfg: config.ProviderConfig{
				Endpoint:       "http://localhost:11434/v1",
				APIKey:         "sk-test-123",
				ChatModel:      "llama3.1:8b",
				EmbeddingModel: "nomic-embed-text",
			},
		},
		{
			name: "missing endpoint",
			cfg: config.ProviderConfig{
				Endpoint:       "",
				ChatModel:      "llama3.1:8b",
				EmbeddingModel: "nomic-embed-text",
			},
			wantErr: true,
		},
		{
			name: "missing chat model",
			cfg: config.ProviderConfig{
				Endpoint:       "http://localhost:11434/v1",
				ChatModel:      "",
				EmbeddingModel: "nomic-embed-text",
			},
			wantErr: true,
		},
		{
			name: "missing embedding model",
			cfg: config.ProviderConfig{
				Endpoint:       "http://localhost:11434/v1",
				ChatModel:      "llama3.1:8b",
				EmbeddingModel: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := llm.NewProvider(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewProvider: %v", err)
			}
			if p == nil {
				t.Fatal("provider is nil")
			}
			client, ok := p.(*llm.OllamaClient)
			if !ok {
				t.Fatalf("expected *OllamaClient, got %T", p)
			}
			if client.Endpoint != tt.cfg.Endpoint {
				t.Errorf("endpoint = %q, want %q", client.Endpoint, tt.cfg.Endpoint)
			}
			if client.ChatModel != tt.cfg.ChatModel {
				t.Errorf("chatModel = %q, want %q", client.ChatModel, tt.cfg.ChatModel)
			}
			if client.EmbeddingModel != tt.cfg.EmbeddingModel {
				t.Errorf("embeddingModel = %q, want %q", client.EmbeddingModel, tt.cfg.EmbeddingModel)
			}
			if tt.cfg.APIKey != "" && client.APIKey != tt.cfg.APIKey {
				t.Errorf("apiKey = %q, want %q", client.APIKey, tt.cfg.APIKey)
			}
			if tt.cfg.Parameters != nil {
				if client.Parameters["temperature"] != tt.cfg.Parameters["temperature"] {
					t.Errorf("temperature = %v, want %v", client.Parameters["temperature"], tt.cfg.Parameters["temperature"])
				}
			}
		})
	}
}
