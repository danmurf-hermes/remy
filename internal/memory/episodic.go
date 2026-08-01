package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

var episodeColumns = []string{"id", "summary", "start_time", "end_time", "message_ids", "importance", "topics"}

func scanEpisode(row rowScanner) (Episode, error) {
	var ep Episode
	if err := row.Scan(&ep.ID, &ep.Summary, &ep.StartTime, &ep.EndTime, &ep.MessageIDs, &ep.Importance, &ep.Topics); err != nil {
		return Episode{}, err
	}
	return ep, nil
}

// SaveEpisode inserts a new episode into the database.
func (s *Store) SaveEpisode(ctx context.Context, ep *Episode) error {
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

// GetEpisode retrieves a single episode by its ID.
func (s *Store) GetEpisode(ctx context.Context, id string) (*Episode, error) {
	var ep Episode
	if err := s.scanRow(ctx, buildSelectByID(episodeColumns, "episodes", id), func(row rowScanner) error {
		var e Episode
		if err := row.Scan(&e.ID, &e.Summary, &e.StartTime, &e.EndTime, &e.MessageIDs, &e.Importance, &e.Topics); err != nil {
			return fmt.Errorf("getting episode: %w", err)
		}
		ep = e
		return nil
	}); err != nil {
		return nil, err
	}
	return &ep, nil
}

// GetEpisodes returns a paginated list of all episodes, ordered by most
// recent first.
func (s *Store) GetEpisodes(ctx context.Context, limit, offset int) ([]Episode, error) {
	qb := sb.Select(episodeColumns...).
		From("episodes").
		OrderBy("end_time DESC").
		Limit(uint64(limit)).   //nolint:gosec // limit from user input, safe for test usage
		Offset(uint64(offset)). //nolint:gosec // offset from user input, safe for test usage

	var episodes []Episode
	if err := s.scanRows(ctx, qb, func(row rowScanner) error {
		ep, err := scanEpisode(row)
		if err != nil {
			return fmt.Errorf("scanning episode row: %w", err)
		}
		episodes = append(episodes, ep)
	}, "querying episodes"); err != nil {
		return nil, err
	}
	return episodes, nil
}

// GetEpisodesByTimeRange returns all episodes whose time range falls within
// the given start and end timestamps.
func (s *Store) GetEpisodesByTimeRange(ctx context.Context, start, end int64) ([]Episode, error) {
	qb := sb.Select(episodeColumns...).
		From("episodes").
		Where(sq.GtOrEq{"start_time": start}).
		Where(sq.LtOrEq{"end_time": end}).
		OrderBy("end_time DESC")

	var episodes []Episode
	if err := s.scanRows(ctx, qb, func(row rowScanner) error {
		ep, err := scanEpisode(row)
		if err != nil {
			return fmt.Errorf("scanning episode row: %w", err)
		}
		episodes = append(episodes, ep)
	}, "querying episodes by time range"); err != nil {
		return nil, err
	}
	return episodes, nil
}

// SearchEpisodes performs a vector similarity search over episodes using the
// given embedding, returning the top-N most similar episodes.
func (s *Store) SearchEpisodes(ctx context.Context, embedding []byte, limit int) ([]Episode, error) {
	qb := sb.Select("e.id", "e.summary", "e.start_time", "e.end_time", "e.message_ids", "e.importance", "e.topics").
		From("episodes e").
		JoinClause(sq.Expr("INNER JOIN (SELECT id, distance FROM episode_vectors WHERE embedding MATCH ? ORDER BY distance LIMIT ?) ev ON ev.id = e.id", embedding, limit))

	var episodes []Episode
	if err := s.scanRows(ctx, qb, func(row rowScanner) error {
		ep, err := scanEpisode(row)
		if err != nil {
			return fmt.Errorf("scanning episode row: %w", err)
		}
		episodes = append(episodes, ep)
	}, "searching episodes"); err != nil {
		return nil, err
	}
	return episodes, nil
}