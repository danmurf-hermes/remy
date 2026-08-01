package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewOllamaClient(t *testing.T) {
	c := NewOllamaClient("http://localhost:11434/v1", "", "llama3.1:8b", "nomic-embed-text", nil)
	if c.endpoint != "http://localhost:11434/v1" {
		t.Errorf("endpoint = %q, want %q", c.endpoint, "http://localhost:11434/v1")
	}
	if c.chatModel != "llama3.1:8b" {
		t.Errorf("chatModel = %q, want %q", c.chatModel, "llama3.1:8b")
	}
	if c.embeddingModel != "nomic-embed-text" {
		t.Errorf("embeddingModel = %q, want %q", c.embeddingModel, "nomic-embed-text")
	}
	if c.httpClient == nil {
		t.Fatal("httpClient is nil")
	}
}

func TestChat(t *testing.T) {
	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		req     ChatRequest
		wantErr bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
					t.Errorf("path = %s, want /chat/completions", r.URL.Path)
				}
				var req ChatRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("decoding request: %v", err)
				}
				if req.Model != "llama3.1:8b" {
					t.Errorf("model = %q, want %q", req.Model, "llama3.1:8b")
				}
				if req.Stream {
					t.Error("expected stream=false for non-streaming chat")
				}
				if len(req.Messages) != 1 || req.Messages[0].Content != "hello" {
					t.Errorf("messages = %v, want [hello]", req.Messages)
				}
				_ = json.NewEncoder(w).Encode(ChatResponse{
					ID: "chat-1", Object: "chat.completion",
					Choices: []Choice{{Index: 0, Message: Message{Role: "assistant", Content: "Hi there!"}, FinishReason: "stop"}},
					Usage:   Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
				})
			},
			req: ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}},
		},
		{
			name: "empty choices",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(ChatResponse{ID: "chat-1", Object: "chat.completion", Choices: []Choice{}})
			},
			req:     ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}},
			wantErr: true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("internal error"))
			},
			req:     ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}},
			wantErr: true,
		},
		{
			name: "timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
			},
			req:     ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}},
			wantErr: true,
		},
		{
			name: "malformed response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{invalid json`))
			},
			req:     ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}},
			wantErr: true,
		},
		{
			name: "with API key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-key-123" {
					t.Errorf("Authorization = %q, want %q", auth, "Bearer test-key-123")
				}
				_ = json.NewEncoder(w).Encode(ChatResponse{
					Choices: []Choice{{Message: Message{Role: "assistant", Content: "ok"}}},
				})
			},
			req: ChatRequest{Messages: []Message{{Role: "user", Content: "hello"}}},
		},
		{
			name: "model override",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var req ChatRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("decoding request: %v", err)
				}
				if req.Model != "custom-model" {
					t.Errorf("model = %q, want %q", req.Model, "custom-model")
				}
				_ = json.NewEncoder(w).Encode(ChatResponse{
					Choices: []Choice{{Message: Message{Role: "assistant", Content: "ok"}}},
				})
			},
			req: ChatRequest{Model: "custom-model", Messages: []Message{{Role: "user", Content: "hello"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			apiKey := ""
			if tt.name == "with API key" {
				apiKey = "test-key-123"
			}
			c := NewOllamaClient(server.URL, apiKey, "llama3.1:8b", "nomic-embed-text", nil)

			ctx := context.Background()
			if tt.name == "timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			_, err := c.Chat(ctx, tt.req)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestChatStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if !req.Stream {
			t.Error("expected stream=true for streaming chat")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		chunks := []StreamChunk{
			{Choices: []StreamChoice{{Delta: Delta{Content: "Hello"}}}},
			{Choices: []StreamChoice{{Delta: Delta{Content: " world"}}}},
			{Choices: []StreamChoice{{Delta: Delta{Content: ""}, FinishReason: "stop"}}},
		}
		for _, chunk := range chunks {
			data, _ := json.Marshal(chunk)
			_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		}
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	c := NewOllamaClient(server.URL, "", "llama3.1:8b", "nomic-embed-text", nil)
	ch, err := c.ChatStream(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}

	var contents []string
	for chunk := range ch {
		for _, choice := range chunk.Choices {
			contents = append(contents, choice.Delta.Content)
		}
	}

	expected := []string{"Hello", " world", ""}
	if len(contents) != len(expected) {
		t.Fatalf("got %d chunks, want %d: %v", len(contents), len(expected), contents)
	}
	for i := range expected {
		if contents[i] != expected[i] {
			t.Errorf("chunk %d = %q, want %q", i, contents[i], expected[i])
		}
	}
}

func TestChatStream_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := NewOllamaClient(server.URL, "", "llama3.1:8b", "nomic-embed-text", nil)
	_, err := c.ChatStream(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}

func TestChatStream_CanceledContext(t *testing.T) {
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		flusher.Flush()
		chunk := StreamChunk{Choices: []StreamChoice{{Delta: Delta{Content: "Hello"}}}}
		data, _ := json.Marshal(chunk)
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		flusher.Flush()
		<-done
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(done) })

	c := NewOllamaClient(server.URL, "", "llama3.1:8b", "nomic-embed-text", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ch, err := c.ChatStream(ctx, ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}

	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Errorf("got %d chunks, want 1", count)
	}
}

func TestEmbed(t *testing.T) {
	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		text    string
		wantErr bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/embeddings") {
					t.Errorf("path = %s, want /embeddings", r.URL.Path)
				}
				var req embedRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("decoding request: %v", err)
				}
				if req.Model != "nomic-embed-text" {
					t.Errorf("model = %q, want %q", req.Model, "nomic-embed-text")
				}
				if len(req.Input) != 1 || req.Input[0] != "hello world" {
					t.Errorf("input = %v, want [hello world]", req.Input)
				}
				_ = json.NewEncoder(w).Encode(embedResponse{
					Data: []struct {
						Embedding []float64 `json:"embedding"`
					}{
						{Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5}},
					},
				})
			},
			text: "hello world",
		},
		{
			name: "empty response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(embedResponse{Data: []struct {
					Embedding []float64 `json:"embedding"`
				}{}})
			},
			text:    "test",
			wantErr: true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			text:    "test",
			wantErr: true,
		},
		{
			name: "timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
			},
			text:    "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			c := NewOllamaClient(server.URL, "", "llama3.1:8b", "nomic-embed-text", nil)

			ctx := context.Background()
			if tt.name == "timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			embedding, err := c.Embed(ctx, tt.text)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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
		})
	}
}
