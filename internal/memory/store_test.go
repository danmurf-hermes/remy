package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore(%q): %v", dbPath, err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestNewStore_CreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "memory.db")
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestNewStore_InvalidPath(t *testing.T) {
	_, err := NewStore("/nonexistent/dir/memory.db")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestStore_Close(t *testing.T) {
	s := newTestStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// closing again should be safe
	s.Close()
}

func TestStore_DB(t *testing.T) {
	s := newTestStore(t)
	if s.DB() == nil {
		t.Fatal("DB() returned nil")
	}
}

func TestSaveAndGetMessage(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	msg := Message{
		ID:        uuid.NewString(),
		UserID:    "user1",
		Role:      "user",
		Content:   "Hello, Remy!",
		Timestamp: time.Now().UnixMilli(),
		Interface: "gui",
		SessionID: "session1",
	}

	if err := s.SaveMessage(ctx, msg); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	got, err := s.GetMessage(ctx, msg.ID)
	if err != nil {
		t.Fatalf("GetMessage: %v", err)
	}

	if got.Content != msg.Content {
		t.Errorf("content = %q, want %q", got.Content, msg.Content)
	}
	if got.Role != msg.Role {
		t.Errorf("role = %q, want %q", got.Role, msg.Role)
	}
	if got.Interface != msg.Interface {
		t.Errorf("interface = %q, want %q", got.Interface, msg.Interface)
	}
}

func TestGetMessage_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetMessage(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent message, got nil")
	}
}

func TestGetMessages(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		msg := Message{
			ID:        uuid.NewString(),
			UserID:    "user1",
			Role:      "user",
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: int64(1000 + i),
			Interface: "gui",
		}
		if err := s.SaveMessage(ctx, msg); err != nil {
			t.Fatalf("SaveMessage: %v", err)
		}
	}

	messages, err := s.GetMessages(ctx, 3, 0)
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(messages) != 3 {
		t.Errorf("got %d messages, want 3", len(messages))
	}

	messages, err = s.GetMessages(ctx, 10, 3)
	if err != nil {
		t.Fatalf("GetMessages with offset: %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("got %d messages with offset, want 2", len(messages))
	}
}

func TestGetMessagesBySession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		msg := Message{
			ID:        uuid.NewString(),
			UserID:    "user1",
			Role:      "user",
			Content:   "Session msg " + string(rune('0'+i)),
			Timestamp: int64(1000 + i),
			Interface: "gui",
			SessionID: "session-a",
		}
		if err := s.SaveMessage(ctx, msg); err != nil {
			t.Fatalf("SaveMessage: %v", err)
		}
	}

	// message in different session
	other := Message{
		ID:        uuid.NewString(),
		UserID:    "user1",
		Role:      "user",
		Content:   "Other session",
		Timestamp: 2000,
		Interface: "gui",
		SessionID: "session-b",
	}
	if err := s.SaveMessage(ctx, other); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	messages, err := s.GetMessagesBySession(ctx, "session-a")
	if err != nil {
		t.Fatalf("GetMessagesBySession: %v", err)
	}
	if len(messages) != 3 {
		t.Errorf("got %d messages, want 3", len(messages))
	}
}

func TestSaveAndGetEpisode(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ep := Episode{
		ID:         uuid.NewString(),
		Summary:    "User asked about Go vs Python",
		StartTime:  1000,
		EndTime:    2000,
		MessageIDs: `["msg1","msg2"]`,
		Importance: 0.8,
		Topics:     `["programming","comparison"]`,
	}

	if err := s.SaveEpisode(ctx, ep); err != nil {
		t.Fatalf("SaveEpisode: %v", err)
	}

	got, err := s.GetEpisode(ctx, ep.ID)
	if err != nil {
		t.Fatalf("GetEpisode: %v", err)
	}

	if got.Summary != ep.Summary {
		t.Errorf("summary = %q, want %q", got.Summary, ep.Summary)
	}
	if got.Importance != ep.Importance {
		t.Errorf("importance = %f, want %f", got.Importance, ep.Importance)
	}
}

