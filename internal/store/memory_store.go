package store

import (
	"sync"

	"go-task-queue/internal/model"
)

type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[string]*model.Task
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tasks: make(map[string]*model.Task),
	}
}

func (s *MemoryStore) Add(task *model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[task.ID]; exists {
		return ErrTaskAlreadyExists
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *MemoryStore) GetTask(id string) (*model.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *MemoryStore) GetTaskStatus(id string) (model.TaskStatus, error) {
	task, exists := s.tasks[id]
	if !exists {
		return "", ErrTaskNotFound
	}
	return task.Status, nil
}

func (s *MemoryStore) UpdateTask(task *model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *MemoryStore) ListTasks(status model.TaskStatus) []*model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Task, 0)
	for _, t := range s.tasks {
		if status == "" || t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

func (s *MemoryStore) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[id]; !exists {
		return ErrTaskNotFound
	}
	delete(s.tasks, id)
	return nil
}
