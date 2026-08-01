package agent

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/remy/internal/llm"
	"github.com/yourname/remy/internal/memory"
)

// mockStore implements the Store interface for testing.
type mockStore struct {
	mu               sync.RWMutex
	messages         map[string]*memory.Message
	episodes         []memory.Episode
	facts            []memory.Fact
	scratchpad       string
	activity         []memory.ActivityEntry
	messageVectors   map[string][]byte
	saveMessageErr   error
	getMessagesErr   error
	searchEpisodesErr error
	searchFactsErr   error
	getScratchpadErr error
	updateScratchpadErr error
	logActivityErr   error
	saveVectorErr    error
}

func newMockStore() *mockStore {
	return &mockStore{
		messages:       make(map[string]*memory.Message),
		messageVectors: make(map[string][]byte),
	}
}

func (m *mockStore) SaveMessage(_ context.Context, msg *memory.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveMessageErr != nil {
		return m.saveMessageErr
	}
	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	m.messages[msg.ID] = msg
	return nil
}

func (m *mockStore) GetMessage(_ context.Context, id string) (*memory.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	msg, ok := m.messages[id]
	if !ok {
		return nil, errors.New("message not found")
	}
	return msg, nil
}

func (m *mockStore) GetMessages(_ context.Context, limit, offset int) ([]memory.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getMessagesErr != nil {
		return nil, m.getMessagesErr
	}
	var all []memory.Message
	for _, msg := range m.messages {
		all = append(all, *msg)
	}
	// Sort by timestamp descending (most recent first)
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].Timestamp > all[i].Timestamp {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	start := offset
	if start > len(all) {
		start = len(all)
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], nil
}

func (m *mockStore) GetMessagesBySession(_ context.Context, sessionID string) ([]memory.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []memory.Message
	for _, msg := range m.messages {
		if msg.SessionID == sessionID {
			result = append(result, *msg)
		}
	}
	return result, nil
}

func (m *mockStore) SearchEpisodes(_ context.Context, _ []byte, _ int) ([]memory.Episode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.searchEpisodesErr != nil {
		return nil, m.searchEpisodesErr
	}
	return m.episodes, nil
}

func (m *mockStore) SearchFacts(_ context.Context, _ []byte, _ int) ([]memory.Fact, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.searchFactsErr != nil {
		return nil, m.searchFactsErr
	}
	return m.facts, nil
}

func (m *mockStore) GetScratchpad(_ context.Context) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getScratchpadErr != nil {
		return "", m.getScratchpadErr
	}
	return m.scratchpad, nil
}

func (m *mockStore) UpdateScratchpad(_ context.Context, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.updateScratchpadErr != nil {
		return m.updateScratchpadErr
	}
	m.scratchpad = content
	return nil
}

func (m *mockStore) LogActivity(_ context.Context, entry *memory.ActivityEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.logActivityErr != nil {
		return m.logActivityErr
	}
	m.activity = append(m.activity, *entry)
	return nil
}

func (m *mockStore) SaveMessageVector(_ context.Context, messageID string, embedding []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveVectorErr != nil {
		return m.saveVectorErr
	}
	m.messageVectors[messageID] = embedding
	return nil
}

// mockProvider implements the llm.Provider interface for testing.
type mockProvider struct {
	chatResponse *llm.ChatResponse
	chatErr      error
	embedResult  []float32
	embedErr     error
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		chatResponse: &llm.ChatResponse{
			ID:     "mock-chat-id",
			Object: "chat.completion",
			Choices: []llm.Choice{
				{
					Index: 0,
					Message: llm.Message{
						Role:    "assistant",
						Content: "Hello! How can I help you today?",
					},
					FinishReason: "stop",
				},
			},
			Usage: llm.Usage{
				PromptTokens:     10,
				CompletionTokens: 8,
				TotalTokens:      18,
			},
		},
		embedResult: make([]float32, 768),
	}
}

func (m *mockProvider) Chat(_ context.Context, _ llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.chatErr != nil {
		return nil, m.chatErr
	}
	return m.chatResponse, nil
}

func (m *mockProvider) ChatStream(_ context.Context, _ llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not implemented")
}

func (m *mockProvider) Embed(_ context.Context, _ string) ([]float32, error) {
	if m.embedErr != nil {
		return nil, m.embedErr
	}
	return m.embedResult, nil
}

// mockEmbedder implements a simple embedder for testing.
type mockEmbedder struct {
	embedResult []float32
	embedErr    error
}

func newMockEmbedder() *mockEmbedder {
	return &mockEmbedder{
		embedResult: make([]float32, 768),
	}
}

