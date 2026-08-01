// Package telegram tests for the Telegram bot interface.
package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/memory"
)

// mockAgentService implements AgentService for testing.
type mockAgentService struct {
	mu          sync.Mutex
	handleFunc  func(ctx context.Context, userMsg string) (*memory.Message, error)
	callCount   int
	lastMessage string
}

func (m *mockAgentService) HandleMessage(ctx context.Context, userMsg string) (*memory.Message, error) {
	m.mu.Lock()
	m.callCount++
	m.lastMessage = userMsg
	fn := m.handleFunc
	m.mu.Unlock()

	if fn != nil {
		return fn(ctx, userMsg)
	}
	return &memory.Message{
		ID:      "test-response-1",
		Role:    "assistant",
		Content: "Hello from Remy!",
	}, nil
}

func (m *mockAgentService) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func (m *mockAgentService) LastMessage() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastMessage
}

// mockStore implements Store interface for testing.
type mockStore struct {
	mu       sync.Mutex
	messages []memory.Message
}

func (s *mockStore) SaveMessage(ctx context.Context, msg *memory.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, *msg)
	return nil
}

// telegramAPIServer creates a test HTTP server that mimics the Telegram Bot API.
// It returns the server, a channel that receives sent messages, and a bot
// configured to use the test server.
func telegramAPIServer(t *testing.T) (*httptest.Server, chan string, *tgbotapi.BotAPI) {
	t.Helper()
	sentCh := make(chan string, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Handle getMe
		if strings.HasSuffix(path, "/getMe") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":12345,"is_bot":true,"first_name":"TestBot","username":"test_bot"}}`))
			return
		}

		// Handle getUpdates - return empty updates
		if strings.HasSuffix(path, "/getUpdates") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"result":[]}`))
			return
		}

		// Handle sendMessage
		if strings.HasSuffix(path, "/sendMessage") {
			if err := r.ParseForm(); err == nil {
				text := r.Form.Get("text")
				sentCh <- text
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1,"chat":{"id":123},"text":"ok"}}`))
			return
		}

		// Handle sendChatAction
		if strings.HasSuffix(path, "/sendChatAction") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":0,"chat":{"id":123},"text":""}}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"ok":false,"error_code":404,"description":"Not Found"}`))
	}))

	// Create a bot that routes through our test server
	apiEndpoint := server.URL + "/bot%s/%s"
	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint("test:token", apiEndpoint)
	if err != nil {
		server.Close()
		t.Fatalf("creating test bot: %v", err)
	}

	return server, sentCh, bot
}

func TestNew(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	if tg == nil {
		t.Fatal("New returned nil")
	}
	if tg.agent != agent {
		t.Error("agent not set")
	}
	if tg.store != store {
		t.Error("store not set")
	}
	if tg.cfg != cfg {
		t.Error("cfg not set")
	}
}

func TestStart_EmptyToken(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "",
	}

	tg := New(agent, store, cfg)
	err := tg.Start(context.Background())
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' error, got: %v", err)
	}
}

func TestStart_DoubleStart(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.started = true

	err := tg.Start(context.Background())
	if err == nil {
		t.Fatal("expected error for double start")
	}
	if !strings.Contains(err.Error(), "already started") {
		t.Errorf("expected 'already started' error, got: %v", err)
	}
}

func TestSend(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	err := tg.Send(12345, "Hello, world!")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case sent := <-sentCh:
		if sent != "Hello, world!" {
			t.Errorf("expected 'Hello, world!', got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for sent message")
	}
}

func TestSend_NoBot(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	err := tg.Send(12345, "test")
	if err == nil {
		t.Fatal("expected error when bot not initialized")
	}
}

func TestStop_Idempotent(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	// Stop without starting should not panic
	tg.Stop()
	// Second stop should also not panic
	tg.Stop()
}

func TestIsUserAllowed_EmptyList(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:      true,
		BotToken:     "test:token",
		AllowedUsers: []string{},
	}

	tg := New(agent, store, cfg)
	if !tg.isUserAllowed(12345) {
		t.Error("expected all users allowed when list is empty")
	}
}

func TestIsUserAllowed_WithList(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:      true,
		BotToken:     "test:token",
		AllowedUsers: []string{"12345", "67890"},
	}

	tg := New(agent, store, cfg)

	tests := []struct {
		name    string
		userID  int64
		allowed bool
	}{
		{"allowed user 1", 12345, true},
		{"allowed user 2", 67890, true},
		{"blocked user", 11111, false},
		{"another blocked", 99999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tg.isUserAllowed(tt.userID)
			if got != tt.allowed {
				t.Errorf("isUserAllowed(%d) = %v, want %v", tt.userID, got, tt.allowed)
			}
		})
	}
}

