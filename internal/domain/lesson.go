package domain

import (
	"strings"
	"time"
)

type LessonStatus string

const (
	LessonStatusScheduled  LessonStatus = "scheduled"
	LessonStatusInProgress LessonStatus = "in_progress"
	LessonStatusCompleted  LessonStatus = "completed"
	LessonStatusCancelled  LessonStatus = "cancelled"
)

type Lesson struct {
	ID         int64
	StudentID  int64
	DateTime   time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
	Platform   string
	URL        string
	Status     LessonStatus
	CreatedAt  time.Time

	Version int64
}

func NewLesson(
	studentID int64,
	dateTime time.Time,
	meetingApp, meetingURL string,
) Lesson {
	return Lesson{
		StudentID: studentID,
		DateTime:  dateTime,
		Platform:  meetingApp,
		URL:       meetingURL,
		Status:    LessonStatusScheduled,
	}
}

func (l *Lesson) Validate() error {
	if l.StudentID <= 0 {
		return ErrInvalidArgument
	}

	if l.DateTime.IsZero() {
		return ErrInvalidArgument
	}

	if l.Platform != "Zoom" &&
		l.Platform != "Google Meet" &&
		l.Platform != "Yandex Telemost" {
		return ErrInvalidArgument
	}

	if strings.TrimSpace(l.URL) == "" {
		return ErrInvalidArgument
	}

	if l.Status != LessonStatusScheduled &&
		l.Status != LessonStatusInProgress &&
		l.Status != LessonStatusCompleted &&
		l.Status != LessonStatusCancelled {
		return ErrInvalidArgument
	}

	return nil
}

func (l *Lesson) StartLesson() error {
	if l.Status != LessonStatusScheduled {
		return ErrInvalidArgument
	}

	now := time.Now()
	l.StartedAt = &now
	l.Status = LessonStatusInProgress

	return nil
}

func (l *Lesson) FinishLesson() error {
	if l.Status != LessonStatusInProgress {
		return ErrInvalidArgument
	}

	now := time.Now()
	l.FinishedAt = &now
	l.Status = LessonStatusCompleted

	return nil
}

func (l *Lesson) CancelLesson() error {
	switch l.Status {
	case LessonStatusScheduled:
		l.Status = LessonStatusCancelled
	case LessonStatusInProgress:
		now := time.Now()
		l.FinishedAt = &now
		l.Status = LessonStatusCancelled
	case LessonStatusCancelled:
		return ErrInvalidArgument
	default:
		return ErrInvalidArgument
	}

	return nil
}
