package model

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID         string
	Name       string
	Payload    string
	Status     TaskStatus
	MaxRetries int
	Retries    int
	CreatedAt  int64
	CompletedAt int64
}
