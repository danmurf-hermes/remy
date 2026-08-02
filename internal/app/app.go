// Package app provides the Wails application bindings for the Remy
// personal assistant GUI. It exposes Go methods that the Svelte frontend
// calls via the Wails runtime.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danmurf/remy/internal/agent"
	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/memory"
	"github.com/danmurf/remy/internal/persona"
	"github.com/google/uuid"
)

// Emitter defines the interface for emitting events to the frontend.
// The Wails runtime provides the production implementation.
type Emitter interface {
	Emit(event string, data any) error
}

// AgentService defines the subset of agent.Agent methods the app needs,
// making it easy to mock in tests.
type AgentService interface {
	HandleMessage(ctx context.Context, userMsg string) (*memory.Message, error)
	HandleMessageStream(ctx context.Context, userMsg string) (<-chan agent.StreamChunk, error)
	ListPersonas(ctx context.Context) ([]persona.Summary, error)
	SetActivePersona(ctx context.Context, name string) error
	ActivePersona() *persona.Persona
	LoadActivePersona(ctx context.Context) error
	ScheduleConsolidation(ctx context.Context) func()
}

// App is the main Wails application struct. Its exported methods become
// bindings callable from the Svelte frontend.
type App struct {
	ctx               context.Context
	agent             AgentService
	store             *memory.Store
	cfg               *config.Config
	cfgPath           string
	emitter           Emitter
	stopConsolidation func()
	stopScheduler     func()
}

// NewApp creates a new App with the given dependencies.
func NewApp(cfg *config.Config, cfgPath string, store *memory.Store, agentSvc AgentService) (*App, error) {
	return &App{
		agent:   agentSvc,
		store:   store,
		cfg:     cfg,
		cfgPath: cfgPath,
	}, nil
}

// SetEmitter sets the event emitter for the app. Used to wire the Wails
// runtime emitter after startup.
func (a *App) SetEmitter(e Emitter) {
	a.emitter = e
}

