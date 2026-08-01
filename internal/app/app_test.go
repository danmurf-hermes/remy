package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danmurf/remy/internal/agent"
	"github.com/danmurf/remy/internal/app"
	"github.com/danmurf/remy/internal/config"
	"github.com/danmurf/remy/internal/memory"
	"github.com/danmurf/remy/internal/persona"
)

// fakeAgent implements app.AgentService for testing.
type fakeAgent struct {
	handleMessageFn         func(ctx context.Context, userMsg string) (*memory.Message, error)
	handleMessageStreamFn   func(ctx context.Context, userMsg string) (<-chan agent.StreamChunk, error)
	listPersonasFn          func(ctx context.Context) ([]persona.Summary, error)
	setActivePersonaFn      func(ctx context.Context, name string) error
	activePersonaFn         func() *persona.Persona
	loadActivePersonaFn     func(ctx context.Context) error
	scheduleConsolidationFn func(ctx context.Context) func()
}

func (f *fakeAgent) HandleMessage(ctx context.Context, userMsg string) (*memory.Message, error) {
	if f.handleMessageFn != nil {
		return f.handleMessageFn(ctx, userMsg)
	}
	return &memory.Message{ID: "msg-1", Role: "assistant", Content: "Hello!", Timestamp: time.Now().UnixMilli()}, nil
}

