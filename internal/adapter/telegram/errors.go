package telegram

import "errors"

var (
	ErrInvalidScriptStep = errors.New("invalid script step")
	ErrDialogNotFound    = errors.New("dialog not found")
)
