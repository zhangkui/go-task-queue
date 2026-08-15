package service

import (
	"context"
	"testing"

	"go-task-queue/internal/model"
	"go-task-queue/internal/store"
)

func TestSubmitAndGetTask(t *testing.T) {
	svc := NewQueueService(store.NewMemoryStore())
	task, err := svc.SubmitTask(context.Background(), "t1", "test", "payload", 3)
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if task.Status != model.StatusPending {
		t.Fatalf("expected pending, got %s", task.Status)
	}
	got, err := svc.GetTask(context.Background(), "t1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Name != "test" {
		t.Fatalf("expected test, got %s", got.Name)
	}
}

func TestExecuteTask(t *testing.T) {
	svc := NewQueueService(store.NewMemoryStore())
	svc.SubmitTask(context.Background(), "t1", "test", "payload", 3)
	if err := svc.ExecuteTask(context.Background(), "t1"); err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	task, _ := svc.GetTask(context.Background(), "t1")
	if task.Status != model.StatusCompleted {
		t.Fatalf("expected completed, got %s", task.Status)
	}
}

func TestExecuteTaskNotFound(t *testing.T) {
	svc := NewQueueService(store.NewMemoryStore())

	if err := svc.ExecuteTask(context.Background(), "missing"); err != store.ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}
func TestRetryTask(t *testing.T) {
	svc := NewQueueService(store.NewMemoryStore())
	svc.SubmitTask(context.Background(), "t1", "test", "payload", 2)
	task, _ := svc.GetTask(context.Background(), "t1")
	task.Status = model.StatusFailed
	svc.store.UpdateTask(task)
	if err := svc.RetryTask(context.Background(), "t1"); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	task, _ = svc.GetTask(context.Background(), "t1")
	if task.Status != model.StatusPending {
		t.Fatalf("expected pending, got %s", task.Status)
	}
}

func TestListTasks(t *testing.T) {
	svc := NewQueueService(store.NewMemoryStore())
	svc.SubmitTask(context.Background(), "t1", "a", "p", 1)
	svc.SubmitTask(context.Background(), "t2", "b", "p", 1)
	tasks := svc.ListTasks(context.Background(), "")
	if len(tasks) != 2 {
		t.Fatalf("expected 2, got %d", len(tasks))
	}
}
