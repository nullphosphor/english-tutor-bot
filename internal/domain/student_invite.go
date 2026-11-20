package domain

import (
	"strings"
	"time"
)

type StudentInviteStatus string

const (
	StudentInviteStatusIssued   StudentInviteStatus = "issued"
	StudentInviteStatusPending  StudentInviteStatus = "pending"
	StudentInviteStatusApproved StudentInviteStatus = "approved"
	StudentInviteStatusRejected StudentInviteStatus = "rejected"
)

type StudentInvite struct {
	ID                  int64
	TokenHash           string
	StudentName         string
	RequesterTelegramID *int64
	Status              StudentInviteStatus
	ExpiresAt           time.Time
	RequestedAt         *time.Time
	DecidedAt           *time.Time
	CreatedAt           time.Time
}

func NewStudentInvite(studentName, tokenHash string, expitesAt time.Time) StudentInvite {
	return StudentInvite{
		TokenHash:   tokenHash,
		StudentName: studentName,
		Status:      StudentInviteStatusIssued,
		ExpiresAt:   expitesAt,
	}
}

func (si *StudentInvite) Validate() error {
	si.StudentName = strings.TrimSpace(si.StudentName)
	studentNameLen := len([]rune(si.StudentName))
	if studentNameLen < 1 || studentNameLen > 100 {
		return ErrInvalidArgument
	}

	return nil
}
