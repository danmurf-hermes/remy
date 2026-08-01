package agent_test

import (
	"context"
	"encoding/json"
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
)

func TestAgent_QuickConsolidation(t *testing.T) {
	tests := []struct {
		name          string
		messages      []memory.Message
		mock          func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder)
		wantErr       bool
		wantSummarize bool
	}{
		{
			name:    "no messages",
			messages: nil,
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
			},
			wantErr: false,
		},
		{
			name: "empty messages",
			messages: []memory.Message{},
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
			},
			wantErr: false,
		},
		{
			name: "summarizes and stores episode",
			messages: []memory.Message{
				{ID: uuid.NewString(), Role: "user", Content: "What do you think about Go?", Timestamp: time.Now().Add(-10 * time.Minute).UnixMilli()},
				{ID: uuid.NewString(), Role: "assistant", Content: "Go is great for backend development!", Timestamp: time.Now().Add(-9 * time.Minute).UnixMilli()},
			},
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-summary-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "User asked about Go programming and was told it's great for backend development."}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveEpisode(gomock.Any(), gomock.Any()).Do(func(_ context.Context, ep *memory.Episode) {
					if ep.ID == "" {
						t.Error("expected non-empty episode ID")
					}
					if ep.Summary == "" {
						t.Error("expected non-empty summary")
					}
					if ep.StartTime == 0 || ep.EndTime == 0 {
						t.Error("expected non-zero timestamps")
					}
					if ep.MessageIDs == "[]" || ep.MessageIDs == "" {
						t.Error("expected non-empty message IDs")
					}
					if ep.Importance <= 0 {
						t.Error("expected positive importance")
					}
				}).Return(nil)
				embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				store.EXPECT().SaveEpisodeVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "LLM summary error",
			messages: []memory.Message{
				{ID: uuid.NewString(), Role: "user", Content: "Hello", Timestamp: time.Now().UnixMilli()},
			},
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(nil, errors.New("LLM is down"))
			},
			wantErr: true,
		},
		{
			name: "LLM returns empty choices",
			messages: []memory.Message{
				{ID: uuid.NewString(), Role: "user", Content: "Hello", Timestamp: time.Now().UnixMilli()},
			},
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{},
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "save episode error",
			messages: []memory.Message{
				{ID: uuid.NewString(), Role: "user", Content: "Hello", Timestamp: time.Now().UnixMilli()},
			},
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Summary."}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveEpisode(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "embedding error",
			messages: []memory.Message{
				{ID: uuid.NewString(), Role: "user", Content: "Hello", Timestamp: time.Now().UnixMilli()},
			},
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Summary."}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveEpisode(gomock.Any(), gomock.Any()).Return(nil)
				embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(nil, errors.New("embedding failed"))
			},
			wantErr: true,
		},
		{
			name: "save episode vector error",
			messages: []memory.Message{
				{ID: uuid.NewString(), Role: "user", Content: "Hello", Timestamp: time.Now().UnixMilli()},
			},
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Summary."}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveEpisode(gomock.Any(), gomock.Any()).Return(nil)
				embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				store.EXPECT().SaveEpisodeVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("vector save failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_agent.NewMockStore(ctrl)
			provider := mock_llm.NewMockProvider(ctrl)
			embedder := mock_agent.NewMockEmbedder(ctrl)
			personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

			tt.mock(store, provider, embedder)

			cfg := agent.DefaultConfig()
			a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

			err := a.QuickConsolidation(context.Background(), tt.messages)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAgent_DeepConsolidation(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder)
		wantErr bool
	}{
		{
			name: "no episodes",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "get episodes error",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return(nil, errors.New("episode retrieval failed"))
			},
			wantErr: true,
		},
		{
			name: "extracts facts and entities",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "User asked about Go programming and prefers it for backend work.", Importance: 0.8},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)

				// Fact extraction
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-fact-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"fact": "User prefers Go for backend development", "category": "preference", "confidence": 0.9}]`}, FinishReason: "stop"}},
				}, nil)

				store.EXPECT().SaveFact(gomock.Any(), gomock.Any()).Do(func(_ context.Context, fact *memory.Fact) {
					if fact.Fact != "User prefers Go for backend development" {
						t.Errorf("fact = %q, want %q", fact.Fact, "User prefers Go for backend development")
					}
					if fact.Category != "preference" {
						t.Errorf("category = %q, want %q", fact.Category, "preference")
					}
				}).Return(nil)

				embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				store.EXPECT().SaveFactVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				// Entity extraction
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-entity-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"name": "Go", "type": "tool", "description": "A programming language"}]`}, FinishReason: "stop"}},
				}, nil)

				store.EXPECT().SaveEntity(gomock.Any(), gomock.Any()).Do(func(_ context.Context, entity memory.Entity) {
					if entity.Name != "Go" {
						t.Errorf("entity name = %q, want %q", entity.Name, "Go")
					}
					if entity.Type != "tool" {
						t.Errorf("entity type = %q, want %q", entity.Type, "tool")
					}
				}).Return(nil)

				// Relationship extraction
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-rel-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"source": "User", "target": "Go", "relationship": "prefers_over", "confidence": 0.8}]`}, FinishReason: "stop"}},
				}, nil)

				store.EXPECT().SaveRelationship(gomock.Any(), gomock.Any()).Do(func(_ context.Context, rel *memory.Relationship) {
					if rel.SourceEntity != "User" {
						t.Errorf("source = %q, want %q", rel.SourceEntity, "User")
					}
					if rel.Relationship != "prefers_over" {
						t.Errorf("relationship = %q, want %q", rel.Relationship, "prefers_over")
					}
				}).Return(nil)

				store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "deduplicates existing facts",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "User likes Go for backend.", Importance: 0.7},
				}, nil)

				existingFactID := uuid.NewString()
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return([]memory.Fact{
					{ID: existingFactID, Fact: "User prefers Go for backend development", Category: "preference", Confidence: 0.7, CreatedAt: time.Now().UnixMilli(), UpdatedAt: time.Now().UnixMilli()},
				}, nil)

				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-fact-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"fact": "User prefers Go for backend development", "category": "preference", "confidence": 0.9}]`}, FinishReason: "stop"}},
				}, nil)

				store.EXPECT().UpdateFact(gomock.Any(), gomock.Any()).Do(func(_ context.Context, fact *memory.Fact) {
					if fact.ID != existingFactID {
						t.Errorf("fact ID = %q, want %q", fact.ID, existingFactID)
					}
					if fact.Confidence < 0.79 || fact.Confidence > 0.81 {
						t.Errorf("confidence = %f, want ~0.8", fact.Confidence)
					}
				}).Return(nil)

				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-entity-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)

				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-rel-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)

				store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "fact extraction error",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "Test conversation.", Importance: 0.5},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(nil, errors.New("LLM error"))
			},
			wantErr: true,
		},
		{
			name: "malformed JSON from LLM",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "Test.", Importance: 0.5},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "not valid json at all"}, FinishReason: "stop"}},
				}, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "fact save error",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "Test.", Importance: 0.5},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"fact": "Test fact", "category": "preference", "confidence": 0.5}]`}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveFact(gomock.Any(), gomock.Any()).Return(errors.New("save failed"))
			},
			wantErr: true,
		},
		{
			name: "fact vector save error",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "Test.", Importance: 0.5},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"fact": "Test fact", "category": "preference", "confidence": 0.5}]`}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveFact(gomock.Any(), gomock.Any()).Return(nil)
				embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(nil, errors.New("embedding failed"))
			},
			wantErr: true,
		},
		{
			name: "update fact error",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "Test.", Importance: 0.5},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return([]memory.Fact{
					{ID: uuid.NewString(), Fact: "Test fact", Category: "preference", Confidence: 0.5},
				}, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"fact": "Test fact", "category": "preference", "confidence": 0.9}]`}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().UpdateFact(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))
			},
			wantErr: true,
		},
		{
			name: "entity save error",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "Test.", Importance: 0.5},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"name": "Go", "type": "tool", "description": "A language"}]`}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveEntity(gomock.Any(), gomock.Any()).Return(errors.New("entity save failed"))
			},
			wantErr: true,
		},
		{
			name: "relationship save error",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "Test.", Importance: 0.5},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"source": "User", "target": "Go", "relationship": "uses", "confidence": 0.8}]`}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveRelationship(gomock.Any(), gomock.Any()).Return(errors.New("relationship save failed"))
			},
			wantErr: true,
		},
		{
			name: "JSON with markdown code fences",
			mock: func(store *mock_agent.MockStore, provider *mock_llm.MockProvider, embedder *mock_agent.MockEmbedder) {
				store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
					{ID: uuid.NewString(), Summary: "Test.", Importance: 0.5},
				}, nil)
				store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "```json\n[{\"fact\": \"User likes testing\", \"category\": \"preference\", \"confidence\": 0.8}]\n```"}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().SaveFact(gomock.Any(), gomock.Any()).Do(func(_ context.Context, fact *memory.Fact) {
					if fact.Fact != "User likes testing" {
						t.Errorf("fact = %q, want %q", fact.Fact, "User likes testing")
					}
				}).Return(nil)
				embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
				store.EXPECT().SaveFactVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)
				provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
					ID: "mock-id", Object: "chat.completion",
					Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
				}, nil)
				store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_agent.NewMockStore(ctrl)
			provider := mock_llm.NewMockProvider(ctrl)
			embedder := mock_agent.NewMockEmbedder(ctrl)
			personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

			tt.mock(store, provider, embedder)

			cfg := agent.DefaultConfig()
			a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

			err := a.DeepConsolidation(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAgent_ScheduleConsolidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

	cfg := agent.DefaultConfig()
	cfg.QuickConsolidationDelayMs = 1
	cfg.DeepConsolidationDelayMs = 2

	a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

	store.EXPECT().GetMessages(gomock.Any(), 10, 0).Return(nil, nil).AnyTimes()

	stop := a.ScheduleConsolidation(context.Background())

	a.SignalActivity()
	time.Sleep(50 * time.Millisecond)

	stop()
}