func (f *fakeAgent) HandleMessageStream(ctx context.Context, userMsg string) (<-chan agent.StreamChunk, error) {
	if f.handleMessageStreamFn != nil {
		return f.handleMessageStreamFn(ctx, userMsg)
	}
	ch := make(chan agent.StreamChunk, 2)
	ch <- agent.StreamChunk{Content: "Hello "}
	ch <- agent.StreamChunk{Content: "world!"}
	ch <- agent.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func (f *fakeAgent) ListPersonas(ctx context.Context) ([]persona.Summary, error) {
	if f.listPersonasFn != nil {
		return f.listPersonasFn(ctx)
	}
	return []persona.Summary{
		{Name: "default", Active: true},
		{Name: "creative", Active: false},
	}, nil
}

func (f *fakeAgent) SetActivePersona(ctx context.Context, name string) error {
	if f.setActivePersonaFn != nil {
		return f.setActivePersonaFn(ctx, name)
	}
	return nil
}

func (f *fakeAgent) ActivePersona() *persona.Persona {
	if f.activePersonaFn != nil {
		return f.activePersonaFn()
	}
	return &persona.Persona{Name: "default"}
}

func (f *fakeAgent) LoadActivePersona(ctx context.Context) error {
	if f.loadActivePersonaFn != nil {
		return f.loadActivePersonaFn(ctx)
	}
	return nil
}

func (f *fakeAgent) ScheduleConsolidation(ctx context.Context) func() {
	if f.scheduleConsolidationFn != nil {
		return f.scheduleConsolidationFn(ctx)
	}
	return func() {}
}

// recordingEmitter records emitted events for verification.
type recordingEmitter struct {
	events []emittedEvent
}

type emittedEvent struct {
	event string
	data  any
}

func (e *recordingEmitter) Emit(event string, data any) error {
	e.events = append(e.events, emittedEvent{event: event, data: data})
	return nil
}

func TestNewApp(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	cfg := config.DefaultConfig()
	appInst, err := app.NewApp(cfg, store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if appInst == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestSendMessage_NotStarted(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	_, err = appInst.SendMessage("hello")
	if err == nil {
		t.Fatal("expected error for unstarted app, got nil")
	}
}

func TestSendMessage_HappyPath(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	msg, err := appInst.SendMessage("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}
	if msg.Role != "assistant" {
		t.Errorf("role = %q, want %q", msg.Role, "assistant")
	}
	if msg.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestSendMessage_AgentError(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	agent := &fakeAgent{
		handleMessageFn: func(ctx context.Context, userMsg string) (*memory.Message, error) {
			return nil, errors.New("agent error")
		},
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, agent)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	_, err = appInst.SendMessage("hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSendMessageStream_NotStarted(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	err = appInst.SendMessageStream("hello")
	if err == nil {
		t.Fatal("expected error for unstarted app, got nil")
	}
}

func TestSendMessageStream_EmitsEvents(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	emitter := &recordingEmitter{}
	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.SetEmitter(emitter)
	appInst.Startup(context.Background())

	err = appInst.SendMessageStream("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for the goroutine to finish
	time.Sleep(100 * time.Millisecond)

	if len(emitter.events) == 0 {
		t.Fatal("expected at least one emitted event")
	}

	// Check we got chunk events and a done event
	gotChunk := false
	gotDone := false
	for _, e := range emitter.events {
		switch e.event {
		case "stream:chunk":
			gotChunk = true
		case "stream:done":
			gotDone = true
		}
	}
	if !gotChunk {
		t.Error("expected stream:chunk event")
	}
	if !gotDone {
		t.Error("expected stream:done event")
	}
}

func TestSendMessageStream_EmitsError(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	agent := &fakeAgent{
		handleMessageStreamFn: func(ctx context.Context, userMsg string) (<-chan agent.StreamChunk, error) {
			ch := make(chan agent.StreamChunk, 1)
			ch <- agent.StreamChunk{Error: "something went wrong"}
			close(ch)
			return ch, nil
		},
	}

	emitter := &recordingEmitter{}
	appInst, err := app.NewApp(config.DefaultConfig(), store, agent)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.SetEmitter(emitter)
	appInst.Startup(context.Background())

	err = appInst.SendMessageStream("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	gotError := false
	for _, e := range emitter.events {
		if e.event == "stream:error" {
			gotError = true
			break
		}
	}
	if !gotError {
		t.Error("expected stream:error event")
	}
}

func TestGetHistory_NotStarted(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	_, err = appInst.GetHistory(10, 0)
	if err == nil {
		t.Fatal("expected error for unstarted app, got nil")
	}
}

func TestGetHistory_Empty(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	msgs, err := appInst.GetHistory(10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("got %d messages, want 0", len(msgs))
	}
}

func TestGetHistory_WithMessages(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Insert a message directly into the store
	now := time.Now().UnixMilli()
	err = store.SaveMessage(context.Background(), &memory.Message{
		ID: "test-1", UserID: "u", Role: "user", Content: "hello", Timestamp: now, Interface: "gui", SessionID: "default",
	})
	if err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	msgs, err := appInst.GetHistory(10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Errorf("content = %q, want %q", msgs[0].Content, "hello")
	}
	if msgs[0].Role != "user" {
		t.Errorf("role = %q, want %q", msgs[0].Role, "user")
	}
}

func TestGetConversations_Empty(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	convs, err := appInst.GetConversations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(convs) != 0 {
		t.Errorf("got %d conversations, want 0", len(convs))
	}
}

func TestGetConversations_WithMessages(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	now := time.Now().UnixMilli()
	err = store.SaveMessage(context.Background(), &memory.Message{
		ID: "m1", UserID: "u", Role: "user", Content: "What's the weather?", Timestamp: now, Interface: "gui", SessionID: "default",
	})
	if err != nil {
		t.Fatalf("failed to save message: %v", err)
	}
	err = store.SaveMessage(context.Background(), &memory.Message{
		ID: "m2", UserID: "u", Role: "assistant", Content: "It's sunny!", Timestamp: now + 1, Interface: "gui", SessionID: "default",
	})
	if err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	convs, err := appInst.GetConversations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("got %d conversations, want 1", len(convs))
	}
	if convs[0].LastMsg != "It's sunny!" {
		t.Errorf("last_msg = %q, want %q", convs[0].LastMsg, "It's sunny!")
	}
}

func TestGetConversations_PreviewTruncation(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	longContent := ""
	for i := 0; i < 200; i++ {
		longContent += "a"
	}

	now := time.Now().UnixMilli()
	err = store.SaveMessage(context.Background(), &memory.Message{
		ID: "m1", UserID: "u", Role: "user", Content: longContent, Timestamp: now, Interface: "gui", SessionID: "default",
	})
	if err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	convs, err := appInst.GetConversations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("got %d conversations, want 1", len(convs))
	}
	if len(convs[0].LastMsg) != 103 {
		t.Errorf("preview length = %d, want 103 (100 + '...')", len(convs[0].LastMsg))
	}
}

func TestGetPersonas(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	personas, err := appInst.GetPersonas()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(personas) != 2 {
		t.Fatalf("got %d personas, want 2", len(personas))
	}
	if personas[0].Name != "default" {
		t.Errorf("name = %q, want %q", personas[0].Name, "default")
	}
	if !personas[0].IsActive {
		t.Error("expected default to be active")
	}
}

func TestGetPersonas_Error(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	agent := &fakeAgent{
		listPersonasFn: func(ctx context.Context) ([]persona.Summary, error) {
			return nil, errors.New("list error")
		},
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, agent)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	_, err = appInst.GetPersonas()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSwitchPersona(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	err = appInst.SwitchPersona("creative")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSwitchPersona_NotStarted(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	err = appInst.SwitchPersona("creative")
	if err == nil {
		t.Fatal("expected error for unstarted app, got nil")
	}
}

func TestGetActivePersona_Default(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	appInst, err := app.NewApp(config.DefaultConfig(), store, &fakeAgent{})
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	name := appInst.GetActivePersona()
	if name != "default" {
		t.Errorf("got %q, want %q", name, "default")
	}
}

func TestGetActivePersona_Nil(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	agent := &fakeAgent{
		activePersonaFn: func() *persona.Persona {
			return nil
		},
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, agent)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	name := appInst.GetActivePersona()
	if name != "default" {
		t.Errorf("got %q, want %q", name, "default")
	}
}

func TestMessageToDTO_Nil(t *testing.T) {
	// messageToDTO is unexported, so we test it indirectly via SendMessage
	// with a nil-returning agent
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	agent := &fakeAgent{
		handleMessageFn: func(ctx context.Context, userMsg string) (*memory.Message, error) {
			return nil, nil //nolint:nilnil // intentional: test nil-handling path
		},
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, agent)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	msg, err := appInst.SendMessage("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Error("expected nil message for nil agent response")
	}
}

func TestStartup_LoadsPersona(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	loaded := false
	agent := &fakeAgent{
		loadActivePersonaFn: func(ctx context.Context) error {
			loaded = true
			return nil
		},
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, agent)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	appInst.Startup(context.Background())

	if !loaded {
		t.Error("expected LoadActivePersona to be called during startup")
	}
}

func TestStartup_LoadPersonaError(t *testing.T) {
	store, err := memory.NewStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	agent := &fakeAgent{
		loadActivePersonaFn: func(ctx context.Context) error {
			return errors.New("load error")
		},
	}

	appInst, err := app.NewApp(config.DefaultConfig(), store, agent)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	// Should not panic on persona load error
	appInst.Startup(context.Background())
}
