package errors

import "errors"

var (
	ErrCommandNotRunning = errors.New("the command is not running already")
	ErrCommandStopped = errors.New("the command is stopped")
)