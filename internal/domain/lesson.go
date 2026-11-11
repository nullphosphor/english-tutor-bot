package domain

import (
	"errors"
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
		return errors.New("invalid studentID")
	}

	if l.DateTime.IsZero() {
		return errors.New("invalid dateTime")
	}

	if l.MeetingApp != "Zoom" && l.MeetingApp != "Google Meet" {
		return errors.New("invalid meetingApp")
	}

	if strings.TrimSpace(l.MeetingURL) == "" {
		return errors.New("invalid meeting url")
	}

	if l.LessonStatus != LessonStatusScheduled &&
		l.LessonStatus != LessonStatusInProgress &&
		l.LessonStatus != LessonStatusCompleted &&
		l.LessonStatus != LessonStatusCancelled {
		return errors.New("invalid lessonStatus")
	}

	return nil
}

func (l *Lesson) StartLesson() error {
	if l.LessonStatus != LessonStatusScheduled {
		return errors.New("statusLesson must be 'scheduled'")
	}

	now := time.Now()
	l.StartedAt = &now
	l.LessonStatus = LessonStatusInProgress

	return nil
}

func (l *Lesson) FinishLesson() error {
	if l.LessonStatus != LessonStatusInProgress {
		return errors.New("statusLesson must be 'in_progress'")
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
		return errors.New("lesson already cancelled")
	default:
		return ErrInvalidArgument
	}

	return nil
}
