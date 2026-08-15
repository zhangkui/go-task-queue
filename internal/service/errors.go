package service

import "errors"

var ErrMaxRetriesExceeded = errors.New("maximum retry count reached")
