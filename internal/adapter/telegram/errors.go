package telegram

import "errors"

var (
	ErrInvalidDialogStep = errors.New("invalid dialog step")
	ErrDialogNotFound    = errors.New("dialog not found")
	ErrNotEditing        = errors.New("not editing")
)
