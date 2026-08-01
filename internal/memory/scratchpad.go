package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func (s *Store) GetScratchpad(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("content").
		From("scratchpad").
		Where(sq.Eq{"id": "default"}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("building select: %w", err)
	}

	var content string
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&content)
	if err != nil {
		return "", fmt.Errorf("getting scratchpad: %w", err)
	}
	return content, nil
}

func (s *Store) UpdateScratchpad(ctx context.Context, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("scratchpad").
		Columns("id", "content", "updated_at").
		Values("default", content, now()).
		Suffix("ON CONFLICT(id) DO UPDATE SET content = ?, updated_at = ?", content, now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("building upsert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating scratchpad: %w", err)
	}
	return nil
}

func (s *Store) InitScratchpad(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("scratchpad").
		Columns("id", "content", "updated_at").
		Values("default", "", now()).
		Suffix("ON CONFLICT(id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("initializing scratchpad: %w", err)
	}
	return nil
}
