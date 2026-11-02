package domain

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidStudentID      = errors.New("invalid studentID")
	ErrInvalidLessonDateTime = errors.New("invalid dateTime")
)