// Startup is called by Wails when the application starts. It initializes
// the agent, loads the active persona, and starts background services.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	if err := a.store.InitScratchpad(ctx); err != nil {
		slog.Warn("Could not initialize scratchpad", "error", err)
	}

	if err := a.agent.LoadActivePersona(ctx); err != nil {
		slog.Warn("Could not load active persona", "error", err)
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
		lastMsg := msgs[0]
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
	if a.emitter != nil {
		return a.emitter.Emit(event, data)
	}
	return nil
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

// FactDTO is a data transfer object for semantic facts.
type FactDTO struct {
	ID         string  `json:"id"`
	Fact       string  `json:"fact"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	CreatedAt  int64   `json:"created_at"`
	UpdatedAt  int64   `json:"updated_at"`
}

// EpisodeDTO is a data transfer object for episodic memory entries.
type EpisodeDTO struct {
	ID         string  `json:"id"`
	Summary    string  `json:"summary"`
	StartTime  int64   `json:"start_time"`
	EndTime    int64   `json:"end_time"`
	MessageIDs string  `json:"message_ids"`
	Importance float64 `json:"importance"`
	Topics     string  `json:"topics"`
}

// EntityDTO is a data transfer object for entities.
type EntityDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

// RelationshipDTO is a data transfer object for entity relationships.
type RelationshipDTO struct {
	ID           string  `json:"id"`
	SourceEntity string  `json:"source_entity"`
	TargetEntity string  `json:"target_entity"`
	Relationship string  `json:"relationship"`
	Confidence   float64 `json:"confidence"`
	CreatedAt    int64   `json:"created_at"`
}

// SearchResultsDTO holds the combined results of a memory search.
type SearchResultsDTO struct {
	Facts    []FactDTO    `json:"facts"`
	Episodes []EpisodeDTO `json:"episodes"`
}

// TaskDTO is a data transfer object for scheduled tasks.
type TaskDTO struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	TriggerAt int64  `json:"trigger_at"`
	CronExpr  string `json:"cron_expr"`
	Action    string `json:"action"`
	Context   string `json:"context"`
	CreatedAt int64  `json:"created_at"`
	FiredAt   int64  `json:"fired_at"`
}

// ActivityEntryDTO is a data transfer object for activity log entries.
type ActivityEntryDTO struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
	Details   string `json:"details"`
	MessageID string `json:"message_id"`
	SessionID string `json:"session_id"`
}

// ConfigDTO mirrors the config.Config struct for frontend consumption.
type ConfigDTO struct {
	Providers       map[string]ProviderConfigDTO `json:"providers"`
	DefaultProvider string                       `json:"default_provider"`
	Memory          MemoryConfigDTO              `json:"memory"`
	Persona         PersonaConfigDTO             `json:"persona"`
	Interfaces      InterfacesConfigDTO          `json:"interfaces"`
}

// ProviderConfigDTO is the DTO for a provider configuration.
type ProviderConfigDTO struct {
	Endpoint       string         `json:"endpoint"`
	ChatModel      string         `json:"chat_model"`
	EmbeddingModel string         `json:"embedding_model"`
	Parameters     map[string]any `json:"parameters,omitempty"`
}

// MemoryConfigDTO is the DTO for memory configuration.
type MemoryConfigDTO struct {
	DBPath                    string `json:"db_path"`
	WorkingMemoryTurns        int    `json:"working_memory_turns"`
	QuickConsolidationDelayMs int    `json:"quick_consolidation_delay_ms"`
	DeepConsolidationDelayMs  int    `json:"deep_consolidation_delay_ms"`
}

// PersonaConfigDTO is the DTO for persona configuration.
type PersonaConfigDTO struct {
	Active    string `json:"active"`
	Directory string `json:"directory"`
}

// InterfacesConfigDTO is the DTO for interface configuration.
type InterfacesConfigDTO struct {
	Telegram TelegramConfigDTO `json:"telegram"`
}

// TelegramConfigDTO is the DTO for Telegram configuration.
type TelegramConfigDTO struct {
	Enabled      bool     `json:"enabled"`
	BotToken     string   `json:"bot_token"`
	AllowedUsers []string `json:"allowed_users"`
}

// GetFacts returns all facts, optionally filtered by category.
func (a *App) GetFacts(category string) ([]FactDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	var facts []memory.Fact
	var err error
	if category == "" {
		facts, err = a.store.GetFacts(a.ctx, 100, 0)
	} else {
		facts, err = a.store.GetFactsByCategory(a.ctx, category)
	}
	if err != nil {
		return nil, fmt.Errorf("getting facts: %w", err)
	}

	dtos := make([]FactDTO, len(facts))
	for i, f := range facts {
		dtos[i] = FactDTO{
			ID:         f.ID,
			Fact:       f.Fact,
			Category:   f.Category,
			Confidence: f.Confidence,
			Source:     f.Source,
			CreatedAt:  f.CreatedAt,
			UpdatedAt:  f.UpdatedAt,
		}
	}
	return dtos, nil
}

// GetFact returns a single fact by its ID.
func (a *App) GetFact(id string) (*FactDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	f, err := a.store.GetFact(a.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting fact: %w", err)
	}
	if f == nil {
		return nil, fmt.Errorf("fact not found: %s", id)
	}

	return &FactDTO{
		ID:         f.ID,
		Fact:       f.Fact,
		Category:   f.Category,
		Confidence: f.Confidence,
		Source:     f.Source,
		CreatedAt:  f.CreatedAt,
		UpdatedAt:  f.UpdatedAt,
	}, nil
}

// UpdateFact updates an existing fact's text, category, and confidence.
func (a *App) UpdateFact(id, fact, category string, confidence float64) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}

	existing, err := a.store.GetFact(a.ctx, id)
	if err != nil {
		return fmt.Errorf("getting fact for update: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("fact not found: %s", id)
	}

	existing.Fact = fact
	existing.Category = category
	existing.Confidence = confidence
	existing.UpdatedAt = now()

	return a.store.UpdateFact(a.ctx, existing)
}

// DeleteFact removes a fact by its ID.
func (a *App) DeleteFact(id string) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}
	return a.store.DeleteFact(a.ctx, id)
}

// GetEpisodes returns a paginated list of episodes.
func (a *App) GetEpisodes(limit, offset int) ([]EpisodeDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	episodes, err := a.store.GetEpisodes(a.ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("getting episodes: %w", err)
	}

	dtos := make([]EpisodeDTO, len(episodes))
	for i, ep := range episodes {
		dtos[i] = EpisodeDTO{
			ID:         ep.ID,
			Summary:    ep.Summary,
			StartTime:  ep.StartTime,
			EndTime:    ep.EndTime,
			MessageIDs: ep.MessageIDs,
			Importance: ep.Importance,
			Topics:     ep.Topics,
		}
	}
	return dtos, nil
}

// GetEpisode returns a single episode by its ID.
func (a *App) GetEpisode(id string) (*EpisodeDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	ep, err := a.store.GetEpisode(a.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting episode: %w", err)
	}
	if ep == nil {
		return nil, fmt.Errorf("episode not found: %s", id)
	}

	return &EpisodeDTO{
		ID:         ep.ID,
		Summary:    ep.Summary,
		StartTime:  ep.StartTime,
		EndTime:    ep.EndTime,
		MessageIDs: ep.MessageIDs,
		Importance: ep.Importance,
		Topics:     ep.Topics,
	}, nil
}

// GetEntities returns all entities.
func (a *App) GetEntities() ([]EntityDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	entities, err := a.store.GetEntities(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting entities: %w", err)
	}

	dtos := make([]EntityDTO, len(entities))
	for i, e := range entities {
		dtos[i] = EntityDTO{
			ID:          e.ID,
			Name:        e.Name,
			Type:        e.Type,
			Description: e.Description,
			CreatedAt:   e.CreatedAt,
		}
	}
	return dtos, nil
}

// GetRelationships returns all relationships.
func (a *App) GetRelationships() ([]RelationshipDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	rels, err := a.store.GetRelationships(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting relationships: %w", err)
	}

	dtos := make([]RelationshipDTO, len(rels))
	for i, r := range rels {
		dtos[i] = RelationshipDTO{
			ID:           r.ID,
			SourceEntity: r.SourceEntity,
			TargetEntity: r.TargetEntity,
			Relationship: r.Relationship,
			Confidence:   r.Confidence,
			CreatedAt:    r.CreatedAt,
		}
	}
	return dtos, nil
}

// GetScratchpad returns the current scratchpad content.
func (a *App) GetScratchpad() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	return a.store.GetScratchpad(a.ctx)
}

// UpdateScratchpad replaces the scratchpad content.
func (a *App) UpdateScratchpad(content string) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}
	return a.store.UpdateScratchpad(a.ctx, content)
}

// SearchMemory performs a semantic or full-text search over facts and episodes.
func (a *App) SearchMemory(query, searchType string) (SearchResultsDTO, error) {
	if a.ctx == nil {
		return SearchResultsDTO{}, fmt.Errorf("app not started")
	}

	switch searchType {
	case "semantic":
		return a.semanticSearch(query)
	default:
		return a.fullTextSearch(query)
	}
}

// semanticSearch performs a vector similarity search using the configured embedder.
func (a *App) semanticSearch(query string) (SearchResultsDTO, error) {
	var results SearchResultsDTO

	provider, ok := a.cfg.Providers[a.cfg.DefaultProvider]
	if !ok {
		return results, fmt.Errorf("default provider %q not configured", a.cfg.DefaultProvider)
	}
	embedder := memory.NewEmbedder(provider.Endpoint, provider.EmbeddingModel)
	embedding, err := embedder.GenerateEmbedding(a.ctx, query)
	if err != nil {
		return results, fmt.Errorf("generating embedding: %w", err)
	}
	vec, err := memory.SerializeVector(embedding)
	if err != nil {
		return results, fmt.Errorf("serializing vector: %w", err)
	}

	facts, err := a.store.SearchFacts(a.ctx, vec, 10)
	if err != nil {
		return results, fmt.Errorf("searching facts: %w", err)
	}
	episodes, err := a.store.SearchEpisodes(a.ctx, vec, 10)
	if err != nil {
		return results, fmt.Errorf("searching episodes: %w", err)
	}

	results.Facts = make([]FactDTO, len(facts))
	for i, f := range facts {
		results.Facts[i] = factToDTO(&f)
	}
	results.Episodes = make([]EpisodeDTO, len(episodes))
	for i, ep := range episodes {
		results.Episodes[i] = episodeToDTO(&ep)
	}
	return results, nil
}

// fullTextSearch performs a substring match over facts and episodes.
func (a *App) fullTextSearch(query string) (SearchResultsDTO, error) {
	var results SearchResultsDTO
	pattern := "%" + query + "%"

	allFacts, err := a.store.GetFacts(a.ctx, 100, 0)
	if err != nil {
		return results, fmt.Errorf("getting facts for search: %w", err)
	}
	for _, f := range allFacts {
		if contains(f.Fact, pattern) || contains(f.Category, pattern) {
			results.Facts = append(results.Facts, factToDTO(&f))
		}
	}

	allEpisodes, err := a.store.GetEpisodes(a.ctx, 100, 0)
	if err != nil {
		return results, fmt.Errorf("getting episodes for search: %w", err)
	}
	for _, ep := range allEpisodes {
		if contains(ep.Summary, pattern) || contains(ep.Topics, pattern) {
			results.Episodes = append(results.Episodes, episodeToDTO(&ep))
		}
	}
	return results, nil
}

// GetTasks returns tasks filtered by status. Pass empty string for all tasks.
func (a *App) GetTasks(status string) ([]TaskDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	tasks, err := a.store.GetTasks(a.ctx, status)
	if err != nil {
		return nil, fmt.Errorf("getting tasks: %w", err)
	}

	dtos := make([]TaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = TaskDTO{
			ID:        t.ID,
			Type:      t.Type,
			Status:    t.Status,
			TriggerAt: t.TriggerAt,
			CronExpr:  t.CronExpr,
			Action:    t.Action,
			Context:   t.Context,
			CreatedAt: t.CreatedAt,
			FiredAt:   t.FiredAt,
		}
	}
	return dtos, nil
}

// CreateTask creates a new task (reminder or scheduled_message).
func (a *App) CreateTask(taskType, triggerAt, cronExpr, action, taskCtx string) (*TaskDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	// Parse triggerAt as Unix timestamp in milliseconds
	var triggerAtMs int64
	if triggerAt != "" {
		if _, err := fmt.Sscanf(triggerAt, "%d", &triggerAtMs); err != nil {
			return nil, fmt.Errorf("invalid trigger_at: %w", err)
		}
	}

	task := &memory.Task{
		ID:        uuid.NewString(),
		Type:      taskType,
		Status:    "pending",
		TriggerAt: triggerAtMs,
		CronExpr:  cronExpr,
		Action:    action,
		Context:   taskCtx,
		CreatedAt: now(),
	}

	if err := a.store.SaveTask(a.ctx, task); err != nil {
		return nil, fmt.Errorf("saving task: %w", err)
	}

	return &TaskDTO{
		ID:        task.ID,
		Type:      task.Type,
		Status:    task.Status,
		TriggerAt: task.TriggerAt,
		CronExpr:  task.CronExpr,
		Action:    task.Action,
		Context:   task.Context,
		CreatedAt: task.CreatedAt,
		FiredAt:   task.FiredAt,
	}, nil
}

// CancelTask cancels a task by its ID.
func (a *App) CancelTask(id string) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}
	return a.store.CancelTask(a.ctx, id)
}

// GetActivityLog returns paginated activity log entries, optionally filtered by type.
func (a *App) GetActivityLog(filter string, limit, offset int) ([]ActivityEntryDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}

	entries, err := a.store.GetActivityLog(a.ctx, filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("getting activity log: %w", err)
	}

	dtos := make([]ActivityEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = ActivityEntryDTO{
			ID:        e.ID,
			Timestamp: e.Timestamp,
			Type:      e.Type,
			Details:   e.Details,
			MessageID: e.MessageID,
			SessionID: e.SessionID,
		}
	}
	return dtos, nil
}

// GetConfig returns the current application configuration.
func (a *App) GetConfig() (*ConfigDTO, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("app not started")
	}
	return configToDTO(a.cfg), nil
}

// UpdateConfig saves a new configuration to disk and updates the in-memory config.
func (a *App) UpdateConfig(cfg *ConfigDTO) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}

	updated := dtoToConfig(cfg)
	if err := config.SaveConfig(a.cfgPath, updated); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	a.cfg = updated
	return nil
}

// GetAvailableModels fetches the list of available models from a provider endpoint.
func (a *App) GetAvailableModels(endpoint string) ([]string, error) {
	modelsURL := strings.TrimSuffix(endpoint, "/v1") + "/v1/models"
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequestWithContext(a.ctx, "GET", modelsURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching models: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	models := make([]string, len(modelsResp.Data))
	for i, m := range modelsResp.Data {
		models[i] = m.ID
	}
	return models, nil
}

func configToDTO(cfg *config.Config) *ConfigDTO {
	if cfg == nil {
		return nil
	}

	providers := make(map[string]ProviderConfigDTO)
	for name, p := range cfg.Providers {
		providers[name] = ProviderConfigDTO{
			Endpoint:       p.Endpoint,
			ChatModel:      p.ChatModel,
			EmbeddingModel: p.EmbeddingModel,
			Parameters:     p.Parameters,
		}
	}

	return &ConfigDTO{
		Providers:       providers,
		DefaultProvider: cfg.DefaultProvider,
		Memory: MemoryConfigDTO{
			DBPath:                    cfg.Memory.DBPath,
			WorkingMemoryTurns:        cfg.Memory.WorkingMemoryTurns,
			QuickConsolidationDelayMs: cfg.Memory.QuickConsolidationDelayMs,
			DeepConsolidationDelayMs:  cfg.Memory.DeepConsolidationDelayMs,
		},
		Persona: PersonaConfigDTO{
			Active:    cfg.Persona.Active,
			Directory: cfg.Persona.Directory,
		},
		Interfaces: InterfacesConfigDTO{
			Telegram: TelegramConfigDTO{
				Enabled:      cfg.Interfaces.Telegram.Enabled,
				BotToken:     cfg.Interfaces.Telegram.BotToken,
				AllowedUsers: cfg.Interfaces.Telegram.AllowedUsers,
			},
		},
	}
}

func dtoToConfig(dto *ConfigDTO) *config.Config {
	if dto == nil {
		return config.DefaultConfig()
	}

	providers := make(map[string]config.ProviderConfig)
	for name, p := range dto.Providers {
		providers[name] = config.ProviderConfig{
			Endpoint:       p.Endpoint,
			ChatModel:      p.ChatModel,
			EmbeddingModel: p.EmbeddingModel,
			Parameters:     p.Parameters,
		}
	}

	return &config.Config{
		Providers:       providers,
		DefaultProvider: dto.DefaultProvider,
		Memory: config.MemoryConfig{
			DBPath:                    dto.Memory.DBPath,
			WorkingMemoryTurns:        dto.Memory.WorkingMemoryTurns,
			QuickConsolidationDelayMs: dto.Memory.QuickConsolidationDelayMs,
			DeepConsolidationDelayMs:  dto.Memory.DeepConsolidationDelayMs,
		},
		Persona: config.PersonaConfig{
			Active:    dto.Persona.Active,
			Directory: dto.Persona.Directory,
		},
		Interfaces: config.InterfacesConfig{
			Telegram: config.TelegramConfig{
				Enabled:      dto.Interfaces.Telegram.Enabled,
				BotToken:     dto.Interfaces.Telegram.BotToken,
				AllowedUsers: dto.Interfaces.Telegram.AllowedUsers,
			},
		},
	}
}

func factToDTO(f *memory.Fact) FactDTO {
	return FactDTO{
		ID:         f.ID,
		Fact:       f.Fact,
		Category:   f.Category,
		Confidence: f.Confidence,
		Source:     f.Source,
		CreatedAt:  f.CreatedAt,
		UpdatedAt:  f.UpdatedAt,
	}
}

func episodeToDTO(ep *memory.Episode) EpisodeDTO {
	return EpisodeDTO{
		ID:         ep.ID,
		Summary:    ep.Summary,
		StartTime:  ep.StartTime,
		EndTime:    ep.EndTime,
		MessageIDs: ep.MessageIDs,
		Importance: ep.Importance,
		Topics:     ep.Topics,
	}
}

// now returns the current Unix timestamp in milliseconds.
func now() int64 {
	return time.Now().UnixMilli()
}

// contains checks if a string contains a LIKE pattern (simple substring match).
func contains(s, pattern string) bool {
	return len(pattern) > 2 && strings.Contains(s, pattern[1:len(pattern)-1])
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
