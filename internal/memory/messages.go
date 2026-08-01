package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

var messageColumns = []string{"id", "user_id", "role", "content", "timestamp", "interface", "session_id"}

func scanMessage(row rowScanner) (Message, error) {
	var msg Message
	if err := row.Scan(&msg.ID, &msg.UserID, &msg.Role, &msg.Content, &msg.Timestamp, &msg.Interface, &msg.SessionID); err != nil {
		return Message{}, err
	}
	return msg, nil
}

// SaveMessage inserts a new message into the database.
func (s *Store) SaveMessage(ctx context.Context, msg *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("messages").
		Columns("id", "user_id", "role", "content", "timestamp", "interface", "session_id").
		Values(msg.ID, msg.UserID, msg.Role, msg.Content, msg.Timestamp, msg.Interface, msg.SessionID).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving message: %w", err)
	}
	return nil
}

// GetMessage retrieves a single message by its ID.
func (s *Store) GetMessage(ctx context.Context, id string) (*Message, error) {
	var msg Message
	if err := s.scanRow(ctx, buildSelectByID(messageColumns, "messages", id), func(row rowScanner) error {
		var m Message
		if err := row.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.Timestamp, &m.Interface, &m.SessionID); err != nil {
			return fmt.Errorf("getting message: %w", err)
		}
		msg = m
		return nil
	}); err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetMessages returns a paginated list of all messages, ordered by most
// recent first.
func (s *Store) GetMessages(ctx context.Context, limit, offset int) ([]Message, error) {
	qb := sb.Select(messageColumns...).
		From("messages").
		OrderBy("timestamp DESC").
		Limit(uint64(limit)).  //nolint:gosec // limit from user input
		Offset(uint64(offset)) //nolint:gosec // offset from user input

	var messages []Message
	if err := s.scanRows(ctx, qb, func(row rowScanner) error {
		msg, err := scanMessage(row)
		if err != nil {
			return fmt.Errorf("scanning message row: %w", err)
		}
		messages = append(messages, msg)
	}, "querying messages"); err != nil {
		return nil, err
	}
	return messages, nil
}

// GetMessagesBySession returns all messages for a given session ID, ordered
// by timestamp ascending.
func (s *Store) GetMessagesBySession(ctx context.Context, sessionID string) ([]Message, error) {
	qb := sb.Select(messageColumns...).
		From("messages").
		Where(sq.Eq{"session_id": sessionID}).
		OrderBy("timestamp ASC")

	var messages []Message
	if err := s.scanRows(ctx, qb, func(row rowScanner) error {
		msg, err := scanMessage(row)
		if err != nil {
			return fmt.Errorf("scanning message row: %w", err)
		}
		messages = append(messages, msg)
	}, "querying messages by session"); err != nil {
		return nil, err
	}
	return messages, nil
}