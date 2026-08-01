package agent_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/danmurf/remy/internal/agent"
	"github.com/danmurf/remy/internal/agent/mock_agent"
	"github.com/danmurf/remy/internal/llm"
	"github.com/danmurf/remy/internal/llm/mock_llm"
	"github.com/danmurf/remy/internal/memory"
	"github.com/danmurf/remy/internal/persona"
)

func TestAgent_NewAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)
	cfg := agent.DefaultConfig()

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	if a == nil {
		t.Fatal("expected non-nil agent")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := agent.DefaultConfig()
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
	if cfg.ActivePersona != "default" {
		t.Errorf("expected ActivePersona='default', got %q", cfg.ActivePersona)
	}
}

type handleMessageMock struct {
	store         *mock_agent.MockStore
	provider      *mock_llm.MockProvider
	embedder      *mock_agent.MockEmbedder
	personaLoader *mock_agent.MockPersonaLoader
	scheduler     *mock_agent.MockScheduler
}

func TestAgent_HandleMessage(t *testing.T) {
	tests := []struct {
		name       string
		userMsg    string
		cfg        agent.Config
		mock       func(m handleMessageMock)
		wantErr    bool
		wantRole   string
		wantPrefix string
	}{
		{
			name:    "normal flow",
			userMsg: "Hello, Remy!",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello, Remy!").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-chat-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Hello! How can I help you today?"}, FinishReason: "stop"}},
					Usage:   llm.Usage{PromptTokens: 10, CompletionTokens: 8, TotalTokens: 18},
				}, nil)
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantRole:   "assistant",
			wantPrefix: "Hello! How can I help you today?",
		},
		{
			name:    "with context",
			userMsg: "What do you think about Rust?",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "What do you think about Rust?").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "User asked about Go programming", StartTime: time.Now().Add(-1 * time.Hour).UnixMilli(), EndTime: time.Now().Add(-30 * time.Minute).UnixMilli(), Importance: 0.8},
				}, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return([]memory.Fact{
					{ID: uuid.NewString(), Fact: "User prefers Go for backend development", Category: "preference", Confidence: 0.9},
				}, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("User is working on a CLI tool project", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-chat-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Rust is great for systems programming!"}, FinishReason: "stop"}},
				}, nil)
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantRole:   "assistant",
			wantPrefix: "Rust is great",
		},
		{
			name:    "empty message",
			userMsg: "",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: ""}, FinishReason: "stop"}},
				}, nil)
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantRole: "assistant",
		},
		{
			name:    "LLM error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(nil, errors.New("LLM is down"))
			},
			wantErr: true,
		},
		{
			name:    "store save error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Return(errors.New("database is locked"))
			},
			wantErr: true,
		},
		{
			name:    "embedding error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(nil, errors.New("embedding service unavailable"))
			},
			wantErr: true,
		},
		{
			name:    "empty LLM response",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{},
				}, nil)
			},
			wantErr: true,
		},
		{
			name:    "custom config",
			userMsg: "Hello from Telegram!",
			cfg:     agent.Config{WorkingMemoryTurns: 5, UserID: "test-user", SessionID: "test-session", Interface: "telegram"},
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello from Telegram!").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 5, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"}},
				}, nil)
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantRole:   "assistant",
			wantPrefix: "Hello!",
		},
		{
			name:    "scratchpad error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", errors.New("scratchpad read error"))
			},
			wantErr: true,
		},
		{
			name:    "episode search error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, errors.New("episode search failed"))
			},
			wantErr: true,
		},
		{
			name:    "fact search error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, errors.New("fact search failed"))
			},
			wantErr: true,
		},
		{
			name:    "get messages error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, errors.New("message retrieval failed"))
			},
			wantErr: true,
		},
		{
			name:    "vector save error on user message",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("vector save failed"))
			},
			wantErr: true,
		},
		{
			name:    "vector save error on response",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "response"}, FinishReason: "stop"}},
				}, nil)
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("vector save failed on response"))
			},
			wantErr: true,
		},
		{
			name:    "save response message error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "response"}, FinishReason: "stop"}},
				}, nil)
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Return(errors.New("save response failed"))
			},
			wantErr: true,
		},
		{
			name:    "response embedding error",
			userMsg: "Hello",
			mock: func(m handleMessageMock) {
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
				m.store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
				m.store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
				m.store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
				m.scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
				m.provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "response"}, FinishReason: "stop"}},
				}, nil)
				m.store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
					if msg.ID == "" {
						msg.ID = uuid.NewString()
					}
				})
				m.store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
				m.embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(nil, errors.New("response embedding failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := handleMessageMock{
				store:         mock_agent.NewMockStore(ctrl),
				provider:      mock_llm.NewMockProvider(ctrl),
				embedder:      mock_agent.NewMockEmbedder(ctrl),
				personaLoader: mock_agent.NewMockPersonaLoader(ctrl),
				scheduler:     mock_agent.NewMockScheduler(ctrl),
			}

			tt.mock(m)

			cfg := agent.DefaultConfig()
			if tt.cfg != (agent.Config{}) {
				cfg = tt.cfg
			}
			a := agent.NewAgent(m.store, m.provider, m.embedder, m.personaLoader, m.scheduler, &cfg)

			resp, err := a.HandleMessage(context.Background(), tt.userMsg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.ID == "" {
				t.Error("expected non-empty response ID")
			}
			if resp.Timestamp == 0 {
				t.Error("expected non-zero timestamp")
			}
			if tt.wantRole != "" && resp.Role != tt.wantRole {
				t.Errorf("role = %q, want %q", resp.Role, tt.wantRole)
			}
			if tt.wantPrefix != "" && resp.Content[:min(len(resp.Content), len(tt.wantPrefix))] != tt.wantPrefix {
				t.Errorf("content prefix = %q, want %q", resp.Content[:min(len(resp.Content), len(tt.wantPrefix))], tt.wantPrefix)
			}
		})
	}
}

