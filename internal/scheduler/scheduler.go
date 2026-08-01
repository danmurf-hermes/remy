// Package scheduler implements a task scheduling engine for Remy. It manages
// one-shot reminders and recurring tasks, checking for due tasks on a periodic
// tick and firing them through the agent.
package scheduler

//go:generate go run go.uber.org/mock/mockgen -destination=mock_scheduler/mock_scheduler.go -package=mock_scheduler . Store,Agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/danmurf/remy/internal/memory"
)

// Store defines the subset of memory.Store methods the scheduler needs,
// making it easy to mock in tests.
type Store interface {
	SaveTask(ctx context.Context, task *memory.Task) error
	GetTask(ctx context.Context, id string) (*memory.Task, error)
	GetTasks(ctx context.Context, status string) ([]memory.Task, error)
	GetDueTasks(ctx context.Context, now int64) ([]memory.Task, error)
	UpdateTaskStatus(ctx context.Context, id, status string, firedAt int64) error
	CancelTask(ctx context.Context, id string) error
}

// Agent defines the interface the scheduler needs to fire tasks through
// the agent's message handling pipeline.
type Agent interface {
	HandleMessage(ctx context.Context, userMsg string) (*memory.Message, error)
}

// Scheduler manages task creation, cancellation, and firing. It runs a
// background loop that checks for due tasks on a configurable interval.
type Scheduler struct {
	store         Store
	agent         Agent
	stopCh        chan struct{}
	checkInterval time.Duration
}

// NewScheduler creates a new Scheduler with the given store and agent.
// The check interval defaults to 30 seconds.
func NewScheduler(store Store, agent Agent) *Scheduler {
	return &Scheduler{
		store:         store,
		agent:         agent,
		stopCh:        make(chan struct{}),
		checkInterval: 30 * time.Second,
	}
}

// NewSchedulerWithInterval creates a new Scheduler with a custom check interval.
func NewSchedulerWithInterval(store Store, agent Agent, interval time.Duration) *Scheduler {
	return &Scheduler{
		store:         store,
		agent:         agent,
		stopCh:        make(chan struct{}),
		checkInterval: interval,
	}
}

// Start begins the background scheduler loop. It checks for due tasks
// every 30 seconds and fires them. Call Stop to terminate the loop.
func (s *Scheduler) Start(ctx context.Context) {
	go s.loop(ctx)
}

// Stop signals the background loop to terminate.
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.FireDueTasks(ctx)
		}
	}
}

// FireDueTasks checks for and fires all due tasks. This is called
// periodically by the background loop but can also be called directly.
func (s *Scheduler) FireDueTasks(ctx context.Context) {
	now := time.Now().UnixMilli()
	tasks, err := s.store.GetDueTasks(ctx, now)
	if err != nil || len(tasks) == 0 {
		return
	}

	for i := range tasks {
		s.fireTask(ctx, &tasks[i])
	}
}

func (s *Scheduler) fireTask(ctx context.Context, task *memory.Task) {
	var action struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(task.Action), &action); err != nil {
		return
	}

	// Build a synthetic user message that triggers the reminder
	var userMsg string
	switch task.Type {
	case "reminder":
		userMsg = fmt.Sprintf("Reminder: %s", action.Text)
	case "scheduled_message":
		userMsg = fmt.Sprintf("Scheduled message: %s", action.Text)
	default:
		return
	}

	// Fire through the agent
	if _, err := s.agent.HandleMessage(ctx, userMsg); err != nil {
		return
	}

	// Mark as fired
	_ = s.store.UpdateTaskStatus(ctx, task.ID, "fired", time.Now().UnixMilli())

	// If recurring, calculate next occurrence and create a new task
	if task.CronExpr != "" {
		s.scheduleNext(ctx, task)
	}
}

func (s *Scheduler) scheduleNext(ctx context.Context, task *memory.Task) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	sched, err := parser.Parse(task.CronExpr)
	if err != nil {
		return
	}

	next := sched.Next(time.Now())
	newTask := &memory.Task{
		ID:        uuid.NewString(),
		Type:      task.Type,
		Status:    "pending",
		TriggerAt: next.UnixMilli(),
		CronExpr:  task.CronExpr,
		Action:    task.Action,
		Context:   task.Context,
		CreatedAt: time.Now().UnixMilli(),
	}
	_ = s.store.SaveTask(ctx, newTask)
}

// CreateTask creates a new one-shot reminder task.
func (s *Scheduler) CreateTask(ctx context.Context, taskType, action string, triggerAt int64) (*memory.Task, error) {
	task := &memory.Task{
		ID:        uuid.NewString(),
		Type:      taskType,
		Status:    "pending",
		TriggerAt: triggerAt,
		Action:    action,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := s.store.SaveTask(ctx, task); err != nil {
		return nil, fmt.Errorf("saving task: %w", err)
	}
	return task, nil
}

// CreateRecurringTask creates a new recurring task with a cron expression.
func (s *Scheduler) CreateRecurringTask(ctx context.Context, taskType, action, cronExpr string) (*memory.Task, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	sched, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("parsing cron expression: %w", err)
	}

	next := sched.Next(time.Now())
	task := &memory.Task{
		ID:        uuid.NewString(),
		Type:      taskType,
		Status:    "pending",
		TriggerAt: next.UnixMilli(),
		CronExpr:  cronExpr,
		Action:    action,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := s.store.SaveTask(ctx, task); err != nil {
		return nil, fmt.Errorf("saving recurring task: %w", err)
	}
	return task, nil
}

// CancelTask cancels a task by its ID.
func (s *Scheduler) CancelTask(ctx context.Context, id string) error {
	return s.store.CancelTask(ctx, id)
}

// GetTasks returns all tasks with the given status. Pass empty string for all.
func (s *Scheduler) GetTasks(ctx context.Context, status string) ([]memory.Task, error) {
	return s.store.GetTasks(ctx, status)
}

// GetUpcomingTasks returns a human-readable summary of upcoming pending tasks
// for inclusion in the system prompt.
func (s *Scheduler) GetUpcomingTasks(ctx context.Context) (string, error) {
	tasks, err := s.store.GetTasks(ctx, "pending")
	if err != nil {
		return "", fmt.Errorf("getting upcoming tasks: %w", err)
	}

	if len(tasks) == 0 {
		return "", nil
	}

	var lines []string
	now := time.Now()
	for _, task := range tasks {
		triggerTime := time.UnixMilli(task.TriggerAt)
		remaining := triggerTime.Sub(now).Round(time.Second)
		if remaining < 0 {
			remaining = 0
		}

		var actionDesc string
		var action struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(task.Action), &action); err == nil && action.Text != "" {
			actionDesc = action.Text
		} else {
			actionDesc = task.Action
		}

		if task.CronExpr != "" {
			lines = append(lines, fmt.Sprintf("- [%s] %s (recurring: %s, next in %s)", task.Type, actionDesc, task.CronExpr, remaining))
		} else {
			lines = append(lines, fmt.Sprintf("- [%s] %s (in %s)", task.Type, actionDesc, remaining))
		}
	}

	return strings.Join(lines, "\n"), nil
}
