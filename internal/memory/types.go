package memory

type Message struct {
	ID        string
	UserID    string
	Role      string
	Content   string
	Timestamp int64
	Interface string
	SessionID string
}

type Episode struct {
	ID         string
	Summary    string
	StartTime  int64
	EndTime    int64
	MessageIDs string
	Importance float64
	Topics     string
}

type Fact struct {
	ID         string
	Fact       string
	Category   string
	Confidence float64
	Source     string
	CreatedAt  int64
	UpdatedAt  int64
}

type Entity struct {
	ID          string
	Name        string
	Type        string
	Description string
	CreatedAt   int64
}

type Relationship struct {
	ID           string
	SourceEntity string
	TargetEntity string
	Relationship string
	Confidence   float64
	CreatedAt    int64
}

type ActivityEntry struct {
	ID        string
	Timestamp int64
	Type      string
	Details   string
	MessageID string
	SessionID string
}
