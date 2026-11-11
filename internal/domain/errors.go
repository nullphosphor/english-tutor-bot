package domain

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrDuplicateLesson = errors.New("duplicate lesson")
	ErrConflict        = errors.New("conflict")
)
