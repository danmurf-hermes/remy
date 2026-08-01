// Package agent implements the core orchestration loop for the Remy
// personal assistant. It receives user messages, retrieves relevant context
// from memory, builds prompts, calls the LLM, and stores responses.
package agent

//go:generate go run go.uber.org/mock/mockgen -destination=mock_agent/mock_agent.go -package=mock_agent . Store,Embedder,PersonaLoader

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/danmurf/remy/internal/llm"
	"github.com/danmurf/remy/internal/memory"
	"github.com/danmurf/remy/internal/persona"
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

// PersonaLoader defines the interface for loading and listing personas,
// making it easy to mock in tests.
type PersonaLoader interface {
	LoadPersona(path string) (*persona.Persona, error)
	ListPersonas(dir string, activeName string) ([]persona.Summary, error)
}

// Config holds configuration for the agent's behavior.
type Config struct {
	WorkingMemoryTurns int
	UserID             string
	SessionID          string
	Interface          string
	PersonaDir         string
	ActivePersona      string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		WorkingMemoryTurns: 20,
		UserID:             "default-user",
		SessionID:          "default",
		Interface:          "gui",
		ActivePersona:      "default",
	}
}

// Agent is the core orchestration engine. It receives user messages,
// retrieves relevant context, builds prompts, calls the LLM, stores
// responses, and returns them.
type Agent struct {
	store         Store
	provider      llm.Provider
	embedder      Embedder
	personaLoader PersonaLoader
	cfg           *Config
	activePersona *persona.Persona
}

// NewAgent creates a new Agent with the given store, LLM provider,
// embedder, persona loader, and configuration.
func NewAgent(store Store, provider llm.Provider, embedder Embedder, personaLoader PersonaLoader, cfg *Config) *Agent {
	return &Agent{
		store:         store,
		provider:      provider,
		embedder:      embedder,
		personaLoader: personaLoader,
		cfg:           cfg,
	}
}

// LoadActivePersona loads the active persona from disk. If the persona
// file doesn't exist, it returns nil (no persona override).
func (a *Agent) LoadActivePersona(ctx context.Context) error {
	if a.cfg.PersonaDir == "" || a.cfg.ActivePersona == "" {
		return nil
	}
	path := a.cfg.PersonaDir + "/" + a.cfg.ActivePersona + ".md"
	p, err := a.personaLoader.LoadPersona(path)
	if err != nil {
		if _, ok := err.(*persona.ErrPersonaNotFound); ok {
			a.activePersona = nil
			return nil
		}
		return fmt.Errorf("loading active persona: %w", err)
	}
	a.activePersona = p
	return nil
}

// ActivePersona returns the currently active persona, or nil if none is loaded.
func (a *Agent) ActivePersona() *persona.Persona {
	return a.activePersona
}

// SetActivePersona sets the active persona by name and loads it from disk.
func (a *Agent) SetActivePersona(ctx context.Context, name string) error {
	if a.cfg.PersonaDir == "" {
		return fmt.Errorf("persona directory not configured")
	}
	path := a.cfg.PersonaDir + "/" + name + ".md"
	p, err := a.personaLoader.LoadPersona(path)
	if err != nil {
		return fmt.Errorf("loading persona %q: %w", name, err)
	}
	a.activePersona = p
	a.cfg.ActivePersona = name
	return nil
}

// ListPersonas returns a summary of all available personas.
func (a *Agent) ListPersonas(ctx context.Context) ([]persona.Summary, error) {
	return a.personaLoader.ListPersonas(a.cfg.PersonaDir, a.cfg.ActivePersona)
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

	// Check for persona switch request
	personaSwitch := DetectPersonaSwitch(userMsg, a.activePersona)
	if personaSwitch != "" {
		if err := a.SetActivePersona(ctx, personaSwitch); err != nil {
			// Log the error but continue with current persona
			a.logActivity(ctx, "error", userMessage.ID, sessionID)
		}
	}

	prompt := BuildPrompt(&PromptInput{
		Scratchpad:     scratchpad,
		Episodes:       episodes,
		Facts:          facts,
		RecentMessages: recentMessages,
		UserMessage:    userMsg,
		Persona:        a.activePersona,
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

// DetectPersonaSwitch checks if the user message contains a request to
// switch personas. It returns the persona name if found, or empty string.
func DetectPersonaSwitch(userMsg string, currentPersona *persona.Persona) string {
	lower := strings.ToLower(strings.TrimSpace(userMsg))

	// Patterns: "switch to <name>", "switch to <name> mode", "change to <name> persona"
	prefixes := []string{
		"switch to ",
		"change to ",
		"use ",
		"activate ",
	}

	for _, prefix := range prefixes {
		if !strings.HasPrefix(lower, prefix) {
			continue
		}
		rest := strings.TrimPrefix(lower, prefix)
		// Remove trailing "mode", "persona", "personality"
		rest = strings.TrimSuffix(rest, " mode")
		rest = strings.TrimSuffix(rest, " persona")
		rest = strings.TrimSuffix(rest, " personality")
		rest = strings.TrimSpace(rest)
		// Take only the first word
		if idx := strings.Index(rest, " "); idx >= 0 {
			rest = rest[:idx]
		}
		if rest != "" && (currentPersona == nil || rest != currentPersona.Name) {
			return rest
		}
	}

	return ""
}
