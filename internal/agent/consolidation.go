package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/danmurf/remy/internal/llm"
	"github.com/danmurf/remy/internal/memory"
)

const (
	quickConsolidationMessageCount = 10
	episodeSearchLimit             = 20
	factExtractionLimit            = 10
)

// QuickConsolidation summarizes recent messages into an episode and stores
// it with a vector embedding. It should be called after a period of inactivity.
func (a *Agent) QuickConsolidation(ctx context.Context, recentMessages []memory.Message) error {
	if len(recentMessages) == 0 {
		return nil
	}

	summary, err := a.summarizeConversation(ctx, recentMessages)
	if err != nil {
		return fmt.Errorf("summarizing conversation: %w", err)
	}

	startTime := recentMessages[0].Timestamp
	endTime := recentMessages[len(recentMessages)-1].Timestamp

	messageIDs := make([]string, len(recentMessages))
	for i, msg := range recentMessages {
		messageIDs[i] = msg.ID
	}
	messageIDsJSON, err := json.Marshal(messageIDs)
	if err != nil {
		return fmt.Errorf("marshaling message IDs: %w", err)
	}

	ep := &memory.Episode{
		ID:         uuid.NewString(),
		Summary:    summary,
		StartTime:  startTime,
		EndTime:    endTime,
		MessageIDs: string(messageIDsJSON),
		Importance: calculateImportance(recentMessages),
		Topics:     "[]",
	}

	if err := a.store.SaveEpisode(ctx, ep); err != nil {
		return fmt.Errorf("saving episode: %w", err)
	}

	embedding, err := a.embedder.GenerateEmbedding(ctx, summary)
	if err != nil {
		return fmt.Errorf("generating episode embedding: %w", err)
	}

	embeddingBytes, err := memory.SerializeVector(embedding)
	if err != nil {
		return fmt.Errorf("serializing episode embedding: %w", err)
	}

	if err := a.store.SaveEpisodeVector(ctx, ep.ID, embeddingBytes); err != nil {
		return fmt.Errorf("saving episode vector: %w", err)
	}

	a.logActivity(ctx, "consolidation", ep.ID, a.cfg.SessionID)

	return nil
}

// DeepConsolidation extracts facts, entities, and relationships from recent
// episodes, deduplicates facts, and updates confidence scores.
func (a *Agent) DeepConsolidation(ctx context.Context) error { //nolint:gocyclo // multiple extraction steps with dedup logic
	recentEpisodes, err := a.store.GetEpisodes(ctx, episodeSearchLimit, 0)
	if err != nil {
		return fmt.Errorf("getting recent episodes: %w", err)
	}

	if len(recentEpisodes) == 0 {
		return nil
	}

	existingFacts, err := a.store.GetFacts(ctx, factExtractionLimit*10, 0)
	if err != nil {
		return fmt.Errorf("getting existing facts: %w", err)
	}

	existingFactMap := make(map[string]*memory.Fact)
	for i := range existingFacts {
		existingFactMap[existingFacts[i].Fact] = &existingFacts[i]
	}

	episodeTexts := make([]string, 0, len(recentEpisodes))
	for _, ep := range recentEpisodes {
		episodeTexts = append(episodeTexts, ep.Summary)
	}
	combinedText := strings.Join(episodeTexts, "\n")

	facts, err := a.extractFacts(ctx, combinedText)
	if err != nil {
		return fmt.Errorf("extracting facts: %w", err)
	}

	for _, fact := range facts {
		if existing, ok := existingFactMap[fact.Fact]; ok {
			newConfidence := existing.Confidence + 0.1
			if newConfidence > 1.0 {
				newConfidence = 1.0
			}
			existing.Confidence = newConfidence
			existing.UpdatedAt = time.Now().UnixMilli()
			if err := a.store.UpdateFact(ctx, existing); err != nil {
				return fmt.Errorf("updating fact: %w", err)
			}
		} else {
			fact.ID = uuid.NewString()
			fact.CreatedAt = time.Now().UnixMilli()
			fact.UpdatedAt = fact.CreatedAt
			if err := a.store.SaveFact(ctx, &fact); err != nil {
				return fmt.Errorf("saving fact: %w", err)
			}

			embedding, err := a.embedder.GenerateEmbedding(ctx, fact.Fact)
			if err != nil {
				return fmt.Errorf("generating fact embedding: %w", err)
			}
			embeddingBytes, err := memory.SerializeVector(embedding)
			if err != nil {
				return fmt.Errorf("serializing fact embedding: %w", err)
			}
			if err := a.store.SaveFactVector(ctx, fact.ID, embeddingBytes); err != nil {
				return fmt.Errorf("saving fact vector: %w", err)
			}
		}
	}

	entities, err := a.extractEntities(ctx, combinedText)
	if err != nil {
		return fmt.Errorf("extracting entities: %w", err)
	}

	for _, entity := range entities {
		if err := a.store.SaveEntity(ctx, entity); err != nil {
			return fmt.Errorf("saving entity: %w", err)
		}
	}

	relationships, err := a.extractRelationships(ctx, combinedText)
	if err != nil {
		return fmt.Errorf("extracting relationships: %w", err)
	}

	for _, rel := range relationships {
		if err := a.store.SaveRelationship(ctx, &rel); err != nil {
			return fmt.Errorf("saving relationship: %w", err)
		}
	}

	a.logActivity(ctx, "deep_consolidation", "", a.cfg.SessionID)

	return nil
}

