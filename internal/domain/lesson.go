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
	ID           int64
	StudentID    int64
	DateTime     time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
	MeetingApp   string
	MeetingURL   string
	LessonStatus LessonStatus
	CreatedAt    time.Time

	Version int64
}

func NewLesson(
	studentID int64,
	dateTime time.Time,
	meetingApp, meetingURL string,
) Lesson {
	return Lesson{
		StudentID:    studentID,
		DateTime:     dateTime,
		MeetingApp:   meetingApp,
		MeetingURL:   meetingURL,
		LessonStatus: LessonStatusScheduled,
	}
}

func (l *Lesson) Validate() error {
	if l.StudentID <= 0 {
		return ErrInvalidArgument
	}

	if l.DateTime.IsZero() {
		return ErrInvalidArgument
	}

	if l.MeetingApp != "Zoom" && l.MeetingApp != "Google Meet" {
		return ErrInvalidArgument
	}

	if strings.TrimSpace(l.MeetingURL) == "" {
		return ErrInvalidArgument
	}

	if l.LessonStatus != LessonStatusScheduled &&
		l.LessonStatus != LessonStatusInProgress &&
		l.LessonStatus != LessonStatusCompleted &&
		l.LessonStatus != LessonStatusCancelled {
		return ErrInvalidArgument
	}

	return nil
}

func (l *Lesson) StartLesson() error {
	if l.LessonStatus != LessonStatusScheduled {
		return ErrInvalidArgument
	}

	now := time.Now()
	l.StartedAt = &now
	l.LessonStatus = LessonStatusInProgress

	return nil
}

func (l *Lesson) FinishLesson() error {
	if l.LessonStatus != LessonStatusInProgress {
		return ErrInvalidArgument
	}

	now := time.Now()
	l.FinishedAt = &now
	l.LessonStatus = LessonStatusCompleted

	return nil
}

func (l *Lesson) CancelLesson() error {
	switch l.LessonStatus {
	case LessonStatusScheduled:
		l.LessonStatus = LessonStatusCancelled
	case LessonStatusInProgress:
		now := time.Now()
		l.FinishedAt = &now
		l.LessonStatus = LessonStatusCancelled
	case LessonStatusCancelled:
		return ErrInvalidArgument
	default:
		return ErrInvalidArgument
	}

	return nil
}
