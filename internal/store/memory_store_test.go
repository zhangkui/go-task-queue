package store

import (
	"sync"
	"testing"

	"go-task-queue/internal/model"
)

func TestAddAndGetTask(t *testing.T) {
	s := NewMemoryStore()
	task := &model.Task{ID: "t1", Name: "test", Status: model.StatusPending}
	if err := s.Add(task); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	got, err := s.GetTask("t1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Name != "test" {
		t.Fatalf("expected test, got %s", got.Name)
	}
}

func TestAddDuplicate(t *testing.T) {
	s := NewMemoryStore()
	s.Add(&model.Task{ID: "t1", Name: "test"})
	if err := s.Add(&model.Task{ID: "t1", Name: "dup"}); err != ErrTaskAlreadyExists {
		t.Fatalf("expected ErrTaskAlreadyExists, got %v", err)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetTask("missing")
	if err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestUpdateTask(t *testing.T) {
	s := NewMemoryStore()
	s.Add(&model.Task{ID: "t1", Name: "test", Status: model.StatusPending})
	task, _ := s.GetTask("t1")
	task.Status = model.StatusCompleted
	s.UpdateTask(task)
	got, _ := s.GetTask("t1")
	if got.Status != model.StatusCompleted {
		t.Fatalf("expected completed, got %s", got.Status)
	}
}

func TestListTasks(t *testing.T) {
	s := NewMemoryStore()
	s.Add(&model.Task{ID: "t1", Status: model.StatusPending})
	s.Add(&model.Task{ID: "t2", Status: model.StatusCompleted})
	s.Add(&model.Task{ID: "t3", Status: model.StatusPending})
	all := s.ListTasks("")
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	pending := s.ListTasks(model.StatusPending)
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending, got %d", len(pending))
	}
}

func TestDeleteTask(t *testing.T) {
	s := NewMemoryStore()
	s.Add(&model.Task{ID: "t1", Name: "test"})
	if err := s.DeleteTask("t1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := s.GetTask("t1"); err != ErrTaskNotFound {
		t.Fatalf("expected not found after delete")
	}
}

// TestGetTaskStatusConcurrent reproduces the "concurrent map read and map write"
// crash: many goroutines query task status while writers mutate the store.
// The read path (/task/status) must hold the store's read lock; otherwise a
// concurrent writer (submit/execute/delete) triggers a runtime fatal error.
// Run with -race to also surface any remaining data race.
func TestGetTaskStatusConcurrent(t *testing.T) {
	s := NewMemoryStore()
	s.Add(&model.Task{ID: "t1", Status: model.StatusPending})

	var writers sync.WaitGroup
	var readers sync.WaitGroup
	stop := make(chan struct{})

	// Readers: concurrent status queries (the /task/status path). They keep
	// hammering the store until the writers finish, then drain.
	readers.Add(64)
	for i := 0; i < 64; i++ {
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
					if _, err := s.GetTaskStatus("t1"); err != nil && err != ErrTaskNotFound {
						t.Errorf("unexpected error: %v", err)
						return
					}
				}
			}
		}()
	}

	// Writers: concurrent mutations of the store (add/update/delete churn).
	writers.Add(16)
	for i := 0; i < 16; i++ {
		go func() {
			defer writers.Done()
			for j := 0; j < 5000; j++ {
				s.UpdateTask(&model.Task{ID: "t1", Status: model.StatusRunning})
				// Force occasional map growth/shrink churn to widen the race window.
				if j%500 == 0 {
					s.Add(&model.Task{ID: "t1-x", Status: model.StatusPending})
					s.DeleteTask("t1-x")
				}
			}
		}()
	}

	// Once writers finish, release the readers and wait for them to drain.
	writers.Wait()
	close(stop)
	readers.Wait()
}
