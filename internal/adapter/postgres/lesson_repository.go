package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type LessonRepository struct {
	pool *Pool
}

func NewLessonRepository(pool *Pool) *LessonRepository {
	return &LessonRepository{
		pool: pool,
	}
}

func (r *LessonRepository) CreateLesson(
	ctx context.Context,
	lesson domain.Lesson,
) (domain.Lesson, error) {
	query := `
		INSERT INTO public.lessons (student_id, date_time,
		meeting_app, meeting_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, student_id, date_time, meeting_app,
		meeting_url, status, created_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		lesson.StudentID,
		lesson.DateTime,
		lesson.MeetingApp,
		lesson.MeetingURL,
	)

	var lessonDomain domain.Lesson
	err := row.Scan(
		&lessonDomain.ID,
		&lessonDomain.StudentID,
		&lessonDomain.DateTime,
		&lessonDomain.MeetingApp,
		&lessonDomain.MeetingURL,
		&lessonDomain.LessonStatus,
		&lessonDomain.CreatedAt,
	)

	if err != nil {
		return domain.Lesson{}, fmt.Errorf("can't create lesson: %w", err)
	}

	return lessonDomain, nil
}

func (r *LessonRepository) GetLessonByID(
	ctx context.Context,
	id int64,
) (domain.Lesson, error) {
	query := `
		SELECT id, student_id, date_time, meeting_app,
		meeting_url, status, created_at FROM public.lessons
		WHERE id=$1	
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		id,
	)

	var lessonDomain domain.Lesson
	err := row.Scan(
		&lessonDomain.ID,
		&lessonDomain.StudentID,
		&lessonDomain.DateTime,
		&lessonDomain.MeetingApp,
		&lessonDomain.MeetingURL,
		&lessonDomain.LessonStatus,
		&lessonDomain.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Lesson{}, fmt.Errorf(
				"lesson with id='%d': %w",
				id, domain.ErrNotFound,
			)
		}
		return domain.Lesson{}, fmt.Errorf("can't get lesson by id: %w", err)
	}

	return lessonDomain, nil
}
