package telegram

import (
	"sync"
	"time"
)

type HomeworkCreationStep string

const (
	HomeworkStepWaitingStudentID   HomeworkCreationStep = "waiting_student_id"
	HomeworkStepWaitingTitle       HomeworkCreationStep = "waiting_title"
	HomeworkStepWaitingDescription HomeworkCreationStep = "waiting_description"
	HomeworkStepWaitingDue         HomeworkCreationStep = "waiting_due"
	HomeworkStepWaitingConfirm     HomeworkCreationStep = "waiting_confirm"
	HomeworkStepCreating           HomeworkCreationStep = "creating"
)

type HomeworkCreationState struct {
	Step        HomeworkCreationStep
	StudentID   int64
	Title       string
	Description string
	DueAt       *time.Time
}

type HomeworkCreationManager struct {
	states map[int64]HomeworkCreationState
	mu     sync.Mutex
}

func NewHomeworkCreationManager() *HomeworkCreationManager {
	return &HomeworkCreationManager{
		states: make(map[int64]HomeworkCreationState),
	}
}

func (w *HomeworkCreationManager) Start(
	adminTelegramID int64,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.states[adminTelegramID] = HomeworkCreationState{
		Step: HomeworkStepWaitingStudentID,
	}
}

func (w *HomeworkCreationManager) GetState(
	adminTelegramID int64,
) (HomeworkCreationState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	val, ok := w.states[adminTelegramID]
	if !ok {
		return HomeworkCreationState{}, ErrDialogNotFound
	}

	return val, nil
}

func (w *HomeworkCreationManager) UpdateStudentID(
	adminTelegramID, studentID int64,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	state.StudentID = studentID
	state.Step = HomeworkStepWaitingTitle

	w.states[adminTelegramID] = state
	return nil
}

func (w *HomeworkCreationManager) UpdateTitle(
	adminTelegramID int64,
	title string,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	state.Title = title
	state.Step = HomeworkStepWaitingDescription

	w.states[adminTelegramID] = state
	return nil
}

func (w *HomeworkCreationManager) UpdateDescription(
	adminTelegramID int64,
	description string,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	state.Description = description
	state.Step = HomeworkStepWaitingDue

	w.states[adminTelegramID] = state
	return nil
}

func (w *HomeworkCreationManager) UpdateDueAt(
	adminTelegramID int64,
	dueAt *time.Time,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	state.DueAt = dueAt
	state.Step = HomeworkStepWaitingConfirm

	w.states[adminTelegramID] = state
	return nil
}

func (w *HomeworkCreationManager) BeginCreation(
	adminTelegramID int64,
) (HomeworkCreationState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return HomeworkCreationState{}, ErrDialogNotFound
	}

	if state.Step != HomeworkStepWaitingConfirm {
		return HomeworkCreationState{}, ErrInvalidDialogStep
	}

	state.Step = HomeworkStepCreating
	w.states[adminTelegramID] = state

	return state, nil
}

func (w *HomeworkCreationManager) ReturnToConfirmation(
	adminTelegramID int64,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	if state.Step != HomeworkStepCreating {
		return ErrInvalidDialogStep
	}

	state.Step = HomeworkStepWaitingConfirm
	w.states[adminTelegramID] = state

	return nil
}

func (w *HomeworkCreationManager) CancelDialog(
	adminTelegramID int64,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	if state.Step != HomeworkStepWaitingConfirm {
		return ErrInvalidDialogStep
	}

	delete(w.states, adminTelegramID)

	return nil
}

func (w *HomeworkCreationManager) FinishCreation(
	adminTelegramID int64,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.states, adminTelegramID)
}
