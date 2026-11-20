package domain

import (
	"strings"
	"time"
)

type Homework struct {
	ID          int64
	StudentID   int64
	Title       string
	Description string
	DueAt       *time.Time
	CreatedAt   time.Time
}

func NewHomework(
	studentID int64,
	title, description string,
	dueAt *time.Time,
) Homework {
	return Homework{
		StudentID:   studentID,
		Title:       title,
		Description: description,
		DueAt:       dueAt,
	}
}

func (w *Homework) Validate() error {
	if w.StudentID <= 0 {
		return ErrInvalidArgument
	}

	w.Title = strings.TrimSpace(w.Title)
	w.Description = strings.TrimSpace(w.Description)

	titleLen := len([]rune(w.Title))
	if titleLen < 3 || titleLen > 30 {
		return ErrInvalidArgument
	}

	descriptionLen := len([]rune(w.Description))
	if descriptionLen < 3 || descriptionLen > 3500 {
		return ErrInvalidArgument
	}

	return nil
}
