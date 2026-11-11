package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type LessonService struct {
	lessonRepository LessonRepository
}

func NewLessonService(
	lessonRepository LessonRepository,
) *LessonService {
	return &LessonService{
		lessonRepository: lessonRepository,
	}
}

type LessonRepository interface {
	CreateLesson(
		ctx context.Context,
		lesson domain.Lesson,
	) (domain.Lesson, error)

	GetLessonByID(
		ctx context.Context,
		id int64,
	) (domain.Lesson, error)

	GetLessonsByPeriod(
		ctx context.Context,
		from, to time.Time,
	) ([]domain.Lesson, error)

	UpdateLesson(
		ctx context.Context,
		lesson domain.Lesson,
	) (domain.Lesson, error)

	ExistsActiveLesson(
		ctx context.Context,
		studentID int64,
		dateTime time.Time,
	) (bool, error)
}

func (s *LessonService) CreateLesson(
	ctx context.Context,
	lesson domain.Lesson,
) (domain.Lesson, error) {
	if err := lesson.Validate(); err != nil {
		return domain.Lesson{}, fmt.Errorf("validate lesson: %w", err)
	}

	exists, err := s.lessonRepository.ExistsActiveLesson(
		ctx,
		lesson.StudentID,
		lesson.DateTime,
	)

	if err != nil {
		return domain.Lesson{}, fmt.Errorf("exists lesson: %w", err)
	}

	if exists {
		return domain.Lesson{}, domain.ErrDuplicateLesson
	}

	lessonService, err := s.lessonRepository.CreateLesson(ctx, lesson)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("create lesson: %w", err)
	}

	return lessonService, nil
}

func (s *LessonService) GetLessonByID(
	ctx context.Context,
	id int64,
) (domain.Lesson, error) {
	lesson, err := s.lessonRepository.GetLessonByID(ctx, id)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("get lesson: %w", err)
	}

	return lesson, nil
}

func (s *LessonService) GetLessonsByPeriod(
	ctx context.Context,
	from, to time.Time,
) ([]domain.Lesson, error) {
	if from.After(to) {
		return []domain.Lesson{}, fmt.Errorf("the beginning of the period must not be later than its end: %w", domain.ErrInvalidArgument)
	}

	lessons, err := s.lessonRepository.GetLessonsByPeriod(
		ctx,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("get lessons: %w", err)
	}

	return lessons, nil
}

func (s *LessonService) StartLesson(
	ctx context.Context,
	id int64,
) (domain.Lesson, error) {
	lesson, err := s.lessonRepository.GetLessonByID(ctx, id)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("get lesson: %w", err)
	}

	if err := lesson.StartLesson(); err != nil {
		return domain.Lesson{}, fmt.Errorf("start lesson: %w", err)
	}

	lessonService, err := s.lessonRepository.UpdateLesson(ctx, lesson)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("update lesson: %w", err)
	}
	return lessonService, nil
}

func (s *LessonService) FinishLesson(
	ctx context.Context,
	id int64,
) (domain.Lesson, error) {
	lesson, err := s.lessonRepository.GetLessonByID(ctx, id)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("get lesson: %w", err)
	}

	if err := lesson.FinishLesson(); err != nil {
		return domain.Lesson{}, fmt.Errorf("finish lesson: %w", err)
	}

	lessonService, err := s.lessonRepository.UpdateLesson(ctx, lesson)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("update lesson: %w", err)
	}

	return lessonService, nil
}

func (s *LessonService) CancelLesson(
	ctx context.Context,
	id int64,
) (domain.Lesson, error) {
	lesson, err := s.lessonRepository.GetLessonByID(ctx, id)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("get lesson: %w", err)
	}

	if err := lesson.CancelLesson(); err != nil {
		return domain.Lesson{}, fmt.Errorf("cancel lesson: %w", err)
	}

	lessonService, err := s.lessonRepository.UpdateLesson(ctx, lesson)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("update lesson: %w", err)
	}

	return lessonService, nil
}
