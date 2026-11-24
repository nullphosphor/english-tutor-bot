package telegram

import (
	"sync"
	"time"
)

type LessonCreationStep string

const (
	LessonStepWaitingStudentID LessonCreationStep = "waiting_student_id"
	LessonStepWaitingDate      LessonCreationStep = "waiting_date"
	LessonStepWaitingPlatform  LessonCreationStep = "waiting_platform"
	LessonStepWaitingURL       LessonCreationStep = "waiting_url"
	LessonStepWaitingConfirm   LessonCreationStep = "waiting_confirm"
	LessonStepCreating         LessonCreationStep = "creating"
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
	mu     sync.Mutex
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
		Step: LessonStepWaitingStudentID,
	}
}

func (l *LessonCreationManager) GetState(
	adminTelegramID int64,
) (LessonCreationState, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

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

	if state.Step != LessonStepWaitingStudentID {
		return ErrInvalidDialogStep
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

	if state.Step != LessonStepWaitingDate {
		return ErrInvalidDialogStep
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

	if state.Step != LessonStepWaitingPlatform {
		return ErrInvalidDialogStep
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

	if state.Step != LessonStepWaitingURL {
		return ErrInvalidDialogStep
	}

	state.URL = url
	state.Step = nextStep

	l.states[adminTelegramID] = state

	return nil
}

func (l *LessonCreationManager) LessonFinish(
	adminTelegramID int64,
) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.states, adminTelegramID)
}

func (l *LessonCreationManager) BeginCreation(
	adminTelegramID int64,
) (LessonCreationState, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[adminTelegramID]
	if !ok {
		return LessonCreationState{}, ErrDialogNotFound
	}

	if state.Step != LessonStepWaitingConfirm {
		return LessonCreationState{}, ErrInvalidDialogStep
	}

	state.Step = LessonStepCreating
	l.states[adminTelegramID] = state

	return state, nil
}

func (l *LessonCreationManager) ReturnToConfirmation(
	adminTelegramID int64,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	if state.Step != LessonStepCreating {
		return ErrInvalidDialogStep
	}

	state.Step = LessonStepWaitingConfirm
	l.states[adminTelegramID] = state

	return nil
}

func (l *LessonCreationManager) CancelDialog(
	adminTelegramID int64,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	if state.Step == LessonStepCreating {
		return ErrInvalidDialogStep
	}

	delete(l.states, adminTelegramID)
	return nil
}
