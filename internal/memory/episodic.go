package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func (s *Store) SaveEpisode(ctx context.Context, ep Episode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("episodes").
		Columns("id", "summary", "start_time", "end_time", "message_ids", "importance", "topics").
		Values(ep.ID, ep.Summary, ep.StartTime, ep.EndTime, ep.MessageIDs, ep.Importance, ep.Topics).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving episode: %w", err)
	}
	return nil
}

func (s *Store) GetEpisode(ctx context.Context, id string) (*Episode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "summary", "start_time", "end_time", "message_ids", "importance", "topics").
		From("episodes").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	var ep Episode
	if err := row.Scan(&ep.ID, &ep.Summary, &ep.StartTime, &ep.EndTime, &ep.MessageIDs, &ep.Importance, &ep.Topics); err != nil {
		return nil, fmt.Errorf("getting episode: %w", err)
	}
	return &ep, nil
}

func (s *Store) GetEpisodes(ctx context.Context, limit, offset int) ([]Episode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "summary", "start_time", "end_time", "message_ids", "importance", "topics").
		From("episodes").
		OrderBy("end_time DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying episodes: %w", err)
	}
	defer rows.Close()

	var episodes []Episode
	for rows.Next() {
		var ep Episode
		if err := rows.Scan(&ep.ID, &ep.Summary, &ep.StartTime, &ep.EndTime, &ep.MessageIDs, &ep.Importance, &ep.Topics); err != nil {
			return nil, fmt.Errorf("scanning episode row: %w", err)
		}
		episodes = append(episodes, ep)
	}
	return episodes, rows.Err()
}

func (s *Store) GetEpisodesByTimeRange(ctx context.Context, start, end int64) ([]Episode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "summary", "start_time", "end_time", "message_ids", "importance", "topics").
		From("episodes").
		Where(sq.GtOrEq{"start_time": start}).
		Where(sq.LtOrEq{"end_time": end}).
		OrderBy("end_time DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying episodes by time range: %w", err)
	}
	defer rows.Close()

	var episodes []Episode
	for rows.Next() {
		var ep Episode
		if err := rows.Scan(&ep.ID, &ep.Summary, &ep.StartTime, &ep.EndTime, &ep.MessageIDs, &ep.Importance, &ep.Topics); err != nil {
			return nil, fmt.Errorf("scanning episode row: %w", err)
		}
		episodes = append(episodes, ep)
	}
	return episodes, rows.Err()
}

func (s *Store) SearchEpisodes(ctx context.Context, embedding []byte, limit int) ([]Episode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("e.id", "e.summary", "e.start_time", "e.end_time", "e.message_ids", "e.importance", "e.topics").
		From("episodes e").
		JoinClause(sq.Expr("INNER JOIN (SELECT id, distance FROM episode_vectors WHERE embedding MATCH ? ORDER BY distance LIMIT ?) ev ON ev.id = e.id", embedding, limit)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building search query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("searching episodes: %w", err)
	}
	defer rows.Close()

	var episodes []Episode
	for rows.Next() {
		var ep Episode
		if err := rows.Scan(&ep.ID, &ep.Summary, &ep.StartTime, &ep.EndTime, &ep.MessageIDs, &ep.Importance, &ep.Topics); err != nil {
			return nil, fmt.Errorf("scanning episode row: %w", err)
		}
		episodes = append(episodes, ep)
	}
	return episodes, rows.Err()
}
