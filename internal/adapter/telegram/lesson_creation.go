package telegram

import (
	"sync"
	"time"
)

type LessonCreationStep string

const (
	StepWaitingStudentID LessonCreationStep = "waiting_student_id"
	StepWaitingDate      LessonCreationStep = "waiting_date"
	StepWaitingPlatform  LessonCreationStep = "waiting_platform"
	StepWaitingURL       LessonCreationStep = "waiting_url"
	StepWaitingConfirm   LessonCreationStep = "waiting_confirm"
)

type LessonCreationState struct {
	Step      LessonCreationStep
	StudentID int64
	Date      time.Time
	Platform  string
	URL       string
}

type LessonCreationManager struct {
	states map[int64]LessonCreationState
	mu     sync.RWMutex
}

func NewLessonCreationManager() *LessonCreationManager {
	return &LessonCreationManager{
		states: make(map[int64]LessonCreationState),
	}
}

func (l *LessonCreationManager) Start(
	adminTelegramID int64,
) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.states[adminTelegramID] = LessonCreationState{
		Step: StepWaitingStudentID,
	}
}

func (l *LessonCreationManager) GetState(
	adminTelegramID int64,
) (LessonCreationState, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, ok := l.states[adminTelegramID]
	if !ok {
		return LessonCreationState{}, ErrDialogNotFound
	}

	return val, nil
}

func (l *LessonCreationManager) UpdateState(
	adminTelegramID int64,
	studentID int64,
	nextStep LessonCreationStep,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	state.StudentID = studentID
	state.Step = nextStep

	l.states[adminTelegramID] = state

	return nil
}

func (l *LessonCreationManager) UpdateDate(
	adminTelegramID int64,
	date time.Time,
	nextStep LessonCreationStep,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	state.Date = date
	state.Step = nextStep

	l.states[adminTelegramID] = state

	return nil
}

func (l *LessonCreationManager) UpdatePlatform(
	adminTelegramID int64,
	platform string,
	nextStep LessonCreationStep,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	state.Platform = platform
	state.Step = nextStep

	l.states[adminTelegramID] = state

	return nil
}

func (l *LessonCreationManager) UpdateURL(
	adminTelegramID int64,
	url string,
	nextStep LessonCreationStep,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	state.URL = url
	state.Step = nextStep

	l.states[adminTelegramID] = state

	return nil
}

func (l *LessonCreationManager) Finish(
	adminTelegramID int64,
) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.states, adminTelegramID)
}
