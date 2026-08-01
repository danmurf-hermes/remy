package memory

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"time"

	vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
)

// Embedder generates vector embeddings by calling an OpenAI-compatible
// embeddings API endpoint.
type Embedder struct {
	Endpoint string
	Model    string
	Client   *http.Client
}

// NewEmbedder creates an Embedder that calls the given endpoint with the
// specified model name.
func NewEmbedder(endpoint, model string) *Embedder {
	return &Embedder{
		Endpoint: endpoint,
		Model:    model,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// GenerateEmbedding calls the embeddings API to produce a vector
// representation of the given text.
func (e *Embedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	req := embeddingRequest{
		Model: e.Model,
		Input: []string{text},
	}

	var resp embeddingResponse
	if err := e.doRequest(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("generating embedding: %w", err)
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

func (e *Embedder) doRequest(ctx context.Context, req, resp any) error {
	var buf bytes.Buffer
	if err := writeJSON(&buf, req); err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, e.Endpoint+"/embeddings", &buf)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := e.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", httpResp.StatusCode)
	}

	return readJSON(httpResp.Body, resp)
}

// SerializeVector converts a float32 slice to a byte slice for storage in
// the vec0 virtual table.
func SerializeVector(v []float32) ([]byte, error) {
	return vec.SerializeFloat32(v)
}

// DeserializeVector converts a byte slice back into a float32 vector.
func DeserializeVector(data []byte) ([]float32, error) {
	r := bytes.NewReader(data)
	var v []float32
	for r.Len() > 0 {
		var f float32
		if err := binary.Read(r, binary.LittleEndian, &f); err != nil {
			return nil, fmt.Errorf("deserializing vector: %w", err)
		}
		v = append(v, f)
	}
	return v, nil
}

// SaveMessageVector stores a vector embedding for a message in the
// message_vectors virtual table.
func (s *Store) SaveMessageVector(ctx context.Context, messageID string, embedding []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("message_vectors").
		Columns("id", "embedding").
		Values(messageID, embedding).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving message vector: %w", err)
	}
	return nil
}

// SaveEpisodeVector stores a vector embedding for an episode in the
// episode_vectors virtual table.
func (s *Store) SaveEpisodeVector(ctx context.Context, episodeID string, embedding []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("episode_vectors").
		Columns("id", "embedding").
		Values(episodeID, embedding).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving episode vector: %w", err)
	}
	return nil
}

// SaveFactVector stores a vector embedding for a fact in the fact_vectors
// virtual table.
func (s *Store) SaveFactVector(ctx context.Context, factID string, embedding []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("fact_vectors").
		Columns("id", "embedding").
		Values(factID, embedding).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving fact vector: %w", err)
	}
	return nil
}
