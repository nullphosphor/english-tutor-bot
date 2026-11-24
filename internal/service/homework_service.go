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

	GetHomeworkByID(
		ctx context.Context,
		homeworkID int64,
	) (domain.Homework, error)

	GetHomeworksByStudentID(
		ctx context.Context,
		studentID int64,
	) ([]domain.Homework, error)

	UpdateHomework(
		ctx context.Context,
		homework domain.Homework,
	) (domain.Homework, error)

	DeleteHomework(
		ctx context.Context,
		homeworkID int64,
		version int64,
	) error
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

func (w *HomeworkService) GetHomeworkByID(
	ctx context.Context,
	homeworkID int64,
) (domain.Homework, error) {
	if homeworkID <= 0 {
		return domain.Homework{}, domain.ErrInvalidArgument
	}

	homework, err := w.homeworkRepository.GetHomeworkByID(ctx, homeworkID)
	if err != nil {
		return domain.Homework{}, fmt.Errorf("get homework from service: %w", err)
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

func (w *HomeworkService) UpdateHomework(
	ctx context.Context,
	homework domain.Homework,
) (domain.Homework, error) {
	if homework.ID <= 0 {
		return domain.Homework{}, domain.ErrInvalidArgument
	}

	if homework.Version <= 0 {
		return domain.Homework{}, domain.ErrInvalidArgument
	}

	if err := homework.Validate(); err != nil {
		return domain.Homework{}, fmt.Errorf("validate homework: %w", err)
	}

	homework, err := w.homeworkRepository.UpdateHomework(
		ctx,
		homework,
	)
	if err != nil {
		return domain.Homework{}, fmt.Errorf("update homework from service: %w", err)
	}

	return homework, nil
}

func (w *HomeworkService) DeleteHomework(
	ctx context.Context,
	homeworkID int64,
	version int64,
) error {
	if homeworkID <= 0 {
		return domain.ErrInvalidArgument
	}

	if version <= 0 {
		return domain.ErrInvalidArgument
	}

	if err := w.homeworkRepository.DeleteHomework(
		ctx,
		homeworkID,
		version,
	); err != nil {
		return fmt.Errorf("delete homework from service: %w", err)
	}

	return nil
}