// ScheduleConsolidation starts a background goroutine that monitors for
// inactivity and triggers quick or deep consolidation as appropriate.
// It returns a function that stops the background goroutine.
func (a *Agent) ScheduleConsolidation(ctx context.Context) func() {
	go a.consolidationLoop(ctx)
	return func() {
		close(a.stopCh)
	}
}

// SignalActivity marks the current time as the last activity time,
// which the consolidation loop uses to detect inactivity.
func (a *Agent) SignalActivity() {
	a.lastActivityUnixMs.Store(time.Now().UnixMilli())
}

func (a *Agent) consolidationLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case sig := <-a.consolidationCh:
			if sig.quick {
				recentMessages, err := a.store.GetMessages(ctx, quickConsolidationMessageCount, 0)
				if err == nil && len(recentMessages) > 0 {
					_ = a.QuickConsolidation(ctx, recentMessages)
				}
			}
			if sig.deep {
				_ = a.DeepConsolidation(ctx)
			}
		case <-ticker.C:
			lastActivity := a.lastActivityUnixMs.Load()
			now := time.Now().UnixMilli()
			elapsed := now - lastActivity

			if elapsed >= int64(a.cfg.DeepConsolidationDelayMs) {
				select {
				case a.consolidationCh <- consolidationSignal{ctx: ctx, deep: true}:
				default:
				}
			} else if elapsed >= int64(a.cfg.QuickConsolidationDelayMs) {
				select {
				case a.consolidationCh <- consolidationSignal{ctx: ctx, quick: true}:
				default:
				}
			}
		}
	}
}

