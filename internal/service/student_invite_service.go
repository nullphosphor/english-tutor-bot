package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type StudentInviteService struct {
	studentInviteRepository StudentInviteRepository
}

func NewStudentInviteService(
	studentInviteRepository StudentInviteRepository,
) *StudentInviteService {
	return &StudentInviteService{
		studentInviteRepository: studentInviteRepository,
	}
}

type StudentInviteRepository interface {
	CreateInvite(
		ctx context.Context,
		studentInvite domain.StudentInvite,
	) (domain.StudentInvite, error)

	RequestRegistration(
		ctx context.Context,
		tokenHash string,
		telegramID int64,
	) (domain.StudentInvite, error)

	GetPendingInvites(
		ctx context.Context,
	) ([]domain.StudentInvite, error)

	ApproveInvite(
		ctx context.Context,
		inviteID int64,
	) (domain.Student, error)

	RejectInvite(
		ctx context.Context,
		inviteID int64,
	) (domain.StudentInvite, error)
}

func (si *StudentInviteService) CreateInvite(
	ctx context.Context,
	studentInvite domain.StudentInvite,
) (domain.StudentInvite, string, error) {
	if err := studentInvite.Validate(); err != nil {
		return domain.StudentInvite{}, "", fmt.Errorf("validate invite: %w", err)
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return domain.StudentInvite{}, "", fmt.Errorf("read 32 bytes: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(b)

	bytes := sha256.Sum256([]byte(token))

	studentInvite.TokenHash = hex.EncodeToString(bytes[:])
	studentInvite.Status = domain.StudentInviteStatusIssued
	studentInvite.ExpiresAt = time.Now().Add(24 * time.Hour)

	studentInviteService := domain.NewStudentInvite(
		studentInvite.StudentName,
		studentInvite.TokenHash,
		studentInvite.ExpiresAt,
	)

	studentInvite, err := si.studentInviteRepository.CreateInvite(
		ctx,
		studentInviteService,
	)

	if err != nil {
		return domain.StudentInvite{}, "", fmt.Errorf("create invite service: %w", err)
	}

	return studentInvite, token, nil
}

func (si *StudentInviteService) RequestRegistration(
	ctx context.Context,
	token string,
	telegramID int64,
) (domain.StudentInvite, error) {
	if telegramID <= 0 {
		return domain.StudentInvite{}, domain.ErrInvalidArgument
	}

	if len(token) == 0 {
		return domain.StudentInvite{}, domain.ErrInvalidArgument
	}

	bytes := sha256.Sum256([]byte(token))

	token = hex.EncodeToString(bytes[:])

	studentInvite, err := si.studentInviteRepository.RequestRegistration(
		ctx,
		token,
		telegramID,
	)

	if err != nil {
		return domain.StudentInvite{}, fmt.Errorf("request registration service: %w", err)
	}

	return studentInvite, nil
}

func (si *StudentInviteService) GetPendingInvites(
	ctx context.Context,
) ([]domain.StudentInvite, error) {
	studentInvites, err := si.studentInviteRepository.GetPendingInvites(ctx)

	if err != nil {
		return nil, fmt.Errorf("get pending invites service: %w", err)
	}

	return studentInvites, nil
}

func (si *StudentInviteService) ApproveInvite(
	ctx context.Context,
	inviteID int64,
) (domain.Student, error) {
	if inviteID <= 0 {
		return domain.Student{}, domain.ErrInvalidArgument
	}

	student, err := si.studentInviteRepository.ApproveInvite(
		ctx,
		inviteID,
	)

	if err != nil {
		return domain.Student{}, fmt.Errorf("approve invite service: %w", err)
	}

	return student, nil
}

func (si *StudentInviteService) RejectInvite(
	ctx context.Context,
	inviteID int64,
) (domain.StudentInvite, error) {
	if inviteID <= 0 {
		return domain.StudentInvite{}, domain.ErrInvalidArgument
	}

	studentInvite, err := si.studentInviteRepository.RejectInvite(
		ctx,
		inviteID,
	)

	if err != nil {
		return domain.StudentInvite{}, fmt.Errorf("reject invite service: %w", err)
	}

	return studentInvite, nil
}