func TestHandleCommand_Start(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	tg.handleCommand(context.Background(), 12345, 67890, "start", "")

	select {
	case sent := <-sentCh:
		if !strings.Contains(sent, "Hello") {
			t.Errorf("expected greeting, got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for response")
	}
}

func TestHandleCommand_Help(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	tg.handleCommand(context.Background(), 12345, 67890, "help", "")

	select {
	case sent := <-sentCh:
		if !strings.Contains(sent, "Commands") {
			t.Errorf("expected help text, got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for response")
	}
}

func TestHandleCommand_Status(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	tg.handleCommand(context.Background(), 12345, 67890, "status", "")

	select {
	case sent := <-sentCh:
		if !strings.Contains(sent, "online") {
			t.Errorf("expected status text, got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for response")
	}
}

func TestHandleCommand_Unknown(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	tg.handleCommand(context.Background(), 12345, 67890, "unknown", "")

	select {
	case sent := <-sentCh:
		if !strings.Contains(sent, "Unknown command") {
			t.Errorf("expected unknown command message, got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for response")
	}
}

func TestHandleUpdate_AgentError(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{
		handleFunc: func(ctx context.Context, userMsg string) (*memory.Message, error) {
			return nil, fmt.Errorf("agent error")
		},
	}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 1,
			Text:      "Hello",
			Chat:      &tgbotapi.Chat{ID: 12345},
			From:      &tgbotapi.User{ID: 67890},
		},
	}

	tg.handleUpdate(context.Background(), &update)

	select {
	case sent := <-sentCh:
		if !strings.Contains(sent, "error") {
			t.Errorf("expected error message, got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for error response")
	}
}

func TestHandleUpdate_EmptyMessage(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)

	// Empty text should not trigger agent call
	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 1,
			Text:      "",
			Chat:      &tgbotapi.Chat{ID: 12345},
			From:      &tgbotapi.User{ID: 67890},
		},
	}

	tg.handleUpdate(context.Background(), &update)

	if agent.CallCount() > 0 {
		t.Error("agent should not be called for empty message")
	}
}

func TestHandleUpdate_UnauthorizedUser(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:      true,
		BotToken:     "test:token",
		AllowedUsers: []string{"12345"},
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	// User 99999 is not in the allowed list
	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 1,
			Text:      "Hello",
			Chat:      &tgbotapi.Chat{ID: 99999},
			From:      &tgbotapi.User{ID: 99999},
		},
	}

	tg.handleUpdate(context.Background(), &update)

	if agent.CallCount() > 0 {
		t.Error("agent should not be called for unauthorized user")
	}

	select {
	case sent := <-sentCh:
		if !strings.Contains(sent, "not authorized") {
			t.Errorf("expected 'not authorized' message, got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for unauthorized response")
	}
}

func TestHandleUpdate_Command(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	// Simulate a /start command message
	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 1,
			Text:      "/start",
			Chat:      &tgbotapi.Chat{ID: 12345},
			From:      &tgbotapi.User{ID: 67890},
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 6},
			},
		},
	}

	tg.handleUpdate(context.Background(), &update)

	// Commands should not go through the agent
	if agent.CallCount() > 0 {
		t.Error("agent should not be called for commands")
	}

	select {
	case sent := <-sentCh:
		if !strings.Contains(sent, "Hello") {
			t.Errorf("expected greeting for /start, got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for command response")
	}
}

func TestHandleUpdate_NormalMessage(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{
		handleFunc: func(ctx context.Context, userMsg string) (*memory.Message, error) {
			return &memory.Message{
				ID:      "resp-1",
				Role:    "assistant",
				Content: "I heard you say: " + userMsg,
			}, nil
		},
	}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 1,
			Text:      "What's the weather?",
			Chat:      &tgbotapi.Chat{ID: 12345},
			From:      &tgbotapi.User{ID: 67890},
		},
	}

	tg.handleUpdate(context.Background(), &update)

	if agent.CallCount() != 1 {
		t.Errorf("expected 1 agent call, got %d", agent.CallCount())
	}
	if agent.LastMessage() != "What's the weather?" {
		t.Errorf("expected 'What's the weather?', got: %s", agent.LastMessage())
	}

	select {
	case sent := <-sentCh:
		if !strings.Contains(sent, "I heard you say") {
			t.Errorf("expected response with echo, got: %s", sent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for response")
	}
}

func TestHandleUpdate_NilMessage(t *testing.T) {
	agent := &mockAgentService{}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)

	// Update with nil Message should be ignored (no panic)
	update := tgbotapi.Update{
		UpdateID: 1,
		Message:  nil,
	}

	tg.handleUpdate(context.Background(), &update)

	if agent.CallCount() > 0 {
		t.Error("agent should not be called for nil message")
	}
}

func TestHandleUpdate_EmptyResponse(t *testing.T) {
	server, sentCh, bot := telegramAPIServer(t)
	defer server.Close()

	agent := &mockAgentService{
		handleFunc: func(ctx context.Context, userMsg string) (*memory.Message, error) {
			return &memory.Message{ID: "resp-empty", Role: "assistant", Content: ""}, nil
		},
	}
	store := &mockStore{}
	cfg := &config.TelegramConfig{
		Enabled:  true,
		BotToken: "test:token",
	}

	tg := New(agent, store, cfg)
	tg.bot = bot

	update := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 1,
			Text:      "Hello",
			Chat:      &tgbotapi.Chat{ID: 12345},
			From:      &tgbotapi.User{ID: 67890},
		},
	}

	tg.handleUpdate(context.Background(), &update)

	// Empty response should not send anything
	select {
	case <-sentCh:
		t.Error("should not send message for empty response")
	case <-time.After(500 * time.Millisecond):
		// Expected - no message sent
	}
}
