package llm

import (
	"testing"

	"github.com/yourname/remy/internal/config"
)

func TestNewProvider_Valid(t *testing.T) {
	cfg := config.ProviderConfig{
		Endpoint:       "http://localhost:11434/v1",
		ChatModel:      "llama3.1:8b",
		EmbeddingModel: "nomic-embed-text",
	}
	p, err := NewProvider(cfg)
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
	if client.endpoint != cfg.Endpoint {
		t.Errorf("endpoint = %q, want %q", client.endpoint, cfg.Endpoint)
	}
	if client.chatModel != cfg.ChatModel {
		t.Errorf("chatModel = %q, want %q", client.chatModel, cfg.ChatModel)
	}
	if client.embeddingModel != cfg.EmbeddingModel {
		t.Errorf("embeddingModel = %q, want %q", client.embeddingModel, cfg.EmbeddingModel)
	}
}

func TestNewProvider_WithParameters(t *testing.T) {
	cfg := config.ProviderConfig{
		Endpoint:       "http://localhost:11434/v1",
		ChatModel:      "llama3.1:8b",
		EmbeddingModel: "nomic-embed-text",
		Parameters: map[string]any{
			"temperature": 0.9,
			"max_tokens":  2048,
		},
	}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	client := p.(*OllamaClient)
	if client.parameters["temperature"] != 0.9 {
		t.Errorf("temperature = %v, want 0.9", client.parameters["temperature"])
	}
	if client.parameters["max_tokens"] != 2048 {
		t.Errorf("max_tokens = %v (%T), want 2048 (int)", client.parameters["max_tokens"], client.parameters["max_tokens"])
	}
}

func TestNewProvider_WithAPIKey(t *testing.T) {
	cfg := config.ProviderConfig{
		Endpoint:       "http://localhost:11434/v1",
		APIKey:         "sk-test-123",
		ChatModel:      "llama3.1:8b",
		EmbeddingModel: "nomic-embed-text",
	}
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	client := p.(*OllamaClient)
	if client.apiKey != "sk-test-123" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "sk-test-123")
	}
}

func TestNewProvider_MissingEndpoint(t *testing.T) {
	cfg := config.ProviderConfig{
		Endpoint:       "",
		ChatModel:      "llama3.1:8b",
		EmbeddingModel: "nomic-embed-text",
	}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error for missing endpoint, got nil")
	}
}

func TestNewProvider_MissingChatModel(t *testing.T) {
	cfg := config.ProviderConfig{
		Endpoint:       "http://localhost:11434/v1",
		ChatModel:      "",
		EmbeddingModel: "nomic-embed-text",
	}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error for missing chat model, got nil")
	}
}

func TestNewProvider_MissingEmbeddingModel(t *testing.T) {
	cfg := config.ProviderConfig{
		Endpoint:       "http://localhost:11434/v1",
		ChatModel:      "llama3.1:8b",
		EmbeddingModel: "",
	}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error for missing embedding model, got nil")
	}
}
