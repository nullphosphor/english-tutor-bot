package service

import (
	"context"
	"fmt"

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
}

func (s *LessonService) CreateLesson(
	ctx context.Context,
	lesson domain.Lesson,
) (domain.Lesson, error) {
	if err := lesson.Validate(); err != nil {
		return domain.Lesson{}, fmt.Errorf("validate lesson from service: %w", err)
	}

	lesson, err := s.lessonRepository.CreateLesson(ctx, lesson)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("create lesson from service: %w", err)
	}

	return lesson, nil
}

func (s *LessonService) GetLessonByID(
	ctx context.Context,
	id int64,
) (domain.Lesson, error) {
	lesson, err := s.lessonRepository.GetLessonByID(ctx, id)
	if err != nil {
		return domain.Lesson{}, fmt.Errorf("get lesson by id from service: %w", err)
	}

	return lesson, nil
}
