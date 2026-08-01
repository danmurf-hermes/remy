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

type OllamaClient struct {
	endpoint       string
	apiKey         string
	chatModel      string
	embeddingModel string
	parameters     map[string]any
	httpClient     *http.Client
}

func NewOllamaClient(endpoint, apiKey, chatModel, embeddingModel string, parameters map[string]any) *OllamaClient {
	return &OllamaClient{
		endpoint:       endpoint,
		apiKey:         apiKey,
		chatModel:      chatModel,
		embeddingModel: embeddingModel,
		parameters:     parameters,
		httpClient:     defaultHTTPClient(),
	}
}

func (c *OllamaClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = c.chatModel
	}
	req.Stream = false

	body, err := c.doRequest(ctx, http.MethodPost, c.endpoint+chatEndpoint, req)
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

func (c *OllamaClient) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	if req.Model == "" {
		req.Model = c.chatModel
	}
	req.Stream = true

	body, err := c.doRequest(ctx, http.MethodPost, c.endpoint+chatEndpoint, req)
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

func (c *OllamaClient) Embed(ctx context.Context, text string) ([]float32, error) {
	req := embedRequest{
		Model: c.embeddingModel,
		Input: []string{text},
	}

	body, err := c.doRequest(ctx, http.MethodPost, c.endpoint+embedEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("embed request: %w", err)
	}
	defer func() { _ = body.Close() }()

	var resp embedResponse
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

	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	httpResp, err := c.httpClient.Do(httpReq)
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

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}
