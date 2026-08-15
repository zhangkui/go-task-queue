package store

import (
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

func TestListTasksReturnsIndependentCopies(t *testing.T) {
	s := NewMemoryStore()
	s.Add(&model.Task{ID: "t1", Status: model.StatusPending})
	s.Add(&model.Task{ID: "t2", Status: model.StatusPending})

	tasks := s.ListTasks("")
	if len(tasks) != 2 {
		t.Fatalf("expected 2, got %d", len(tasks))
	}

	// Mutating the returned tasks must not affect server-side state.
	tasks[0].Status = model.StatusCompleted
	tasks[1].Status = model.StatusFailed

	for _, id := range []string{"t1", "t2"} {
		status, err := s.GetTaskStatus(id)
		if err != nil {
			t.Fatalf("get status for %s failed: %v", id, err)
		}
		if status != model.StatusPending {
			t.Fatalf("expected server task %s to remain pending, got %s", id, status)
		}
	}
}

func TestGetTaskReturnsIndependentCopy(t *testing.T) {
	s := NewMemoryStore()
	s.Add(&model.Task{ID: "t1", Name: "test", Status: model.StatusPending})

	got, _ := s.GetTask("t1")
	got.Status = model.StatusCompleted // mutate the returned copy

	status, _ := s.GetTaskStatus("t1")
	if status != model.StatusPending {
		t.Fatalf("expected server task to remain pending, got %s", status)
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
