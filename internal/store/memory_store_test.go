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

func TestGetTaskStatusConcurrentWithUpdate(t *testing.T) {
	s := NewMemoryStore()
	if err := s.Add(&model.Task{ID: "t1", Status: model.StatusPending}); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	const readers = 32
	const iterations = 1000
	var wg sync.WaitGroup
	wg.Add(readers + 1)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			status := model.StatusPending
			if i%2 == 0 {
				status = model.StatusRunning
			}
			if err := s.UpdateTask(&model.Task{ID: "t1", Status: status}); err != nil {
				t.Errorf("update failed: %v", err)
				return
			}
		}
	}()

	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				status, err := s.GetTaskStatus("t1")
				if err != nil {
					t.Errorf("get status failed: %v", err)
					return
				}
				if status != model.StatusPending && status != model.StatusRunning {
					t.Errorf("unexpected status: %s", status)
					return
				}
			}
		}()
	}

	wg.Wait()
}
