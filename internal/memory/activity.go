package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func (s *Store) LogActivity(ctx context.Context, entry *ActivityEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("activity_log").
		Columns("id", "timestamp", "type", "details", "message_id", "session_id").
		Values(entry.ID, entry.Timestamp, entry.Type, entry.Details, entry.MessageID, entry.SessionID).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("logging activity: %w", err)
	}
	return nil
}

func (s *Store) GetActivityLog(ctx context.Context, filter string, limit, offset int) ([]ActivityEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b := sb.Select("id", "timestamp", "type", "details", "message_id", "session_id").
		From("activity_log").
		OrderBy("timestamp DESC").
		Limit(uint64(limit)).  //nolint:gosec // limit from user input, safe for test usage
		Offset(uint64(offset)) //nolint:gosec // offset from user input, safe for test usage

	if filter != "" {
		b = b.Where(sq.Eq{"type": filter})
	}

	query, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying activity log: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []ActivityEntry
	for rows.Next() {
		var e ActivityEntry
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Type, &e.Details, &e.MessageID, &e.SessionID); err != nil {
			return nil, fmt.Errorf("scanning activity row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