func (m *mockEmbedder) GenerateEmbedding(_ context.Context, _ string) ([]float32, error) {
	if m.embedErr != nil {
		return nil, m.embedErr
	}
	return m.embedResult, nil
}

func newTestAgent(store Store, provider llm.Provider, embedder Embedder) *Agent {
	cfg := DefaultConfig()
	cfg.WorkingMemoryTurns = 20
	return NewAgent(store, provider, embedder, cfg)
}

func newTestAgentWithConfig(store Store, provider llm.Provider, embedder Embedder, cfg Config) *Agent {
	return NewAgent(store, provider, embedder, cfg)
}

// TestAgent_NewAgent verifies the agent is created with the correct config.
func TestAgent_NewAgent(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()
	cfg := DefaultConfig()

	agent := NewAgent(store, provider, embedder, cfg)
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
}

// TestAgent_HandleMessage_NormalFlow verifies the full message handling pipeline.
func TestAgent_HandleMessage_NormalFlow(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()
	agent := newTestAgent(store, provider, embedder)

	response, err := agent.HandleMessage(context.Background(), "Hello, Remy!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("expected non-nil response")
	}
	if response.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", response.Role)
	}
	if response.Content != "Hello! How can I help you today?" {
		t.Errorf("unexpected response content: %q", response.Content)
	}
	if response.ID == "" {
		t.Error("expected non-empty response ID")
	}
	if response.Timestamp == 0 {
		t.Error("expected non-zero timestamp")
	}

	// Verify user message was stored
	messages, err := store.GetMessages(context.Background(), 100, 0)
	if err != nil {
		t.Fatalf("unexpected error getting messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages (user + assistant), got %d", len(messages))
	}

	var userMsg, assistantMsg *memory.Message
	for i := range messages {
		switch messages[i].Role {
		case "user":
			userMsg = &messages[i]
		case "assistant":
			assistantMsg = &messages[i]
		}
	}
	if userMsg == nil {
		t.Fatal("expected a user message")
	}
	if userMsg.Content != "Hello, Remy!" {
		t.Errorf("unexpected user message content: %q", userMsg.Content)
	}
	if assistantMsg == nil {
		t.Fatal("expected an assistant message")
	}
	if assistantMsg.Role != "assistant" {
		t.Errorf("expected assistant message role 'assistant', got %q", assistantMsg.Role)
	}

	// Verify vectors were stored
	if len(store.messageVectors) != 2 {
		t.Errorf("expected 2 message vectors, got %d", len(store.messageVectors))
	}

	// Verify activity was logged
	if len(store.activity) != 2 {
		t.Errorf("expected 2 activity entries, got %d", len(store.activity))
	}
}

// TestAgent_HandleMessage_WithContext verifies context retrieval is included.
func TestAgent_HandleMessage_WithContext(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	// Pre-populate episodes and facts
	store.episodes = []memory.Episode{
		{
			ID:         uuid.NewString(),
			Summary:    "User asked about Go programming",
			StartTime:  time.Now().Add(-1 * time.Hour).UnixMilli(),
			EndTime:    time.Now().Add(-30 * time.Minute).UnixMilli(),
			Importance: 0.8,
		},
	}
	store.facts = []memory.Fact{
		{
			ID:         uuid.NewString(),
			Fact:       "User prefers Go for backend development",
			Category:   "preference",
			Confidence: 0.9,
		},
	}
	store.scratchpad = "User is working on a CLI tool project"

	agent := newTestAgent(store, provider, embedder)

	response, err := agent.HandleMessage(context.Background(), "What do you think about Rust?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("expected non-nil response")
	}
	if response.Content == "" {
		t.Error("expected non-empty response content")
	}
}

// TestAgent_HandleMessage_EmptyMessage verifies handling of empty messages.
func TestAgent_HandleMessage_EmptyMessage(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()
	agent := newTestAgent(store, provider, embedder)

	response, err := agent.HandleMessage(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("expected non-nil response")
	}
}

// TestAgent_HandleMessage_LLMError verifies graceful handling of LLM failures.
func TestAgent_HandleMessage_LLMError(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	provider.chatErr = errors.New("LLM is down")

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from LLM failure")
	}
}

// TestAgent_HandleMessage_StoreSaveError verifies handling of store save failures.
func TestAgent_HandleMessage_StoreSaveError(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	store.saveMessageErr = errors.New("database is locked")

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from store save failure")
	}
}

