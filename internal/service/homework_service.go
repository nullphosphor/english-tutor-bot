package service

import (
	"context"
	"fmt"

	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type HomeworkService struct {
	homeworkRepository HomeworkRepository
}

func NewHomeworkService(
	homeworkRepository HomeworkRepository,
) *HomeworkService {
	return &HomeworkService{
		homeworkRepository: homeworkRepository,
	}
}

type HomeworkRepository interface {
	CreateHomework(
		ctx context.Context,
		homework domain.Homework,
	) (domain.Homework, error)

	GetHomeworksByStudentID(
		ctx context.Context,
		studentID int64,
	) ([]domain.Homework, error)
}

func (w *HomeworkService) CreateHomework(
	ctx context.Context,
	homework domain.Homework,
) (domain.Homework, error) {
	if err := homework.Validate(); err != nil {
		return domain.Homework{}, fmt.Errorf("validate homework: %w", err)
	}

	homework, err := w.homeworkRepository.CreateHomework(
		ctx,
		homework,
	)
	if err != nil {
		return domain.Homework{}, fmt.Errorf("create homework: %w", err)
	}

	return homework, nil
}

func (w *HomeworkService) GetHomeworksByStudentID(
	ctx context.Context,
	studentID int64,
) ([]domain.Homework, error) {
	if studentID <= 0 {
		return nil, domain.ErrInvalidArgument
	}

	homeworks, err := w.homeworkRepository.GetHomeworksByStudentID(
		ctx,
		studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("get homeworks by student id: %w", err)
	}

	return homeworks, nil
}
