package service

import (
	"context"
	"fmt"

	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type StudentService struct {
	studentRepository StudentRepository
}

func NewStudentService(
	studentRepository StudentRepository,
) *StudentService {
	return &StudentService{
		studentRepository: studentRepository,
	}
}

type StudentRepository interface {
	GetStudentByTelegramID(
		ctx context.Context,
		telegramID int64,
	) (domain.Student, error)
}

func (s *StudentService) GetStudentByTelegramID(
	ctx context.Context,
	telegramID int64,
) (domain.Student, error) {
	student, err := s.studentRepository.GetStudentByTelegramID(
		ctx,
		telegramID,
	)

	if err != nil {
		return domain.Student{}, fmt.Errorf("get studentByTelegramID: %w", err)
	}

	return student, nil
}