func TestGetEpisodes(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		ep := Episode{
			ID:         uuid.NewString(),
			Summary:    "Episode " + string(rune('0'+i)),
			StartTime:  int64(1000 + i*100),
			EndTime:    int64(2000 + i*100),
			MessageIDs: "[]",
			Importance: 0.5,
		}
		if err := s.SaveEpisode(ctx, ep); err != nil {
			t.Fatalf("SaveEpisode: %v", err)
		}
	}

	episodes, err := s.GetEpisodes(ctx, 2, 0)
	if err != nil {
		t.Fatalf("GetEpisodes: %v", err)
	}
	if len(episodes) != 2 {
		t.Errorf("got %d episodes, want 2", len(episodes))
	}
}

func TestGetEpisodesByTimeRange(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	episodes := []Episode{
		{ID: uuid.NewString(), Summary: "early", StartTime: 100, EndTime: 200, MessageIDs: "[]", Importance: 0.5},
		{ID: uuid.NewString(), Summary: "middle", StartTime: 300, EndTime: 400, MessageIDs: "[]", Importance: 0.5},
		{ID: uuid.NewString(), Summary: "late", StartTime: 500, EndTime: 600, MessageIDs: "[]", Importance: 0.5},
	}

	for _, ep := range episodes {
		if err := s.SaveEpisode(ctx, ep); err != nil {
			t.Fatalf("SaveEpisode: %v", err)
		}
	}

	got, err := s.GetEpisodesByTimeRange(ctx, 250, 450)
	if err != nil {
		t.Fatalf("GetEpisodesByTimeRange: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d episodes, want 1", len(got))
	}
	if got[0].Summary != "middle" {
		t.Errorf("summary = %q, want %q", got[0].Summary, "middle")
	}
}

