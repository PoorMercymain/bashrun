package errors

import "errors"

var (
	ErrEmptyCommand = errors.New("empty command provided")
)
