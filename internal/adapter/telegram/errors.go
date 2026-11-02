package telegram

import "errors"

var (
	ErrLessonNotFound               = errors.New("lesson not found")
	ErrInvalidScriptStep            = errors.New("invalid script step")
	ErrDialogNotFound               = errors.New("dialog not found")
	ErrLessonCreationAlreadyStarted = errors.New("lesson creation already started")
)
