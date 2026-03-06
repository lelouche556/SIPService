package util

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrValidation          = errors.New("validation error")
	ErrInvalidState        = errors.New("invalid state")
	ErrConflict            = errors.New("conflict")
	ErrExecutionInProgress = errors.New("execution in progress")
)
