package errors

import "errors"

var (
	ErrWrongLimit = errors.New("limit should be a number in range [1:50]")
	ErrWrongOffset = errors.New("offset should be a non-negative number")
)