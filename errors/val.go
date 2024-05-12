package errors

import "errors"

var (
	ErrWrongMIME          = errors.New("wrong MIME type provided")
	ErrSomethingWentWrong = errors.New("something went wrong, please, try again later")
	ErrWrongJSON          = errors.New("wrong JSON provided")
	ErrDuplicateInJSON = errors.New("duplicate found in provided JSON")
)