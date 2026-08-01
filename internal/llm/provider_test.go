package llm

import (
	"testing"

	"github.com/yourname/remy/internal/config"
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
			p, err := NewProvider(tt.cfg)
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
			client, ok := p.(*OllamaClient)
			if !ok {
				t.Fatalf("expected *OllamaClient, got %T", p)
			}
			if client.endpoint != tt.cfg.Endpoint {
				t.Errorf("endpoint = %q, want %q", client.endpoint, tt.cfg.Endpoint)
			}
			if client.chatModel != tt.cfg.ChatModel {
				t.Errorf("chatModel = %q, want %q", client.chatModel, tt.cfg.ChatModel)
			}
			if client.embeddingModel != tt.cfg.EmbeddingModel {
				t.Errorf("embeddingModel = %q, want %q", client.embeddingModel, tt.cfg.EmbeddingModel)
			}
			if tt.cfg.APIKey != "" && client.apiKey != tt.cfg.APIKey {
				t.Errorf("apiKey = %q, want %q", client.apiKey, tt.cfg.APIKey)
			}
			if tt.cfg.Parameters != nil {
				if client.parameters["temperature"] != tt.cfg.Parameters["temperature"] {
					t.Errorf("temperature = %v, want %v", client.parameters["temperature"], tt.cfg.Parameters["temperature"])
				}
			}
		})
	}
}
