package telegram

import "sync"

type StudentInviteCreationManager struct {
	states map[int64]struct{}
	mu     sync.Mutex
}

func NewStudentInviteCreationManager() *StudentInviteCreationManager {
	return &StudentInviteCreationManager{
		states: make(map[int64]struct{}),
	}
}

func (si *StudentInviteCreationManager) Start(
	adminTelegramID int64,
) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.states[adminTelegramID] = struct{}{}
}

func (si *StudentInviteCreationManager) IsActive(
	adminTelegramID int64,
) bool {
	si.mu.Lock()
	defer si.mu.Unlock()

	_, ok := si.states[adminTelegramID]
	return ok
}

func (si *StudentInviteCreationManager) Finish(
	adminTelegramID int64,
) {
	si.mu.Lock()
	defer si.mu.Unlock()

	delete(si.states, adminTelegramID)
}
