package errors

import "errors"

var (
	ErrRowsNotAffected = errors.New("no rows were affected by query")
	ErrNoRows = errors.New("no rows found in DB for this limit and offset")
)