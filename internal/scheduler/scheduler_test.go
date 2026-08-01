package scheduler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/danmurf/remy/internal/memory"
	"github.com/danmurf/remy/internal/scheduler"
	"github.com/danmurf/remy/internal/scheduler/mock_scheduler"
)

func TestNewScheduler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	s := scheduler.NewScheduler(store, agent)
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestScheduler_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().SaveTask(gomock.Any(), gomock.Any()).Do(func(_ context.Context, task *memory.Task) {
		if task.ID == "" {
			task.ID = uuid.NewString()
		}
	}).Return(nil)

	s := scheduler.NewScheduler(store, agent)
	triggerAt := time.Now().Add(1 * time.Hour).UnixMilli()
	task, err := s.CreateTask(context.Background(), "reminder", `{"type":"send_message","text":"Call dentist"}`, triggerAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
	if task.Type != "reminder" {
		t.Errorf("type = %q, want %q", task.Type, "reminder")
	}
	if task.TriggerAt != triggerAt {
		t.Errorf("trigger_at = %d, want %d", task.TriggerAt, triggerAt)
	}
	if task.Status != "pending" {
		t.Errorf("status = %q, want %q", task.Status, "pending")
	}
}

func TestScheduler_CreateTask_StoreError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().SaveTask(gomock.Any(), gomock.Any()).Return(errors.New("database error"))

	s := scheduler.NewScheduler(store, agent)
	_, err := s.CreateTask(context.Background(), "reminder", `{}`, time.Now().UnixMilli())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestScheduler_CreateRecurringTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().SaveTask(gomock.Any(), gomock.Any()).Do(func(_ context.Context, task *memory.Task) {
		if task.ID == "" {
			task.ID = uuid.NewString()
		}
	}).Return(nil)

	s := scheduler.NewScheduler(store, agent)
	task, err := s.CreateRecurringTask(context.Background(), "scheduled_message", `{"type":"send_message","text":"Good morning!"}`, "0 8 * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected non-nil task")
	}
	if task.Type != "scheduled_message" {
		t.Errorf("type = %q, want %q", task.Type, "scheduled_message")
	}
	if task.CronExpr != "0 8 * * *" {
		t.Errorf("cron_expr = %q, want %q", task.CronExpr, "0 8 * * *")
	}
	if task.Status != "pending" {
		t.Errorf("status = %q, want %q", task.Status, "pending")
	}
	if task.TriggerAt <= time.Now().UnixMilli() {
		t.Error("trigger_at should be in the future")
	}
}

func TestScheduler_CreateRecurringTask_InvalidCron(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	s := scheduler.NewScheduler(store, agent)
	_, err := s.CreateRecurringTask(context.Background(), "scheduled_message", `{}`, "invalid cron")
	if err == nil {
		t.Fatal("expected error for invalid cron, got nil")
	}
}

func TestScheduler_CancelTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().CancelTask(gomock.Any(), "task-123").Return(nil)

	s := scheduler.NewScheduler(store, agent)
	if err := s.CancelTask(context.Background(), "task-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScheduler_GetTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().GetTasks(gomock.Any(), "pending").Return([]memory.Task{
		{ID: uuid.NewString(), Type: "reminder", Status: "pending", TriggerAt: time.Now().Add(1 * time.Hour).UnixMilli(), Action: `{"type":"send_message","text":"Test"}`},
	}, nil)

	s := scheduler.NewScheduler(store, agent)
	tasks, err := s.GetTasks(context.Background(), "pending")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
}

func TestScheduler_GetUpcomingTasks_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().GetTasks(gomock.Any(), "pending").Return(nil, nil)

	s := scheduler.NewScheduler(store, agent)
	summary, err := s.GetUpcomingTasks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "" {
		t.Errorf("expected empty summary, got %q", summary)
	}
}

func TestScheduler_GetUpcomingTasks_WithTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().GetTasks(gomock.Any(), "pending").Return([]memory.Task{
		{
			ID:        uuid.NewString(),
			Type:      "reminder",
			Status:    "pending",
			TriggerAt: time.Now().Add(2 * time.Hour).UnixMilli(),
			Action:    `{"type":"send_message","text":"Call dentist"}`,
		},
		{
			ID:        uuid.NewString(),
			Type:      "scheduled_message",
			Status:    "pending",
			TriggerAt: time.Now().Add(8 * time.Hour).UnixMilli(),
			CronExpr:  "0 8 * * *",
			Action:    `{"type":"send_message","text":"Good morning!"}`,
		},
	}, nil)

	s := scheduler.NewScheduler(store, agent)
	summary, err := s.GetUpcomingTasks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestScheduler_GetUpcomingTasks_StoreError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().GetTasks(gomock.Any(), "pending").Return(nil, errors.New("db error"))

	s := scheduler.NewScheduler(store, agent)
	_, err := s.GetUpcomingTasks(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestScheduler_StartAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	s := scheduler.NewScheduler(store, agent)
	s.Start(context.Background())
	s.Stop()
}

