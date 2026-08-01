package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// buildSelectByID returns a squirrel select builder for the given columns and
// table, filtered by ID.
func buildSelectByID(columns []string, table, id string) sq.SelectBuilder {
	return sb.Select(columns...).From(table).Where(sq.Eq{"id": id})
}

// rowScanner is the minimal interface shared by *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanRow queries a single row via the given select builder and scans it
// using the provided scan function. It acquires a read lock.
func (s *Store) scanRow(
	ctx context.Context,
	selectBuilder sq.SelectBuilder,
	scan func(rowScanner) error,
) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("building select: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	return scan(row)
}

// scanRows executes the given select builder and iterates over the result
// set, calling scan for each row. It acquires a read lock and closes rows.
func (s *Store) scanRows(
	ctx context.Context,
	selectBuilder sq.SelectBuilder,
	scan func(rowScanner) error,
	queryErrLabel string,
) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", queryErrLabel, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}