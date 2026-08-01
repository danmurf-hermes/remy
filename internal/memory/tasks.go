package memory

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// SaveTask inserts a new task into the database.
func (s *Store) SaveTask(ctx context.Context, task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, args, err := sb.Insert("tasks").
		Columns("id", "type", "status", "trigger_at", "cron_expr", "action", "context", "created_at", "fired_at").
		Values(task.ID, task.Type, task.Status, task.TriggerAt, task.CronExpr, task.Action, task.Context, task.CreatedAt, task.FiredAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("saving task: %w", err)
	}
	return nil
}

// GetTask retrieves a single task by its ID.
func (s *Store) GetTask(ctx context.Context, id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "type", "status", "trigger_at", "cron_expr", "action", "context", "created_at", "fired_at").
		From("tasks").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	var t Task
	if err := row.Scan(&t.ID, &t.Type, &t.Status, &t.TriggerAt, &t.CronExpr, &t.Action, &t.Context, &t.CreatedAt, &t.FiredAt); err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}
	return &t, nil
}

// GetTasks returns tasks filtered by status, ordered by trigger_at ascending.
func (s *Store) GetTasks(ctx context.Context, status string) ([]Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b := sb.Select("id", "type", "status", "trigger_at", "cron_expr", "action", "context", "created_at", "fired_at").
		From("tasks").
		OrderBy("trigger_at ASC")

	if status != "" {
		b = b.Where(sq.Eq{"status": status})
	}

	query, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Type, &t.Status, &t.TriggerAt, &t.CronExpr, &t.Action, &t.Context, &t.CreatedAt, &t.FiredAt); err != nil {
			return nil, fmt.Errorf("scanning task row: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetDueTasks returns all pending tasks whose trigger_at is <= the given timestamp.
func (s *Store) GetDueTasks(ctx context.Context, now int64) ([]Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, args, err := sb.Select("id", "type", "status", "trigger_at", "cron_expr", "action", "context", "created_at", "fired_at").
		From("tasks").
		Where(sq.Eq{"status": "pending"}).
		Where(sq.LtOrEq{"trigger_at": now}).
		OrderBy("trigger_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying due tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Type, &t.Status, &t.TriggerAt, &t.CronExpr, &t.Action, &t.Context, &t.CreatedAt, &t.FiredAt); err != nil {
			return nil, fmt.Errorf("scanning task row: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// UpdateTaskStatus updates the status and optionally fired_at of a task.
func (s *Store) UpdateTaskStatus(ctx context.Context, id, status string, firedAt int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := sb.Update("tasks").
		Set("status", status).
		Where(sq.Eq{"id": id})

	if firedAt > 0 {
		b = b.Set("fired_at", firedAt)
	}

	query, args, err := b.ToSql()
	if err != nil {
		return fmt.Errorf("building update: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating task status: %w", err)
	}
	return nil
}

// CancelTask sets a task's status to "cancelled".
func (s *Store) CancelTask(ctx context.Context, id string) error {
	return s.UpdateTaskStatus(ctx, id, "cancelled", 0)
}