func TestScheduler_FireDueTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	now := time.Now().UnixMilli()
	taskID := uuid.NewString()

	store.EXPECT().GetDueTasks(gomock.Any(), gomock.Any()).Return([]memory.Task{
		{
			ID:        taskID,
			Type:      "reminder",
			Status:    "pending",
			TriggerAt: now - 1000,
			Action:    `{"type":"send_message","text":"Test reminder"}`,
		},
	}, nil)

	agent.EXPECT().HandleMessage(gomock.Any(), "Reminder: Test reminder").Return(&memory.Message{
		ID: uuid.NewString(), Role: "assistant", Content: "OK", Timestamp: now,
	}, nil)

	store.EXPECT().UpdateTaskStatus(gomock.Any(), taskID, "fired", gomock.Any()).Return(nil)

	s := scheduler.NewScheduler(store, agent)
	s.FireDueTasks(context.Background())
}

func TestScheduler_FireDueTasks_Recurring(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	now := time.Now().UnixMilli()
	taskID := uuid.NewString()

	store.EXPECT().GetDueTasks(gomock.Any(), gomock.Any()).Return([]memory.Task{
		{
			ID:        taskID,
			Type:      "scheduled_message",
			Status:    "pending",
			TriggerAt: now - 1000,
			CronExpr:  "0 8 * * *",
			Action:    `{"type":"send_message","text":"Daily briefing"}`,
		},
	}, nil)

	agent.EXPECT().HandleMessage(gomock.Any(), "Scheduled message: Daily briefing").Return(&memory.Message{
		ID: uuid.NewString(), Role: "assistant", Content: "OK", Timestamp: now,
	}, nil)

	store.EXPECT().UpdateTaskStatus(gomock.Any(), taskID, "fired", gomock.Any()).Return(nil)

	store.EXPECT().SaveTask(gomock.Any(), gomock.Any()).Do(func(_ context.Context, task *memory.Task) {
		if task.ID == "" {
			task.ID = uuid.NewString()
		}
	}).Return(nil)

	s := scheduler.NewScheduler(store, agent)
	s.FireDueTasks(context.Background())
}

func TestScheduler_FireDueTasks_NoTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().GetDueTasks(gomock.Any(), gomock.Any()).Return(nil, nil)

	s := scheduler.NewScheduler(store, agent)
	s.FireDueTasks(context.Background())
}

func TestScheduler_FireDueTasks_AgentError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	now := time.Now().UnixMilli()

	store.EXPECT().GetDueTasks(gomock.Any(), gomock.Any()).Return([]memory.Task{
		{
			ID:        uuid.NewString(),
			Type:      "reminder",
			Status:    "pending",
			TriggerAt: now - 1000,
			Action:    `{"type":"send_message","text":"Test"}`,
		},
	}, nil)

	agent.EXPECT().HandleMessage(gomock.Any(), gomock.Any()).Return(nil, errors.New("agent error"))

	s := scheduler.NewScheduler(store, agent)
	s.FireDueTasks(context.Background())
}

func TestScheduler_FireDueTasks_GetDueTasksError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	store.EXPECT().GetDueTasks(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	s := scheduler.NewScheduler(store, agent)
	s.FireDueTasks(context.Background())
}

func TestScheduler_FireDueTasks_InvalidActionJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	now := time.Now().UnixMilli()

	store.EXPECT().GetDueTasks(gomock.Any(), gomock.Any()).Return([]memory.Task{
		{
			ID:        uuid.NewString(),
			Type:      "reminder",
			Status:    "pending",
			TriggerAt: now - 1000,
			Action:    "not valid json",
		},
	}, nil)

	s := scheduler.NewScheduler(store, agent)
	s.FireDueTasks(context.Background())
}

func TestScheduler_FireDueTasks_UnknownType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock_scheduler.NewMockStore(ctrl)
	agent := mock_scheduler.NewMockAgent(ctrl)

	now := time.Now().UnixMilli()

	store.EXPECT().GetDueTasks(gomock.Any(), gomock.Any()).Return([]memory.Task{
		{
			ID:        uuid.NewString(),
			Type:      "unknown_type",
			Status:    "pending",
			TriggerAt: now - 1000,
			Action:    `{"type":"send_message","text":"Test"}`,
		},
	}, nil)

	s := scheduler.NewScheduler(store, agent)
	s.FireDueTasks(context.Background())
}
