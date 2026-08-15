package store

import "errors"

var (
	ErrTaskNotFound  = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
)
