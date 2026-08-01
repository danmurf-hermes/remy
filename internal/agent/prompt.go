package agent

import (
	"fmt"
	"strings"

	"github.com/yourname/remy/internal/llm"
	"github.com/yourname/remy/internal/memory"
)

// PromptInput holds all the data needed to build a prompt for the LLM.
type PromptInput struct {
	Scratchpad     string
	Episodes       []memory.Episode
	Facts          []memory.Fact
	RecentMessages []memory.Message
	UserMessage    string
}

// PromptResult contains the constructed messages ready to send to the LLM.
type PromptResult struct {
	Messages []llm.Message
}

// BuildPrompt constructs the full prompt from the given input, assembling
// the system prompt, scratchpad, retrieved context, conversation history,
// and the current user message.
func BuildPrompt(input *PromptInput) PromptResult {
	var sections []string

	sections = append(sections, `You are Remy, a helpful personal assistant. You are friendly, concise, and proactive.

You have access to a memory system that stores:
- Episodic memory: summaries of past conversations
- Semantic memory: facts about the user
- A scratchpad: your working notes

Use the retrieved context below to inform your responses. If the context is relevant, reference it naturally. If it's not relevant, ignore it.`)

	if input.Scratchpad != "" {
		sections = append(sections, fmt.Sprintf(`## Scratchpad (Your Working Notes)

%s`, input.Scratchpad))
	}

	if len(input.Episodes) > 0 {
		epStrs := make([]string, 0, len(input.Episodes))
		for _, ep := range input.Episodes {
			epStrs = append(epStrs, fmt.Sprintf("- %s (importance: %.1f)", ep.Summary, ep.Importance))
		}
		sections = append(sections, fmt.Sprintf(`## Relevant Past Conversations

%s`, strings.Join(epStrs, "\n")))
	}

	if len(input.Facts) > 0 {
		factStrs := make([]string, 0, len(input.Facts))
		for _, f := range input.Facts {
			factStrs = append(factStrs, fmt.Sprintf("- [%s] %s (confidence: %.1f)", f.Category, f.Fact, f.Confidence))
		}
		sections = append(sections, fmt.Sprintf(`## Facts About the User

%s`, strings.Join(factStrs, "\n")))
	}

	systemPrompt := strings.Join(sections, "\n\n")

	msgCount := 1 + len(input.RecentMessages) + 1 // system + history + user
	messages := make([]llm.Message, 0, msgCount)
	messages = append(messages, llm.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, msg := range input.RecentMessages {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, llm.Message{
		Role:    "user",
		Content: input.UserMessage,
	})

	return PromptResult{Messages: messages}
}
