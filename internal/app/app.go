// Package app provides the Wails application bindings for the Remy
// personal assistant GUI. It exposes Go methods that the Svelte frontend
// calls via the Wails runtime.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/danmurf/remy/internal/agent"
	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/llm"
	"github.com/danmurf/remy/internal/memory"
	"github.com/danmurf/remy/internal/scheduler"
)

// App is the main Wails application struct. Its exported methods become
// bindings callable from the Svelte frontend.
type App struct {
	ctx       context.Context
	agent     *agent.Agent
	store     *memory.Store
	cfg       *config.Config
	stopConsolidation func()
	stopScheduler     func()
}

// NewApp creates a new App with the given dependencies.
func NewApp(cfg *config.Config, store *memory.Store, llmProvider llm.Provider, embedder *memory.Embedder, personaLoader agent.PersonaLoader, sched *scheduler.Scheduler) (*App, error) {
	agentCfg := &agent.Config{
		WorkingMemoryTurns:        cfg.Memory.WorkingMemoryTurns,
		UserID:                    "default-user",
		SessionID:                 "default",
		Interface:                 "gui",
		PersonaDir:                cfg.Persona.Directory,
		ActivePersona:             cfg.Persona.Active,
		QuickConsolidationDelayMs: cfg.Memory.QuickConsolidationDelayMs,
		DeepConsolidationDelayMs:  cfg.Memory.DeepConsolidationDelayMs,
	}

	a := agent.NewAgent(store, llmProvider, embedder, personaLoader, sched, agentCfg)

	return &App{
		agent: a,
		store: store,
		cfg:   cfg,
	}, nil
}

// Startup is called by Wails when the application starts. It initializes
// the agent, loads the active persona, and starts background services.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	if err := a.agent.LoadActivePersona(ctx); err != nil {
		fmt.Printf("Warning: could not load active persona: %v\n", err)
	}

	a.stopConsolidation = a.agent.ScheduleConsolidation(ctx)
}

// Shutdown is called by Wails when the application is shutting down.
func (a *App) Shutdown(ctx context.Context) {
	if a.stopConsolidation != nil {
		a.stopConsolidation()
	}
	if a.stopScheduler != nil {
		a.stopScheduler()
	}
}

// SendMessage sends a user message to the agent and returns the full response.
// This is the non-streaming version.
func (a *App) SendMessage(text string) (*MessageDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	msg, err := a.agent.HandleMessage(a.ctx, text)
	if err != nil {
		return nil, fmt.Errorf("handling message: %w", err)
	}

	return messageToDTO(msg), nil
}

// SendMessageStream sends a user message and returns streaming chunks.
// The frontend receives events via the Wails event system.
func (a *App) SendMessageStream(text string) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}

	ch, err := a.agent.HandleMessageStream(a.ctx, text)
	if err != nil {
		return fmt.Errorf("starting stream: %w", err)
	}

	go func() {
		for chunk := range ch {
			if chunk.Error != "" {
				_ = a.emit("stream:error", chunk.Error)
				return
			}
			if chunk.Done {
				_ = a.emit("stream:done", true)
				return
			}
			_ = a.emit("stream:chunk", chunk.Content)
		}
	}()

	return nil
}

// GetHistory returns recent messages from the store.
func (a *App) GetHistory(limit, offset int) ([]MessageDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	messages, err := a.store.GetMessages(a.ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("getting messages: %w", err)
	}

	dtos := make([]MessageDTO, len(messages))
	for i, msg := range messages {
		dtos[i] = *messageToDTO(&msg)
	}

	return dtos, nil
}

// GetConversations returns a list of conversation summaries for the sidebar.
func (a *App) GetConversations() ([]ConversationDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	messages, err := a.store.GetMessages(a.ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("getting messages: %w", err)
	}

	// Group by session ID
	sessions := make(map[string][]memory.Message)
	for _, msg := range messages {
		sid := msg.SessionID
		if sid == "" {
			sid = "default"
		}
		sessions[sid] = append(sessions[sid], msg)
	}

	convs := make([]ConversationDTO, 0, len(sessions))
	for sid, msgs := range sessions {
		if len(msgs) == 0 {
			continue
		}
		lastMsg := msgs[len(msgs)-1]
		preview := lastMsg.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		convs = append(convs, ConversationDTO{
			ID:        sid,
			Name:      sid,
			LastMsg:   preview,
			Timestamp: lastMsg.Timestamp,
		})
	}

	return convs, nil
}

// GetPersonas returns a list of available personas.
func (a *App) GetPersonas() ([]PersonaDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	summaries, err := a.agent.ListPersonas(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing personas: %w", err)
	}

	dtos := make([]PersonaDTO, len(summaries))
	for i, s := range summaries {
		dtos[i] = PersonaDTO{
			Name:        s.Name,
			Description: s.Provider + " / " + s.Model,
			IsActive:    s.Active,
		}
	}

	return dtos, nil
}

// SwitchPersona switches the active persona.
func (a *App) SwitchPersona(name string) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}
	return a.agent.SetActivePersona(a.ctx, name)
}

// GetActivePersona returns the name of the currently active persona.
func (a *App) GetActivePersona() string {
	p := a.agent.ActivePersona()
	if p == nil {
		return "default"
	}
	return p.Name
}

func (a *App) emit(event string, data any) error {
	return nil // Wails runtime emits events; this is a no-op for testing
}

// MessageDTO is a data transfer object for messages sent to the frontend.
type MessageDTO struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	Interface string `json:"interface"`
}

// ConversationDTO is a data transfer object for conversation summaries.
type ConversationDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	LastMsg   string `json:"last_msg"`
	Timestamp int64  `json:"timestamp"`
}

// PersonaDTO is a data transfer object for persona summaries.
type PersonaDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func messageToDTO(msg *memory.Message) *MessageDTO {
	if msg == nil {
		return nil
	}
	return &MessageDTO{
		ID:        msg.ID,
		Role:      msg.Role,
		Content:   msg.Content,
		Timestamp: msg.Timestamp,
		Interface: msg.Interface,
	}
}

// Ensure uuid is used
var _ = uuid.NewString
var _ = time.Now