func TestSaveAndGetFact(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	fact := Fact{
		ID:         uuid.NewString(),
		Fact:       "User prefers async/await over callbacks",
		Category:   "preference",
		Confidence: 0.9,
		Source:     "episode:abc",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.SaveFact(ctx, fact); err != nil {
		t.Fatalf("SaveFact: %v", err)
	}

	got, err := s.GetFact(ctx, fact.ID)
	if err != nil {
		t.Fatalf("GetFact: %v", err)
	}

	if got.Fact != fact.Fact {
		t.Errorf("fact = %q, want %q", got.Fact, fact.Fact)
	}
	if got.Category != fact.Category {
		t.Errorf("category = %q, want %q", got.Category, fact.Category)
	}
}

func TestGetFactsByCategory(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	facts := []Fact{
		{ID: uuid.NewString(), Fact: "likes Go", Category: "preference", Confidence: 0.8, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.NewString(), Fact: "likes Python", Category: "preference", Confidence: 0.7, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.NewString(), Fact: "name is Dan", Category: "personal_info", Confidence: 0.9, CreatedAt: now, UpdatedAt: now},
	}

	for _, f := range facts {
		if err := s.SaveFact(ctx, f); err != nil {
			t.Fatalf("SaveFact: %v", err)
		}
	}

	got, err := s.GetFactsByCategory(ctx, "preference")
	if err != nil {
		t.Fatalf("GetFactsByCategory: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d facts, want 2", len(got))
	}
}

func TestUpdateFact(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	fact := Fact{
		ID:         uuid.NewString(),
		Fact:       "User likes Go",
		Category:   "preference",
		Confidence: 0.5,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.SaveFact(ctx, fact); err != nil {
		t.Fatalf("SaveFact: %v", err)
	}

	fact.Fact = "User really likes Go"
	fact.Confidence = 0.9
	fact.UpdatedAt = time.Now().UnixMilli()

	if err := s.UpdateFact(ctx, fact); err != nil {
		t.Fatalf("UpdateFact: %v", err)
	}

	got, err := s.GetFact(ctx, fact.ID)
	if err != nil {
		t.Fatalf("GetFact: %v", err)
	}

	if got.Fact != "User really likes Go" {
		t.Errorf("fact = %q, want %q", got.Fact, "User really likes Go")
	}
	if got.Confidence != 0.9 {
		t.Errorf("confidence = %f, want 0.9", got.Confidence)
	}
}

func TestDeleteFact(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	fact := Fact{
		ID: uuid.NewString(), Fact: "to delete", Category: "test",
		Confidence: 0.5, CreatedAt: now, UpdatedAt: now,
	}

	if err := s.SaveFact(ctx, fact); err != nil {
		t.Fatalf("SaveFact: %v", err)
	}

	if err := s.DeleteFact(ctx, fact.ID); err != nil {
		t.Fatalf("DeleteFact: %v", err)
	}

	_, err := s.GetFact(ctx, fact.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestSaveAndGetEntity(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entity := Entity{
		ID:          uuid.NewString(),
		Name:        "Go",
		Type:        "language",
		Description: "A statically typed compiled programming language",
		CreatedAt:   time.Now().UnixMilli(),
	}

	if err := s.SaveEntity(ctx, entity); err != nil {
		t.Fatalf("SaveEntity: %v", err)
	}

	got, err := s.GetEntity(ctx, entity.ID)
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}

	if got.Name != entity.Name {
		t.Errorf("name = %q, want %q", got.Name, entity.Name)
	}
	if got.Type != entity.Type {
		t.Errorf("type = %q, want %q", got.Type, entity.Type)
	}
}

func TestGetEntities(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entities := []Entity{
		{ID: uuid.NewString(), Name: "Go", Type: "language", CreatedAt: 1000},
		{ID: uuid.NewString(), Name: "Python", Type: "language", CreatedAt: 2000},
	}

	for _, e := range entities {
		if err := s.SaveEntity(ctx, e); err != nil {
			t.Fatalf("SaveEntity: %v", err)
		}
	}

	got, err := s.GetEntities(ctx)
	if err != nil {
		t.Fatalf("GetEntities: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d entities, want 2", len(got))
	}
}

func TestSaveAndGetRelationship(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	e1 := Entity{ID: uuid.NewString(), Name: "Dan", Type: "person", CreatedAt: now}
	e2 := Entity{ID: uuid.NewString(), Name: "Go", Type: "language", CreatedAt: now}

	if err := s.SaveEntity(ctx, e1); err != nil {
		t.Fatalf("SaveEntity: %v", err)
	}
	if err := s.SaveEntity(ctx, e2); err != nil {
		t.Fatalf("SaveEntity: %v", err)
	}

	rel := Relationship{
		ID:           uuid.NewString(),
		SourceEntity: e1.ID,
		TargetEntity: e2.ID,
		Relationship: "uses",
		Confidence:   0.9,
		CreatedAt:    now,
	}

	if err := s.SaveRelationship(ctx, rel); err != nil {
		t.Fatalf("SaveRelationship: %v", err)
	}

	rels, err := s.GetRelationships(ctx)
	if err != nil {
		t.Fatalf("GetRelationships: %v", err)
	}
	if len(rels) != 1 {
		t.Errorf("got %d relationships, want 1", len(rels))
	}
	if rels[0].Relationship != "uses" {
		t.Errorf("relationship = %q, want %q", rels[0].Relationship, "uses")
	}
}

func TestScratchpad(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.InitScratchpad(ctx); err != nil {
		t.Fatalf("InitScratchpad: %v", err)
	}

	content, err := s.GetScratchpad(ctx)
	if err != nil {
		t.Fatalf("GetScratchpad: %v", err)
	}
	if content != "" {
		t.Errorf("initial scratchpad = %q, want empty", content)
	}

	if err := s.UpdateScratchpad(ctx, "Remember: user likes Go"); err != nil {
		t.Fatalf("UpdateScratchpad: %v", err)
	}

	content, err = s.GetScratchpad(ctx)
	if err != nil {
		t.Fatalf("GetScratchpad: %v", err)
	}
	if content != "Remember: user likes Go" {
		t.Errorf("scratchpad = %q, want %q", content, "Remember: user likes Go")
	}
}

func TestActivityLog(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entry := ActivityEntry{
		ID:        uuid.NewString(),
		Timestamp: time.Now().UnixMilli(),
		Type:      "message_received",
		Details:   `{"content": "hello"}`,
		MessageID: "msg1",
		SessionID: "session1",
	}

	if err := s.LogActivity(ctx, entry); err != nil {
		t.Fatalf("LogActivity: %v", err)
	}

	entries, err := s.GetActivityLog(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("GetActivityLog: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("got %d entries, want 1", len(entries))
	}
	if entries[0].Type != "message_received" {
		t.Errorf("type = %q, want %q", entries[0].Type, "message_received")
	}
}

func TestActivityLog_Filtered(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	entries := []ActivityEntry{
		{ID: uuid.NewString(), Timestamp: now, Type: "message_received", Details: "{}"},
		{ID: uuid.NewString(), Timestamp: now + 1, Type: "llm_request", Details: "{}"},
		{ID: uuid.NewString(), Timestamp: now + 2, Type: "message_received", Details: "{}"},
	}

	for _, e := range entries {
		if err := s.LogActivity(ctx, e); err != nil {
			t.Fatalf("LogActivity: %v", err)
		}
	}

	got, err := s.GetActivityLog(ctx, "llm_request", 10, 0)
	if err != nil {
		t.Fatalf("GetActivityLog: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d entries, want 1", len(got))
	}
}

func TestGetFacts(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	for i := 0; i < 5; i++ {
		f := Fact{
			ID: uuid.NewString(), Fact: "fact " + string(rune('0'+i)),
			Category: "test", Confidence: 0.5, CreatedAt: now, UpdatedAt: now,
		}
		if err := s.SaveFact(ctx, f); err != nil {
			t.Fatalf("SaveFact: %v", err)
		}
	}

	facts, err := s.GetFacts(ctx, 3, 0)
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) != 3 {
		t.Errorf("got %d facts, want 3", len(facts))
	}
}

func TestDuplicateMessageID(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	msg := Message{
		ID: "dup-id", UserID: "u1", Role: "user",
		Content: "first", Timestamp: 1000, Interface: "gui",
	}

	if err := s.SaveMessage(ctx, msg); err != nil {
		t.Fatalf("first SaveMessage: %v", err)
	}

	msg.Content = "second"
	if err := s.SaveMessage(ctx, msg); err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
}

func TestLargeContent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	large := make([]byte, 100000)
	for i := range large {
		large[i] = 'a' + byte(i%26)
	}

	msg := Message{
		ID: uuid.NewString(), UserID: "u1", Role: "user",
		Content: string(large), Timestamp: 1000, Interface: "gui",
	}

	if err := s.SaveMessage(ctx, msg); err != nil {
		t.Fatalf("SaveMessage with large content: %v", err)
	}

	got, err := s.GetMessage(ctx, msg.ID)
	if err != nil {
		t.Fatalf("GetMessage: %v", err)
	}
	if len(got.Content) != len(large) {
		t.Errorf("content length = %d, want %d", len(got.Content), len(large))
	}
}

func TestEmptyResults(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	messages, err := s.GetMessages(ctx, 10, 0)
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("got %d messages, want 0", len(messages))
	}

	episodes, err := s.GetEpisodes(ctx, 10, 0)
	if err != nil {
		t.Fatalf("GetEpisodes: %v", err)
	}
	if len(episodes) != 0 {
		t.Errorf("got %d episodes, want 0", len(episodes))
	}

	facts, err := s.GetFacts(ctx, 10, 0)
	if err != nil {
		t.Fatalf("GetFacts: %v", err)
	}
	if len(facts) != 0 {
		t.Errorf("got %d facts, want 0", len(facts))
	}
}