// TestAgent_HandleMessage_EmbeddingError verifies handling of embedding failures.
func TestAgent_HandleMessage_EmbeddingError(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	embedder.embedErr = errors.New("embedding service unavailable")

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from embedding failure")
	}
}

// TestAgent_HandleMessage_EmptyLLMResponse verifies handling when LLM returns no choices.
func TestAgent_HandleMessage_EmptyLLMResponse(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	provider.chatResponse = &llm.ChatResponse{
		ID:      "mock-id",
		Object:  "chat.completion",
		Choices: []llm.Choice{},
	}

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from empty LLM response")
	}
}

// TestAgent_HandleMessage_WithCustomConfig verifies custom config is used.
func TestAgent_HandleMessage_WithCustomConfig(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	cfg := Config{
		WorkingMemoryTurns: 5,
		UserID:             "test-user",
		SessionID:          "test-session",
		Interface:          "telegram",
	}

	agent := newTestAgentWithConfig(store, provider, embedder, cfg)

	response, err := agent.HandleMessage(context.Background(), "Hello from Telegram!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.Interface != "telegram" {
		t.Errorf("expected interface 'telegram', got %q", response.Interface)
	}
	if response.SessionID != "test-session" {
		t.Errorf("expected session 'test-session', got %q", response.SessionID)
	}
}

// TestAgent_HandleMessage_ScratchpadError verifies scratchpad errors are handled.
func TestAgent_HandleMessage_ScratchpadError(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	store.getScratchpadErr = errors.New("scratchpad read error")

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from scratchpad failure")
	}
}

// TestAgent_HandleMessage_SearchEpisodesError verifies episode search errors are handled.
func TestAgent_HandleMessage_SearchEpisodesError(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	store.searchEpisodesErr = errors.New("episode search failed")

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from episode search failure")
	}
}

// TestAgent_HandleMessage_SearchFactsError verifies fact search errors are handled.
func TestAgent_HandleMessage_SearchFactsError(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	store.searchFactsErr = errors.New("fact search failed")

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from fact search failure")
	}
}

// TestAgent_HandleMessage_GetMessagesError verifies message retrieval errors are handled.
func TestAgent_HandleMessage_GetMessagesError(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	store.getMessagesErr = errors.New("message retrieval failed")

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from message retrieval failure")
	}
}

// TestAgent_HandleMessage_SaveVectorError verifies vector save errors are handled.
func TestAgent_HandleMessage_SaveVectorError(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()

	store.saveVectorErr = errors.New("vector save failed")

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from vector save failure")
	}
}

// TestAgent_HandleMessage_ConcurrentMessages verifies the agent handles
// multiple messages in sequence correctly.
func TestAgent_HandleMessage_ConcurrentMessages(t *testing.T) {
	store := newMockStore()
	provider := newMockProvider()
	embedder := newMockEmbedder()
	agent := newTestAgent(store, provider, embedder)

	// Send two messages in sequence
	resp1, err := agent.HandleMessage(context.Background(), "First message")
	if err != nil {
		t.Fatalf("unexpected error on first message: %v", err)
	}
	if resp1 == nil {
		t.Fatal("expected non-nil response for first message")
	}

	resp2, err := agent.HandleMessage(context.Background(), "Second message")
	if err != nil {
		t.Fatalf("unexpected error on second message: %v", err)
	}
	if resp2 == nil {
		t.Fatal("expected non-nil response for second message")
	}

	// Verify all 4 messages are stored (2 user + 2 assistant)
	messages, err := store.GetMessages(context.Background(), 100, 0)
	if err != nil {
		t.Fatalf("unexpected error getting messages: %v", err)
	}
	if len(messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(messages))
	}

	// Verify 4 vectors stored
	if len(store.messageVectors) != 4 {
		t.Errorf("expected 4 message vectors, got %d", len(store.messageVectors))
	}

	// Verify 4 activity entries
	if len(store.activity) != 4 {
		t.Errorf("expected 4 activity entries, got %d", len(store.activity))
	}
}

// TestDefaultConfig verifies default config values.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.WorkingMemoryTurns != 20 {
		t.Errorf("expected WorkingMemoryTurns=20, got %d", cfg.WorkingMemoryTurns)
	}
	if cfg.UserID != "default-user" {
		t.Errorf("expected UserID='default-user', got %q", cfg.UserID)
	}
	if cfg.SessionID != "default" {
		t.Errorf("expected SessionID='default', got %q", cfg.SessionID)
	}
	if cfg.Interface != "gui" {
		t.Errorf("expected Interface='gui', got %q", cfg.Interface)
	}
}
