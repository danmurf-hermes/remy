package llm

// Message represents a single message in a chat conversation, with a
// role (user/assistant/system) and content.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request body for the OpenAI-compatible chat completions
// endpoint.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse is the response from the OpenAI-compatible chat completions
// endpoint.
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a single completion choice returned by the LLM.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage contains token usage statistics from an LLM response.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk is a single chunk from a streaming chat completion response.
type StreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a single choice within a streaming chunk.
type StreamChoice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

// Delta contains the incremental content update within a streaming choice.
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}
