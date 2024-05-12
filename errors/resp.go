package errors

import "errors"

var (
	ErrCommandNotFound = errors.New("command with requested id not found")
)
