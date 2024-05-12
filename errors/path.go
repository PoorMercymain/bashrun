package errors

import "errors"

var (
	ErrWrongID = errors.New("command_id should be a number and more than zero")
	ErrEmptyID = errors.New("command_id should be provided as path value")
)