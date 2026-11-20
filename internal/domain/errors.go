package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrDuplicateLesson   = errors.New("duplicate lesson")
	ErrConflict          = errors.New("conflict")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInviteUnavailable = errors.New("invite unavailable")
)
