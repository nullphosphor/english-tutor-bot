package domain

import (
	"errors"
	"strings"
	"time"
)

type LessonStatus string

const (
	LessonStatusScheduled LessonStatus = "scheduled"
	LessonStatusCompleted LessonStatus = "completed"
	LessonStatusCancelled LessonStatus = "cancelled"
)

type Lesson struct {
	ID           int64
	StudentID    int64
	DateTime     time.Time
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

	if l.LessonStatus != LessonStatusScheduled && l.LessonStatus != LessonStatusCompleted && l.LessonStatus != LessonStatusCancelled {
		return errors.New("invalid lessonStatus")
	}

	return nil
}