func (a *Agent) summarizeConversation(ctx context.Context, messages []memory.Message) (string, error) {
	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}

	prompt := fmt.Sprintf(`Summarize the following conversation in 1-2 sentences. Focus on the key topics discussed, decisions made, and any important information shared.

Conversation:
%s

Summary:`, sb.String())

	resp, err := a.provider.Chat(ctx, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful assistant that summarizes conversations concisely."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("calling LLM for summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices for summary")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (a *Agent) extractFacts(ctx context.Context, text string) ([]memory.Fact, error) {
	prompt := fmt.Sprintf(`Extract facts about the user from the following conversation summaries. Return a JSON array of objects with "fact" (the fact statement), "category" (one of: preference, trait, personal_info, habit), and "confidence" (0.0 to 1.0). Only include clear, specific facts. If no facts are found, return an empty array.

Conversation summaries:
%s

JSON response:`, text)

	resp, err := a.provider.Chat(ctx, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You extract factual information about users from conversation text. Return only valid JSON arrays."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("calling LLM for fact extraction: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	content = stripJSONMarkdown(content)

	var extracted []struct {
		Fact       string  `json:"fact"`
		Category   string  `json:"category"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(content), &extracted); err != nil {
		return nil, nil
	}

	facts := make([]memory.Fact, 0, len(extracted))
	for _, ef := range extracted {
		if ef.Fact == "" {
			continue
		}
		if ef.Category == "" {
			ef.Category = "preference"
		}
		if ef.Confidence <= 0 {
			ef.Confidence = 0.5
		}
		facts = append(facts, memory.Fact{
			Fact:       ef.Fact,
			Category:   ef.Category,
			Confidence: ef.Confidence,
		})
	}

	return facts, nil
}

func (a *Agent) extractEntities(ctx context.Context, text string) ([]memory.Entity, error) {
	prompt := fmt.Sprintf(`Extract named entities from the following conversation summaries. Return a JSON array of objects with "name" (the entity name), "type" (one of: person, place, concept, project, tool), and "description" (a brief description). Only include clearly identifiable entities. If no entities are found, return an empty array.

Conversation summaries:
%s

JSON response:`, text)

	resp, err := a.provider.Chat(ctx, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You extract named entities from conversation text. Return only valid JSON arrays."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("calling LLM for entity extraction: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	content = stripJSONMarkdown(content)

	var extracted []struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(content), &extracted); err != nil {
		return nil, nil
	}

	entities := make([]memory.Entity, 0, len(extracted))
	for _, ee := range extracted {
		if ee.Name == "" {
			continue
		}
		if ee.Type == "" {
			ee.Type = "concept"
		}
		entities = append(entities, memory.Entity{
			ID:          uuid.NewString(),
			Name:        ee.Name,
			Type:        ee.Type,
			Description: ee.Description,
			CreatedAt:   time.Now().UnixMilli(),
		})
	}

	return entities, nil
}

func (a *Agent) extractRelationships(ctx context.Context, text string) ([]memory.Relationship, error) {
	prompt := fmt.Sprintf(`Extract relationships between entities from the following conversation summaries. Return a JSON array of objects with "source" (the source entity name), "target" (the target entity name), "relationship" (the relationship type, e.g. works_on, lives_in, prefers_over, is_a), and "confidence" (0.0 to 1.0). Only include clear, explicit relationships. If no relationships are found, return an empty array.

Conversation summaries:
%s

JSON response:`, text)

	resp, err := a.provider.Chat(ctx, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You extract relationships between entities from conversation text. Return only valid JSON arrays."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("calling LLM for relationship extraction: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	content = stripJSONMarkdown(content)

	var extracted []struct {
		Source       string  `json:"source"`
		Target       string  `json:"target"`
		Relationship string  `json:"relationship"`
		Confidence   float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(content), &extracted); err != nil {
		return nil, nil
	}

	relationships := make([]memory.Relationship, 0, len(extracted))
	for _, er := range extracted {
		if er.Source == "" || er.Target == "" || er.Relationship == "" {
			continue
		}
		if er.Confidence <= 0 {
			er.Confidence = 0.5
		}
		relationships = append(relationships, memory.Relationship{
			ID:           uuid.NewString(),
			SourceEntity: er.Source,
			TargetEntity: er.Target,
			Relationship: er.Relationship,
			Confidence:   er.Confidence,
			CreatedAt:    time.Now().UnixMilli(),
		})
	}

	return relationships, nil
}

func calculateImportance(messages []memory.Message) float64 {
	if len(messages) == 0 {
		return 0.5
	}
	totalLen := 0
	for _, msg := range messages {
		totalLen += len(msg.Content)
	}
	avgLen := float64(totalLen) / float64(len(messages))
	importance := avgLen / 500.0
	if importance > 1.0 {
		importance = 1.0
	}
	if importance < 0.1 {
		importance = 0.1
	}
	return importance
}

func stripJSONMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}
