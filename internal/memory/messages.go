package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func (s *Store) SaveMessage(ctx context.Context, msg Message) error {
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

func (s *Store) GetMessage(ctx context.Context, id string) (*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "user_id", "role", "content", "timestamp", "interface", "session_id").
		From("messages").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	var msg Message
	if err := row.Scan(&msg.ID, &msg.UserID, &msg.Role, &msg.Content, &msg.Timestamp, &msg.Interface, &msg.SessionID); err != nil {
		return nil, fmt.Errorf("getting message: %w", err)
	}
	return &msg, nil
}

func (s *Store) GetMessages(ctx context.Context, limit, offset int) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "user_id", "role", "content", "timestamp", "interface", "session_id").
		From("messages").
		OrderBy("timestamp DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Role, &msg.Content, &msg.Timestamp, &msg.Interface, &msg.SessionID); err != nil {
			return nil, fmt.Errorf("scanning message row: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (s *Store) GetMessagesBySession(ctx context.Context, sessionID string) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "user_id", "role", "content", "timestamp", "interface", "session_id").
		From("messages").
		Where(sq.Eq{"session_id": sessionID}).
		OrderBy("timestamp ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying messages by session: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Role, &msg.Content, &msg.Timestamp, &msg.Interface, &msg.SessionID); err != nil {
			return nil, fmt.Errorf("scanning message row: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}
