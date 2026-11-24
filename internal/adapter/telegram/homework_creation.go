package telegram

import (
	"sync"
	"time"

	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type HomeworkCreationStep string

const (
	HomeworkStepWaitingStudentID   HomeworkCreationStep = "waiting_student_id"
	HomeworkStepWaitingTitle       HomeworkCreationStep = "waiting_title"
	HomeworkStepWaitingDescription HomeworkCreationStep = "waiting_description"
	HomeworkStepWaitingDue         HomeworkCreationStep = "waiting_due"
	HomeworkStepWaitingConfirm     HomeworkCreationStep = "waiting_confirm"
	HomeworkStepCreating           HomeworkCreationStep = "creating"
	HomeworkStepEditingMenu        HomeworkCreationStep = "editing_menu"
	HomeworkStepEditingConfirm     HomeworkCreationStep = "editing_confirm"
)

type HomeworkCreationState struct {
	Step        HomeworkCreationStep
	StudentID   int64
	Title       string
	Description string
	DueAt       *time.Time

	HomeworkID int64
	Version    int64
	IsEditing  bool
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

	if state.Step != HomeworkStepWaitingStudentID {
		return ErrInvalidDialogStep
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

	if state.Step != HomeworkStepWaitingTitle {
		return ErrInvalidDialogStep
	}

	if state.IsEditing {
		state.Step = HomeworkStepEditingMenu
	} else {
		state.Step = HomeworkStepWaitingDescription
	}

	state.Title = title

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

	if state.Step != HomeworkStepWaitingDescription {
		return ErrInvalidDialogStep
	}

	if state.IsEditing {
		state.Step = HomeworkStepEditingMenu
	} else {
		state.Step = HomeworkStepWaitingDue
	}

	state.Description = description

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

	if state.Step != HomeworkStepWaitingDue {
		return ErrInvalidDialogStep
	}

	if state.IsEditing {
		state.Step = HomeworkStepEditingMenu
	} else {
		state.Step = HomeworkStepWaitingConfirm
	}

	state.DueAt = dueAt

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

	if state.IsEditing || state.Step != HomeworkStepWaitingConfirm {
		return HomeworkCreationState{}, ErrInvalidDialogStep
	}

	state.Step = HomeworkStepCreating
	w.states[adminTelegramID] = state

	return state, nil
}

func (w *HomeworkCreationManager) BeginUpdate(
	adminTelegramID int64,
) (HomeworkCreationState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return HomeworkCreationState{}, ErrDialogNotFound
	}

	if !state.IsEditing || state.Step != HomeworkStepEditingConfirm {
		return HomeworkCreationState{}, ErrInvalidDialogStep
	}

	state.Step = HomeworkStepCreating
	w.states[adminTelegramID] = state

	return state, nil
}

func (w *HomeworkCreationManager) ReturnToEditConfirmation(
	adminTelegramID int64,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	if !state.IsEditing || state.Step != HomeworkStepCreating {
		return ErrInvalidDialogStep
	}

	state.Step = HomeworkStepEditingConfirm
	w.states[adminTelegramID] = state

	return nil
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

	if state.Step == HomeworkStepCreating {
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

func (w *HomeworkCreationManager) StartEdit(
	adminTelegramID int64,
	homework domain.Homework,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.states[adminTelegramID] = HomeworkCreationState{
		HomeworkID:  homework.ID,
		StudentID:   homework.StudentID,
		Version:     homework.Version,
		Title:       homework.Title,
		Description: homework.Description,
		DueAt:       homework.DueAt,
		IsEditing:   true,
		Step:        HomeworkStepEditingMenu,
	}
}

func (w *HomeworkCreationManager) StartEditingField(
	adminTelegramID int64,
	targetStep HomeworkCreationStep,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	if !state.IsEditing {
		return ErrNotEditing
	}

	if state.Step != HomeworkStepEditingMenu {
		return ErrInvalidDialogStep
	}

	switch targetStep {
	case HomeworkStepWaitingTitle, HomeworkStepWaitingDescription, HomeworkStepWaitingDue:
		state.Step = targetStep
	default:
		return ErrInvalidDialogStep
	}

	w.states[adminTelegramID] = state

	return nil
}

func (w *HomeworkCreationManager) BeginEditConfirmation(
	adminTelegramID int64,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.states[adminTelegramID]
	if !ok {
		return ErrDialogNotFound
	}

	if !state.IsEditing {
		return ErrNotEditing
	}

	if state.Step != HomeworkStepEditingMenu {
		return ErrInvalidDialogStep
	}

	state.Step = HomeworkStepEditingConfirm
	w.states[adminTelegramID] = state

	return nil
}