func TestAgent_HandleMessage_ConcurrentMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "First message").Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
	scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "First response"}, FinishReason: "stop"}},
	}, nil)
	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Second message").Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
	scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)
	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Second response"}, FinishReason: "stop"}},
	}, nil)
	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	cfg := agent.DefaultConfig()
	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)

	resp1, err := a.HandleMessage(context.Background(), "First message")
	if err != nil {
		t.Fatalf("unexpected error on first message: %v", err)
	}
	if resp1 == nil {
		t.Fatal("expected non-nil response for first message")
	}

	resp2, err := a.HandleMessage(context.Background(), "Second message")
	if err != nil {
		t.Fatalf("unexpected error on second message: %v", err)
	}
	if resp2 == nil {
		t.Fatal("expected non-nil response for second message")
	}
}

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name     string
		input    *agent.PromptInput
		wantN    int
		wantLast string
	}{
		{
			name:     "no context",
			input:    &agent.PromptInput{UserMessage: "hello"},
			wantN:    2,
			wantLast: "hello",
		},
		{
			name:     "with scratchpad",
			input:    &agent.PromptInput{Scratchpad: "Remember: user likes Go", UserMessage: "hello"},
			wantN:    2,
			wantLast: "hello",
		},
		{
			name:     "with episodes",
			input:    &agent.PromptInput{Episodes: []memory.Episode{{Summary: "User asked about Go", Importance: 0.8}}, UserMessage: "hello"},
			wantN:    2,
			wantLast: "hello",
		},
		{
			name:     "with facts",
			input:    &agent.PromptInput{Facts: []memory.Fact{{Fact: "User prefers Go", Category: "preference", Confidence: 0.9}}, UserMessage: "hello"},
			wantN:    2,
			wantLast: "hello",
		},
		{
			name: "with recent messages",
			input: &agent.PromptInput{
				RecentMessages: []memory.Message{
					{Role: "user", Content: "prev question"},
					{Role: "assistant", Content: "prev answer"},
				},
				UserMessage: "hello",
			},
			wantN:    4,
			wantLast: "hello",
		},
		{
			name: "all context types",
			input: &agent.PromptInput{
				Scratchpad: "working notes",
				Episodes:   []memory.Episode{{Summary: "past convo", Importance: 0.7}},
				Facts:      []memory.Fact{{Fact: "user fact", Category: "general", Confidence: 0.8}},
				RecentMessages: []memory.Message{
					{Role: "user", Content: "prev"},
				},
				UserMessage: "hello",
			},
			wantN:    3,
			wantLast: "hello",
		},
		{
			name: "with persona body",
			input: &agent.PromptInput{
				Persona:     &persona.Persona{Name: "creative", Body: "# Creative\n\nYou are a creative assistant."},
				UserMessage: "hello",
			},
			wantN:    2,
			wantLast: "hello",
		},
		{
			name: "with upcoming tasks",
			input: &agent.PromptInput{
				UpcomingTasks: "- [reminder] Call dentist (in 2h0m0s)",
				UserMessage:   "hello",
			},
			wantN:    2,
			wantLast: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.BuildPrompt(tt.input)
			if len(result.Messages) != tt.wantN {
				t.Fatalf("got %d messages, want %d", len(result.Messages), tt.wantN)
			}
			if result.Messages[0].Role != "system" {
				t.Errorf("first message role = %q, want 'system'", result.Messages[0].Role)
			}
			last := result.Messages[len(result.Messages)-1]
			if last.Role != "user" {
				t.Errorf("last message role = %q, want 'user'", last.Role)
			}
			if last.Content != tt.wantLast {
				t.Errorf("last message content = %q, want %q", last.Content, tt.wantLast)
			}
		})
	}
}

