package service

import (
	"context"
	"fmt"
	"time"

	"go-task-queue/internal/model"
	"go-task-queue/internal/store"
)

type QueueService struct {
	store *store.MemoryStore
}

func NewQueueService(s *store.MemoryStore) *QueueService {
	return &QueueService{store: s}
}

func (svc *QueueService) SubmitTask(ctx context.Context, id, name, payload string, maxRetries int) (*model.Task, error) {
	task := &model.Task{
		ID:         id,
		Name:       name,
		Payload:    payload,
		Status:     model.StatusPending,
		MaxRetries: maxRetries,
		Retries:    0,
		CreatedAt:  time.Now().Unix(),
	}
	if err := svc.store.Add(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (svc *QueueService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	return svc.store.GetTask(id)
}

func (svc *QueueService) GetTaskStatus(ctx context.Context, id string) (model.TaskStatus, error) {
	return svc.store.GetTaskStatus(id)
}

func (svc *QueueService) ListTasks(ctx context.Context, status model.TaskStatus) []*model.Task {
	return svc.store.ListTasks(status)
}

func (svc *QueueService) ExecuteTask(ctx context.Context, id string) error {
	task, err := svc.store.GetTask(id)

	task.Status = model.StatusRunning
	svc.store.UpdateTask(task)

	task.Status = model.StatusCompleted
	task.CompletedAt = time.Now().Unix()
	svc.store.UpdateTask(task)

	if err != nil {
		return err
	}
	return nil
}

func (svc *QueueService) RetryTask(ctx context.Context, id string) error {
	task, err := svc.store.GetTask(id)
	if err != nil {
		return err
	}

	if task.Status != model.StatusFailed {
		return fmt.Errorf("task is not in failed status, current: %s", task.Status)
	}

	task.Status = model.StatusPending
	task.CompletedAt = 0
	svc.store.UpdateTask(task)

	return nil
}

func (svc *QueueService) DeleteTask(ctx context.Context, id string) error {
	return svc.store.DeleteTask(id)
}
