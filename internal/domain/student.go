package domain

import "strings"

type Student struct {
	ID         int64
	TelegramID int64
	Name       string
	IsActive   bool
}

func (s *Student) Validate() error {
	if s.TelegramID <= 0 {
		return ErrInvalidArgument
	}

	s.Name = strings.TrimSpace(s.Name)
	nameLen := len([]rune(s.Name))

	if nameLen < 1 || nameLen > 100 {
		return ErrInvalidArgument
	}

	return nil
}
