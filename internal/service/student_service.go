package service

import (
	"context"
	"fmt"
	"strings"

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
	CreateStudent(
		ctx context.Context,
		student domain.Student,
	) (domain.Student, error)

	GetStudentByID(
		ctx context.Context,
		studentID int64,
	) (domain.Student, error)

	GetStudentByTelegramID(
		ctx context.Context,
		telegramID int64,
	) (domain.Student, error)

	GetStudents(
		ctx context.Context,
		isActive bool,
	) ([]domain.Student, error)

	UpdateStudentName(
		ctx context.Context,
		name string,
		studentID int64,
	) (domain.Student, error)

	SetStudentActive(
		ctx context.Context,
		isActive bool,
		studentID int64,
	) (domain.Student, error)
}

func (s *StudentService) CreateStudent(
	ctx context.Context,
	student domain.Student,
) (domain.Student, error) {
	if err := student.Validate(); err != nil {
		return domain.Student{}, fmt.Errorf("validate student: %w", err)
	}

	student, err := s.studentRepository.CreateStudent(
		ctx,
		student,
	)
	if err != nil {
		return domain.Student{}, fmt.Errorf("create student from service: %w", err)
	}

	return student, nil
}

func (s *StudentService) GetStudentByID(
	ctx context.Context,
	studentID int64,
) (domain.Student, error) {
	if studentID <= 0 {
		return domain.Student{}, domain.ErrInvalidArgument
	}

	student, err := s.studentRepository.GetStudentByID(
		ctx,
		studentID,
	)

	if err != nil {
		return domain.Student{}, fmt.Errorf("getStudentByID from service: %w", err)
	}

	return student, nil
}

func (s *StudentService) GetStudentByTelegramID(
	ctx context.Context,
	telegramID int64,
) (domain.Student, error) {
	if telegramID <= 0 {
		return domain.Student{}, domain.ErrInvalidArgument
	}

	student, err := s.studentRepository.GetStudentByTelegramID(
		ctx,
		telegramID,
	)

	if err != nil {
		return domain.Student{}, fmt.Errorf("getStudentByTelegramID from service: %w", err)
	}

	return student, nil
}

func (s *StudentService) GetStudents(
	ctx context.Context,
	isActive bool,
) ([]domain.Student, error) {
	students, err := s.studentRepository.GetStudents(
		ctx,
		isActive,
	)

	if err != nil {
		return nil, fmt.Errorf("GetStudents from service: %w", err)
	}

	return students, nil
}

func (s *StudentService) UpdateStudentName(
	ctx context.Context,
	name string,
	studentID int64,
) (domain.Student, error) {
	if studentID <= 0 {
		return domain.Student{}, domain.ErrInvalidArgument
	}

	name = strings.TrimSpace(name)
	nameLen := len([]rune(name))

	if nameLen < 1 || nameLen > 100 {
		return domain.Student{}, domain.ErrInvalidArgument
	}

	student, err := s.studentRepository.UpdateStudentName(
		ctx,
		name,
		studentID,
	)

	if err != nil {
		return domain.Student{}, fmt.Errorf("UpdateStudentName from service: %w", err)
	}

	return student, nil
}

func (s *StudentService) SetStudentActive(
	ctx context.Context,
	isActive bool,
	studentID int64,
) (domain.Student, error) {
	if studentID <= 0 {
		return domain.Student{}, domain.ErrInvalidArgument
	}

	student, err := s.studentRepository.SetStudentActive(
		ctx,
		isActive,
		studentID,
	)

	if err != nil {
		return domain.Student{}, fmt.Errorf("SetStudentActive from service: %w", err)
	}

	return student, nil
}
