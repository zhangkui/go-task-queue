package service

import (
	"context"
	"errors"
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
	if task.Retries != 1 {
		t.Fatalf("expected 1 retry, got %d", task.Retries)
	}
}

func TestRetryTaskStopsAtMaxRetries(t *testing.T) {
	svc := NewQueueService(store.NewMemoryStore())
	svc.SubmitTask(context.Background(), "t1", "test", "payload", 2)

	for retry := 1; retry <= 2; retry++ {
		task, _ := svc.GetTask(context.Background(), "t1")
		task.Status = model.StatusFailed
		if err := svc.store.UpdateTask(task); err != nil {
			t.Fatalf("mark task failed: %v", err)
		}

		if err := svc.RetryTask(context.Background(), "t1"); err != nil {
			t.Fatalf("retry %d failed: %v", retry, err)
		}
	}

	task, _ := svc.GetTask(context.Background(), "t1")
	task.Status = model.StatusFailed
	if err := svc.store.UpdateTask(task); err != nil {
		t.Fatalf("mark task failed: %v", err)
	}

	err := svc.RetryTask(context.Background(), "t1")
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Fatalf("expected ErrMaxRetriesExceeded, got %v", err)
	}

	task, _ = svc.GetTask(context.Background(), "t1")
	if task.Retries != 2 {
		t.Fatalf("expected retries to remain 2, got %d", task.Retries)
	}
	if task.Status != model.StatusFailed {
		t.Fatalf("expected failed status, got %s", task.Status)
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
