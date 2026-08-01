package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/remy/internal/agent/mock_agent"
	"github.com/yourname/remy/internal/llm"
	"github.com/yourname/remy/internal/llm/mock_llm"
	"github.com/yourname/remy/internal/memory"
	"go.uber.org/mock/gomock"
)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	cfg := DefaultConfig()

	agent := NewAgent(store, provider, embedder, cfg)
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
}

// TestAgent_HandleMessage_NormalFlow verifies the full message handling pipeline.
func TestAgent_HandleMessage_NormalFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	// Capture stored messages for later verification
	var savedMessages []*memory.Message
	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, msg *memory.Message) error {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
		savedMessages = append(savedMessages, msg)
		return nil
	}).Times(2)

	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello, Remy!").Return(make([]float32, 768), nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
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
	}, nil)

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
	if len(savedMessages) != 2 {
		t.Fatalf("expected 2 saved messages, got %d", len(savedMessages))
	}
	if savedMessages[0].Role != "user" || savedMessages[0].Content != "Hello, Remy!" {
		t.Errorf("unexpected first message: %+v", savedMessages[0])
	}
	if savedMessages[1].Role != "assistant" {
		t.Errorf("expected second message role 'assistant', got %q", savedMessages[1].Role)
	}
}

// TestAgent_HandleMessage_WithContext verifies context retrieval is included.
func TestAgent_HandleMessage_WithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	}).Times(2)
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "What do you think about Rust?").Return(make([]float32, 768), nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return([]memory.Episode{
		{
			ID:         uuid.NewString(),
			Summary:    "User asked about Go programming",
			StartTime:  time.Now().Add(-1 * time.Hour).UnixMilli(),
			EndTime:    time.Now().Add(-30 * time.Minute).UnixMilli(),
			Importance: 0.8,
		},
	}, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return([]memory.Fact{
		{
			ID:         uuid.NewString(),
			Fact:       "User prefers Go for backend development",
			Category:   "preference",
			Confidence: 0.9,
		},
	}, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("User is working on a CLI tool project", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID:     "mock-chat-id",
		Object: "chat.completion",
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    "assistant",
					Content: "Rust is great for systems programming!",
				},
				FinishReason: "stop",
			},
		},
	}, nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: ""}, FinishReason: "stop"}},
	}, nil)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(nil, errors.New("LLM is down"))

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from LLM failure")
	}
}

// TestAgent_HandleMessage_StoreSaveError verifies handling of store save failures.
func TestAgent_HandleMessage_StoreSaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Return(errors.New("database is locked"))

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from store save failure")
	}
}

// TestAgent_HandleMessage_EmbeddingError verifies handling of embedding failures.
func TestAgent_HandleMessage_EmbeddingError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(nil, errors.New("embedding service unavailable"))

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from embedding failure")
	}
}

// TestAgent_HandleMessage_EmptyLLMResponse verifies handling when LLM returns no choices.
func TestAgent_HandleMessage_EmptyLLMResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{},
	}, nil)

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from empty LLM response")
	}
}

// TestAgent_HandleMessage_WithCustomConfig verifies custom config is used.
func TestAgent_HandleMessage_WithCustomConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello from Telegram!").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 5, 0).Return(nil, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"}},
	}, nil)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", errors.New("scratchpad read error"))

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from scratchpad failure")
	}
}

// TestAgent_HandleMessage_SearchEpisodesError verifies episode search errors are handled.
func TestAgent_HandleMessage_SearchEpisodesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, errors.New("episode search failed"))

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from episode search failure")
	}
}

// TestAgent_HandleMessage_SearchFactsError verifies fact search errors are handled.
func TestAgent_HandleMessage_SearchFactsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, errors.New("fact search failed"))

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from fact search failure")
	}
}

// TestAgent_HandleMessage_GetMessagesError verifies message retrieval errors are handled.
func TestAgent_HandleMessage_GetMessagesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, errors.New("message retrieval failed"))

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from message retrieval failure")
	}
}

// TestAgent_HandleMessage_SaveVectorError verifies vector save errors are handled.
func TestAgent_HandleMessage_SaveVectorError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)

	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("vector save failed"))

	agent := newTestAgent(store, provider, embedder)

	_, err := agent.HandleMessage(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error from vector save failure")
	}
}

// TestAgent_HandleMessage_ConcurrentMessages verifies the agent handles
// multiple messages in sequence correctly.
func TestAgent_HandleMessage_ConcurrentMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)

	// First message expectations
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

	// Second message expectations
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

	agent := newTestAgent(store, provider, embedder)

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