func TestAgent_SignalActivity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

	cfg := agent.DefaultConfig()
	a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

	a.SignalActivity()
}

func TestAgent_Consolidation_MessageIDsJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

	msg1ID := uuid.NewString()
	msg2ID := uuid.NewString()

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Summary."}, FinishReason: "stop"}},
	}, nil)

	store.EXPECT().SaveEpisode(gomock.Any(), gomock.Any()).Do(func(_ context.Context, ep *memory.Episode) {
		var ids []string
		if err := json.Unmarshal([]byte(ep.MessageIDs), &ids); err != nil {
			t.Fatalf("failed to unmarshal message IDs: %v", err)
		}
		if len(ids) != 2 {
			t.Fatalf("got %d message IDs, want 2", len(ids))
		}
		if ids[0] != msg1ID {
			t.Errorf("ids[0] = %q, want %q", ids[0], msg1ID)
		}
		if ids[1] != msg2ID {
			t.Errorf("ids[1] = %q, want %q", ids[1], msg2ID)
		}
	}).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
	store.EXPECT().SaveEpisodeVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	cfg := agent.DefaultConfig()
	a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

	err := a.QuickConsolidation(context.Background(), []memory.Message{
		{ID: msg1ID, Role: "user", Content: "Hello", Timestamp: 1000},
		{ID: msg2ID, Role: "assistant", Content: "Hi!", Timestamp: 2000},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgent_Consolidation_ImportanceCalculation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Summary."}, FinishReason: "stop"}},
	}, nil)

	store.EXPECT().SaveEpisode(gomock.Any(), gomock.Any()).Do(func(_ context.Context, ep *memory.Episode) {
		if ep.Importance < 0.1 || ep.Importance > 1.0 {
			t.Errorf("importance %f out of range [0.1, 1.0]", ep.Importance)
		}
	}).Return(nil)

	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
	store.EXPECT().SaveEpisodeVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	cfg := agent.DefaultConfig()
	a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

	err := a.QuickConsolidation(context.Background(), []memory.Message{
		{ID: uuid.NewString(), Role: "user", Content: "Hi", Timestamp: 1000},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgent_Consolidation_EmptyFactExtraction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

	store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
		{ID: uuid.NewString(), Summary: "Casual chat about weather.", Importance: 0.3},
	}, nil)
	store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return(nil, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
	}, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
	}, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
	}, nil)

	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	cfg := agent.DefaultConfig()
	a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

	err := a.DeepConsolidation(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgent_Consolidation_ConfidenceCap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

	existingFactID := uuid.NewString()
	store.EXPECT().GetEpisodes(gomock.Any(), 20, 0).Return([]memory.Episode{
		{ID: uuid.NewString(), Summary: "Test.", Importance: 0.5},
	}, nil)
	store.EXPECT().GetFacts(gomock.Any(), 100, 0).Return([]memory.Fact{
		{ID: existingFactID, Fact: "Test fact", Category: "preference", Confidence: 0.95},
	}, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[{"fact": "Test fact", "category": "preference", "confidence": 0.9}]`}, FinishReason: "stop"}},
	}, nil)

	store.EXPECT().UpdateFact(gomock.Any(), gomock.Any()).Do(func(_ context.Context, fact *memory.Fact) {
		if fact.Confidence > 1.0 {
			t.Errorf("confidence %f exceeds 1.0", fact.Confidence)
		}
	}).Return(nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
	}, nil)

	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: `[]`}, FinishReason: "stop"}},
	}, nil)

	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil)

	cfg := agent.DefaultConfig()
	a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

	err := a.DeepConsolidation(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgent_Consolidation_HandleMessageSignalsActivity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_agent.NewMockStore(ctrl)
	provider := mock_llm.NewMockProvider(ctrl)
	embedder := mock_agent.NewMockEmbedder(ctrl)
	personaLoader := mock_agent.NewMockPersonaLoader(ctrl)

	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	store.EXPECT().LogActivity(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), "Hello").Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SearchEpisodes(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().SearchFacts(gomock.Any(), gomock.Any(), 5).Return(nil, nil)
	store.EXPECT().GetScratchpad(gomock.Any()).Return("", nil)
	store.EXPECT().GetMessages(gomock.Any(), 20, 0).Return(nil, nil)
	provider.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(&llm.ChatResponse{
		ID: "mock-id", Object: "chat.completion",
		Choices: []llm.Choice{{Index: 0, Message: llm.Message{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"}},
	}, nil)
	store.EXPECT().SaveMessage(gomock.Any(), gomock.Any()).Do(func(_ context.Context, msg *memory.Message) {
		if msg.ID == "" {
			msg.ID = uuid.NewString()
		}
	})
	embedder.EXPECT().GenerateEmbedding(gomock.Any(), gomock.Any()).Return(make([]float32, 768), nil)
	store.EXPECT().SaveMessageVector(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	cfg := agent.DefaultConfig()
	a := agent.NewAgent(store, provider, embedder, personaLoader, &cfg)

	_, err := a.HandleMessage(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
