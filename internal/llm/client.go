// Package llm provides an OpenAI-compatible client for interacting with LLM
// providers (e.g., Ollama). It supports chat, streaming, and embeddings.
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	chatEndpoint   = "/chat/completions"
	embedEndpoint  = "/embeddings"
	requestTimeout = 60 * time.Second
)

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: requestTimeout}
}

// OllamaClient implements the Provider interface for OpenAI-compatible
// LLM endpoints (Ollama, OpenAI, etc.).
type OllamaClient struct {
	Endpoint       string
	APIKey         string
	ChatModel      string
	EmbeddingModel string
	Parameters     map[string]any
	HTTPClient     *http.Client
}

// NewOllamaClient creates a new OllamaClient with the given endpoint, auth,
// model names, and optional parameters (temperature, max_tokens, etc.).
func NewOllamaClient(endpoint, apiKey, chatModel, embeddingModel string, parameters map[string]any) *OllamaClient {
	return &OllamaClient{
		Endpoint:       endpoint,
		APIKey:         apiKey,
		ChatModel:      chatModel,
		EmbeddingModel: embeddingModel,
		Parameters:     parameters,
		HTTPClient:     defaultHTTPClient(),
	}
}

// Chat sends a chat completion request and returns the full response.
// If req.Model is empty, the client's configured chat model is used.
func (c *OllamaClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = c.ChatModel
	}
	req.Stream = false

	body, err := c.doRequest(ctx, http.MethodPost, c.Endpoint+chatEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("chat request: %w", err)
	}
	defer func() { _ = body.Close() }()

	var resp ChatResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decoding chat response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in chat response")
	}

	return &resp, nil
}

// ChatStream sends a streaming chat completion request and returns a
// channel of StreamChunks. The caller must read from the channel until
// it is closed. The stream is canceled when the context is done.
func (c *OllamaClient) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	if req.Model == "" {
		req.Model = c.ChatModel
	}
	req.Stream = true

	body, err := c.doRequest(ctx, http.MethodPost, c.Endpoint+chatEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("chat stream request: %w", err)
	}

	ch := make(chan StreamChunk, 64)
	go func() {
		defer func() { _ = body.Close() }()
		defer close(ch)

		scanner := bufio.NewScanner(body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// Embed generates a vector embedding for the given text using the
// client's configured embedding model.
func (c *OllamaClient) Embed(ctx context.Context, text string) ([]float32, error) {
	req := EmbedRequest{
		Model: c.EmbeddingModel,
		Input: []string{text},
	}

	body, err := c.doRequest(ctx, http.MethodPost, c.Endpoint+embedEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("embed request: %w", err)
	}
	defer func() { _ = body.Close() }()

	var resp EmbedResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decoding embed response: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	embedding := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

func (c *OllamaClient) doRequest(ctx context.Context, method, url string, reqBody any) (io.ReadCloser, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		_ = httpResp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d: %s", httpResp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return httpResp.Body, nil
}

// EmbedRequest is the request body for the embeddings API.
type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbedResponse is the response from the embeddings API.
type EmbedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}
