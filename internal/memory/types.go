package memory

// Message represents a single message exchanged between a user and the
// agent, stored in the database.
type Message struct {
	ID        string
	UserID    string
	Role      string
	Content   string
	Timestamp int64
	Interface string
	SessionID string
}

// Episode is a summarized block of conversation, stored as episodic memory.
type Episode struct {
	ID         string
	Summary    string
	StartTime  int64
	EndTime    int64
	MessageIDs string
	Importance float64
	Topics     string
}

// Fact is a piece of semantic knowledge about the user (preferences,
// traits, personal info) extracted from conversations.
type Fact struct {
	ID         string
	Fact       string
	Category   string
	Confidence float64
	Source     string
	CreatedAt  int64
	UpdatedAt  int64
}

// Entity represents a person, place, concept, or thing that the agent
// knows about.
type Entity struct {
	ID          string
	Name        string
	Type        string
	Description string
	CreatedAt   int64
}

// Relationship describes a link between two entities (e.g., "works_on",
// "lives_in").
type Relationship struct {
	ID           string
	SourceEntity string
	TargetEntity string
	Relationship string
	Confidence   float64
	CreatedAt    int64
}

// ActivityEntry is a log entry recording an agent action or decision
// for audit and debugging purposes.
type ActivityEntry struct {
	ID        string
	Timestamp int64
	Type      string
	Details   string
	MessageID string
	SessionID string
}

// Task represents a scheduled action — either a one-shot reminder or a
// recurring task. Tasks are stored in the database and checked periodically
// by the scheduler.
type Task struct {
	ID        string
	Type      string // "reminder" | "scheduled_message"
	Status    string // "pending" | "fired" | "canceled"
	TriggerAt int64  // Unix timestamp for one-shot
	CronExpr  string // Cron expression for recurring (e.g., "0 8 * * *")
	Action    string // JSON describing the action
	Context   string // JSON context to inject when firing
	CreatedAt int64
	FiredAt   int64
}
