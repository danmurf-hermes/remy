// Package agent implements the core orchestration loop for the Remy
// personal assistant. It receives user messages, retrieves relevant context
// from memory, builds prompts, calls the LLM, and stores responses.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/remy/internal/llm"
	"github.com/yourname/remy/internal/memory"
)

// Store defines the subset of memory.Store methods the agent needs,
// making it easy to mock in tests.
type Store interface {
	SaveMessage(ctx context.Context, msg *memory.Message) error
	GetMessage(ctx context.Context, id string) (*memory.Message, error)
	GetMessages(ctx context.Context, limit, offset int) ([]memory.Message, error)
	GetMessagesBySession(ctx context.Context, sessionID string) ([]memory.Message, error)
	SearchEpisodes(ctx context.Context, embedding []byte, limit int) ([]memory.Episode, error)
	SearchFacts(ctx context.Context, embedding []byte, limit int) ([]memory.Fact, error)
	GetScratchpad(ctx context.Context) (string, error)
	UpdateScratchpad(ctx context.Context, content string) error
	LogActivity(ctx context.Context, entry *memory.ActivityEntry) error
	SaveMessageVector(ctx context.Context, messageID string, embedding []byte) error
}

// Embedder defines the interface for generating vector embeddings.
type Embedder interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// Config holds configuration for the agent's behavior.
type Config struct {
	WorkingMemoryTurns int
	UserID             string
	SessionID          string
	Interface          string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		WorkingMemoryTurns: 20,
		UserID:             "default-user",
		SessionID:          "default",
		Interface:          "gui",
	}
}

// Agent is the core orchestration engine. It receives user messages,
// retrieves relevant context, builds prompts, calls the LLM, stores
// responses, and returns them.
type Agent struct {
	store    Store
	provider llm.Provider
	embedder Embedder
	cfg      Config
}

// NewAgent creates a new Agent with the given store, LLM provider,
// embedder, and configuration.
func NewAgent(store Store, provider llm.Provider, embedder Embedder, cfg Config) *Agent {
	return &Agent{
		store:    store,
		provider: provider,
		embedder: embedder,
		cfg:      cfg,
	}
}

// HandleMessage is the core agent loop. It processes a single user message
// through the full pipeline: store, retrieve context, build prompt, call
// LLM, store response, and return it.
func (a *Agent) HandleMessage(ctx context.Context, userMsg string) (*memory.Message, error) {
	now := time.Now().UnixMilli()

	userID := a.cfg.UserID
	sessionID := a.cfg.SessionID
	interfaceName := a.cfg.Interface

	userMessage := &memory.Message{
		ID:        uuid.NewString(),
		UserID:    userID,
		Role:      "user",
		Content:   userMsg,
		Timestamp: now,
		Interface: interfaceName,
		SessionID: sessionID,
	}

	if err := a.store.SaveMessage(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("saving user message: %w", err)
	}

	a.logActivity(ctx, "message_received", userMessage.ID, sessionID)

	userEmbedding, err := a.embedder.GenerateEmbedding(ctx, userMsg)
	if err != nil {
		return nil, fmt.Errorf("generating embedding: %w", err)
	}

	embeddingBytes, err := memory.SerializeVector(userEmbedding)
	if err != nil {
		return nil, fmt.Errorf("serializing embedding: %w", err)
	}

	if err := a.store.SaveMessageVector(ctx, userMessage.ID, embeddingBytes); err != nil {
		return nil, fmt.Errorf("saving message vector: %w", err)
	}

	episodes, err := a.store.SearchEpisodes(ctx, embeddingBytes, 5)
	if err != nil {
		return nil, fmt.Errorf("searching episodes: %w", err)
	}

	facts, err := a.store.SearchFacts(ctx, embeddingBytes, 5)
	if err != nil {
		return nil, fmt.Errorf("searching facts: %w", err)
	}

	scratchpad, err := a.store.GetScratchpad(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting scratchpad: %w", err)
	}

	recentMessages, err := a.store.GetMessages(ctx, a.cfg.WorkingMemoryTurns, 0)
	if err != nil {
		return nil, fmt.Errorf("getting recent messages: %w", err)
	}

	prompt := BuildPrompt(&PromptInput{
		Scratchpad:     scratchpad,
		Episodes:       episodes,
		Facts:          facts,
		RecentMessages: recentMessages,
		UserMessage:    userMsg,
	})

	llmResp, err := a.provider.Chat(ctx, llm.ChatRequest{
		Messages: prompt.Messages,
	})
	if err != nil {
		return nil, fmt.Errorf("calling LLM: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	responseContent := llmResp.Choices[0].Message.Content

	responseTime := time.Now().UnixMilli()
	responseMessage := &memory.Message{
		ID:        uuid.NewString(),
		UserID:    userID,
		Role:      "assistant",
		Content:   responseContent,
		Timestamp: responseTime,
		Interface: interfaceName,
		SessionID: sessionID,
	}

	if err := a.store.SaveMessage(ctx, responseMessage); err != nil {
		return nil, fmt.Errorf("saving response message: %w", err)
	}

	a.logActivity(ctx, "message_sent", responseMessage.ID, sessionID)

	respEmbedding, err := a.embedder.GenerateEmbedding(ctx, responseContent)
	if err != nil {
		return nil, fmt.Errorf("generating response embedding: %w", err)
	}

	respEmbeddingBytes, err := memory.SerializeVector(respEmbedding)
	if err != nil {
		return nil, fmt.Errorf("serializing response embedding: %w", err)
	}

	if err := a.store.SaveMessageVector(ctx, responseMessage.ID, respEmbeddingBytes); err != nil {
		return nil, fmt.Errorf("saving response message vector: %w", err)
	}

	return responseMessage, nil
}

func (a *Agent) logActivity(ctx context.Context, activityType, messageID, sessionID string) {
	_ = a.store.LogActivity(ctx, &memory.ActivityEntry{
		ID:        uuid.NewString(),
		Timestamp: time.Now().UnixMilli(),
		Type:      activityType,
		Details:   "{}",
		MessageID: messageID,
		SessionID: sessionID,
	})
}