func TestAgent_LoadActivePersona(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	cfg := agent.DefaultConfig()
	cfg.PersonaDir = "/tmp/personas"
	cfg.ActivePersona = "default"

	personaLoader.EXPECT().LoadPersona("/tmp/personas/default.md").Return(&persona.Persona{
		Name: "default", Body: "# Default\n\nYou are the default persona.",
	}, nil)

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	if err := a.LoadActivePersona(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ActivePersona() == nil {
		t.Fatal("expected non-nil active persona")
	}
	if a.ActivePersona().Name != "default" {
		t.Errorf("name = %q, want %q", a.ActivePersona().Name, "default")
	}
}

func TestAgent_LoadActivePersona_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	cfg := agent.DefaultConfig()
	cfg.PersonaDir = "/tmp/personas"
	cfg.ActivePersona = "nonexistent"

	personaLoader.EXPECT().LoadPersona("/tmp/personas/nonexistent.md").Return(nil, &persona.ErrPersonaNotFound{Name: "nonexistent"})

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	if err := a.LoadActivePersona(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ActivePersona() != nil {
		t.Fatal("expected nil active persona when not found")
	}
}

func TestAgent_LoadActivePersona_NoDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	cfg := agent.DefaultConfig()

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	if err := a.LoadActivePersona(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgent_SetActivePersona(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	cfg := agent.DefaultConfig()
	cfg.PersonaDir = "/tmp/personas"

	personaLoader.EXPECT().LoadPersona("/tmp/personas/creative.md").Return(&persona.Persona{
		Name: "creative", Body: "# Creative\n\nYou are creative.",
	}, nil)

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	if err := a.SetActivePersona(context.Background(), "creative"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ActivePersona() == nil {
		t.Fatal("expected non-nil active persona")
	}
	if a.ActivePersona().Name != "creative" {
		t.Errorf("name = %q, want %q", a.ActivePersona().Name, "creative")
	}
}

func TestAgent_SetActivePersona_NoDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	cfg := agent.DefaultConfig()

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	if err := a.SetActivePersona(context.Background(), "creative"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAgent_SetActivePersona_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	cfg := agent.DefaultConfig()
	cfg.PersonaDir = "/tmp/personas"

	personaLoader.EXPECT().LoadPersona("/tmp/personas/nonexistent.md").Return(nil, &persona.ErrPersonaNotFound{Name: "nonexistent"})

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	if err := a.SetActivePersona(context.Background(), "nonexistent"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAgent_ListPersonas(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	cfg := agent.DefaultConfig()
	cfg.PersonaDir = "/tmp/personas"
	cfg.ActivePersona = "default"

	personaLoader.EXPECT().ListPersonas("/tmp/personas", "default").Return([]persona.Summary{
		{Name: "creative", Active: false},
		{Name: "default", Active: true},
	}, nil)

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	summaries, err := a.ListPersonas(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("got %d summaries, want 2", len(summaries))
	}
}

func TestDetectPersonaSwitch(t *testing.T) {
	tests := []struct {
		name           string
		userMsg        string
		currentPersona *persona.Persona
		want           string
	}{
		{
			name:    "switch to creative",
			userMsg: "switch to creative",
			want:    "creative",
		},
		{
			name:    "switch to creative mode",
			userMsg: "switch to creative mode",
			want:    "creative",
		},
		{
			name:    "change to quick persona",
			userMsg: "change to quick persona",
			want:    "quick",
		},
		{
			name:    "use default",
			userMsg: "use default",
			want:    "default",
		},
		{
			name:    "activate creative personality",
			userMsg: "activate creative personality",
			want:    "creative",
		},
		{
			name:    "no switch - normal message",
			userMsg: "What do you think about Go?",
			want:    "",
		},
		{
			name:           "no switch - same persona",
			userMsg:        "switch to default",
			currentPersona: &persona.Persona{Name: "default"},
			want:           "",
		},
		{
			name:    "case insensitive",
			userMsg: "Switch To Creative",
			want:    "creative",
		},
		{
			name:    "switch with trailing text",
			userMsg: "switch to creative mode please",
			want:    "creative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.DetectPersonaSwitch(tt.userMsg, tt.currentPersona)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgent_HandleMessage_WithPersonaSwitch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)
	scheduler := mock_agent.NewMockScheduler(ctrl)

	cfg := agent.DefaultConfig()
	cfg.PersonaDir = "/tmp/personas"
	cfg.ActivePersona = "default"

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "switch to creative").Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
	scheduler.EXPECT().GetUpcomingTasks(gomock.Any()).Return("", nil)

	personaLoader.EXPECT().LoadPersona("/tmp/personas/creative.md").Return(&persona.Persona{
		Name: "creative", Body: "# Creative\n\nYou are a creative assistant.",
	}, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "I'm now in creative mode!"}, FinishReason: "stop"}},
	}, nil)
	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	a := agent.NewAgent(store, provider, embedder, personaLoader, scheduler, &cfg)
	resp, err := a.HandleMessage(context.Background(), "switch to creative")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if a.ActivePersona() == nil {
		t.Fatal("expected active persona to be set")
	}
	if a.ActivePersona().Name != "creative" {
		t.Errorf("active persona = %q, want %q", a.ActivePersona().Name, "creative")
	}
}
