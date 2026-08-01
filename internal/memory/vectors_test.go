package memory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestSerializeDeserializeVector(t *testing.T) {
	original := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	data, err := SerializeVector(original)
	if err != nil {
		t.Fatalf("SerializeVector: %v", err)
	}

	got, err := DeserializeVector(data)
	if err != nil {
		t.Fatalf("DeserializeVector: %v", err)
	}

	if len(got) != len(original) {
		t.Fatalf("length = %d, want %d", len(got), len(original))
	}
	for i := range original {
		if got[i] != original[i] {
			t.Errorf("element %d = %f, want %f", i, got[i], original[i])
		}
	}
}

func TestDeserializeVector_Empty(t *testing.T) {
	got, err := DeserializeVector([]byte{})
	if err != nil {
		t.Fatalf("DeserializeVector empty: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("length = %d, want 0", len(got))
	}
}

func TestNewEmbedder(t *testing.T) {
	e := NewEmbedder("http://localhost:11434/v1", "nomic-embed-text")
	if e.Endpoint != "http://localhost:11434/v1" {
		t.Errorf("endpoint = %q, want %q", e.Endpoint, "http://localhost:11434/v1")
	}
	if e.Model != "nomic-embed-text" {
		t.Errorf("model = %q, want %q", e.Model, "nomic-embed-text")
	}
	if e.Client == nil {
		t.Fatal("Client is nil")
	}
}

func TestGenerateEmbedding_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/embeddings" {
			t.Errorf("path = %s, want /embeddings", r.URL.Path)
		}

		var req embeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if req.Model != "nomic-embed-text" {
			t.Errorf("model = %q, want %q", req.Model, "nomic-embed-text")
		}
		if len(req.Input) != 1 || req.Input[0] != "hello world" {
			t.Errorf("input = %v, want [hello world]", req.Input)
		}

		resp := embeddingResponse{
			Data: []struct {
				Embedding []float64 `json:"embedding"`
			}{
				{Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	e := NewEmbedder(server.URL, "nomic-embed-text")
	embedding, err := e.GenerateEmbedding(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("GenerateEmbedding: %v", err)
	}

	expected := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	if len(embedding) != len(expected) {
		t.Fatalf("length = %d, want %d", len(embedding), len(expected))
	}
	for i := range expected {
		if embedding[i] != expected[i] {
			t.Errorf("element %d = %f, want %f", i, embedding[i], expected[i])
		}
	}
}

func TestGenerateEmbedding_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	e := NewEmbedder(server.URL, "nomic-embed-text")
	_, err := e.GenerateEmbedding(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}

func TestGenerateEmbedding_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(embeddingResponse{Data: []struct {
			Embedding []float64 `json:"embedding"`
		}{}})
	}))
	defer server.Close()

	e := NewEmbedder(server.URL, "nomic-embed-text")
	_, err := e.GenerateEmbedding(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
}

func TestGenerateEmbedding_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// never respond
		<-r.Context().Done()
	}))
	defer server.Close()

	e := NewEmbedder(server.URL, "nomic-embed-text")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := e.GenerateEmbedding(ctx, "test")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func make768Vector() []float32 {
	v := make([]float32, 768)
	for i := range v {
		v[i] = float32(i) / 768.0
	}
	return v
}

func TestSaveMessageVector(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	msg := Message{
		ID: uuid.NewString(), UserID: "u1", Role: "user",
		Content: "test", Timestamp: 1000, Interface: "gui",
	}
	if err := s.SaveMessage(ctx, msg); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	embedding, err := SerializeVector(make768Vector())
	if err != nil {
		t.Fatalf("SerializeVector: %v", err)
	}

	if err := s.SaveMessageVector(ctx, msg.ID, embedding); err != nil {
		t.Fatalf("SaveMessageVector: %v", err)
	}
}

func TestSaveEpisodeVector(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ep := Episode{
		ID: uuid.NewString(), Summary: "test",
		StartTime: 1000, EndTime: 2000, MessageIDs: "[]",
	}
	if err := s.SaveEpisode(ctx, ep); err != nil {
		t.Fatalf("SaveEpisode: %v", err)
	}

	embedding, err := SerializeVector(make768Vector())
	if err != nil {
		t.Fatalf("SerializeVector: %v", err)
	}

	if err := s.SaveEpisodeVector(ctx, ep.ID, embedding); err != nil {
		t.Fatalf("SaveEpisodeVector: %v", err)
	}
}

func TestSearchEpisodes(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ep := Episode{
		ID: uuid.NewString(), Summary: "user asked about Go programming",
		StartTime: 1000, EndTime: 2000, MessageIDs: "[]", Importance: 0.8,
	}
	if err := s.SaveEpisode(ctx, ep); err != nil {
		t.Fatalf("SaveEpisode: %v", err)
	}

	embedding, err := SerializeVector(make768Vector())
	if err != nil {
		t.Fatalf("SerializeVector: %v", err)
	}

	if err := s.SaveEpisodeVector(ctx, ep.ID, embedding); err != nil {
		t.Fatalf("SaveEpisodeVector: %v", err)
	}

	results, err := s.SearchEpisodes(ctx, embedding, 5)
	if err != nil {
		t.Fatalf("SearchEpisodes: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if results[0].Summary != ep.Summary {
		t.Errorf("summary = %q, want %q", results[0].Summary, ep.Summary)
	}
}

func TestSearchFacts(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := now()
	fact := Fact{
		ID: uuid.NewString(), Fact: "user prefers Go over Python",
		Category: "preference", Confidence: 0.9, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.SaveFact(ctx, fact); err != nil {
		t.Fatalf("SaveFact: %v", err)
	}

	embedding, err := SerializeVector(make768Vector())
	if err != nil {
		t.Fatalf("SerializeVector: %v", err)
	}

	if err := s.SaveFactVector(ctx, fact.ID, embedding); err != nil {
		t.Fatalf("SaveFactVector: %v", err)
	}

	results, err := s.SearchFacts(ctx, embedding, 5)
	if err != nil {
		t.Fatalf("SearchFacts: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if results[0].Fact != fact.Fact {
		t.Errorf("fact = %q, want %q", results[0].Fact, fact.Fact)
	}
}

func TestSaveFactVector(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := now()
	fact := Fact{
		ID: uuid.NewString(), Fact: "test fact",
		Category: "test", Confidence: 0.5, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.SaveFact(ctx, fact); err != nil {
		t.Fatalf("SaveFact: %v", err)
	}

	embedding, err := SerializeVector(make768Vector())
	if err != nil {
		t.Fatalf("SerializeVector: %v", err)
	}

	if err := s.SaveFactVector(ctx, fact.ID, embedding); err != nil {
		t.Fatalf("SaveFactVector: %v", err)
	}
}
