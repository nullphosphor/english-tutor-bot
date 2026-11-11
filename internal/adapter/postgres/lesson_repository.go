package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		RETURNING id, student_id, date_time, started_at, finished_at,
		meeting_app, meeting_url, status, created_at, version;
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
		&lessonDomain.StartedAt,
		&lessonDomain.FinishedAt,
		&lessonDomain.MeetingApp,
		&lessonDomain.MeetingURL,
		&lessonDomain.LessonStatus,
		&lessonDomain.CreatedAt,
		&lesson.Version,
	)

	pgError := &pgconn.PgError{}
	if err != nil {
		if errors.As(err, &pgError) {
			if pgError.Code == "23505" &&
				pgError.ConstraintName == "lessons_active_student_datetime_uidx" {
				return domain.Lesson{}, fmt.Errorf("pgx query: %w", domain.ErrDuplicateLesson)
			}
		}
		return domain.Lesson{}, fmt.Errorf("can't create lesson: %w", err)
	}

	return lessonDomain, nil
}

func (r *LessonRepository) GetLessonByID(
	ctx context.Context,
	id int64,
) (domain.Lesson, error) {
	query := `
		SELECT id, student_id, date_time, started_at, finished_at, 
		meeting_app, meeting_url, status, created_at, version 
		FROM public.lessons
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
		&lessonDomain.StartedAt,
		&lessonDomain.FinishedAt,
		&lessonDomain.MeetingApp,
		&lessonDomain.MeetingURL,
		&lessonDomain.LessonStatus,
		&lessonDomain.CreatedAt,
		&lessonDomain.Version,
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

func (r *LessonRepository) GetLessonsByPeriod(
	ctx context.Context,
	from, to time.Time,
) ([]domain.Lesson, error) {
	query := `
		SELECT id, student_id, date_time, started_at, finished_at,
		meeting_app, meeting_url, status, created_at, version
		FROM public.lessons
		WHERE date_time >= $1 AND date_time < $2
		ORDER BY date_time ASC, id ASC;
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		from,
		to,
	)
	if err != nil {
		return []domain.Lesson{}, fmt.Errorf("select lessons: %w", err)
	}
	defer rows.Close()

	var lessonDomains []domain.Lesson
	for rows.Next() {
		var lessonDomain domain.Lesson
		err := rows.Scan(
			&lessonDomain.ID,
			&lessonDomain.StudentID,
			&lessonDomain.DateTime,
			&lessonDomain.StartedAt,
			&lessonDomain.FinishedAt,
			&lessonDomain.MeetingApp,
			&lessonDomain.MeetingURL,
			&lessonDomain.LessonStatus,
			&lessonDomain.CreatedAt,
			&lessonDomain.Version,
		)
		if err != nil {
			return []domain.Lesson{}, fmt.Errorf("scan lessons: %w", err)
		}

		lessonDomains = append(lessonDomains, lessonDomain)
	}

	if err := rows.Err(); err != nil {
		return []domain.Lesson{}, fmt.Errorf("rows err: %w", err)
	}

	return lessonDomains, nil
}

func (r *LessonRepository) UpdateLesson(
	ctx context.Context,
	lesson domain.Lesson,
) (domain.Lesson, error) {
	query := `
		UPDATE public.lessons
		SET	started_at=$1, finished_at=$2, status=$3, version=version+1
		WHERE id=$4 AND version=$5
		RETURNING id, student_id, date_time, started_at, finished_at,
		meeting_app, meeting_url, status, created_at, version;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		lesson.StartedAt,
		lesson.FinishedAt,
		lesson.LessonStatus,
		lesson.ID,
		lesson.Version,
	)

	var lessonDomain domain.Lesson
	err := row.Scan(
		&lessonDomain.ID,
		&lessonDomain.StudentID,
		&lessonDomain.DateTime,
		&lessonDomain.StartedAt,
		&lessonDomain.FinishedAt,
		&lessonDomain.MeetingApp,
		&lessonDomain.MeetingURL,
		&lessonDomain.LessonStatus,
		&lessonDomain.CreatedAt,
		&lessonDomain.Version,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Lesson{}, fmt.Errorf(
				"lesson with id=%d concurrently accessed:: %w",
				lesson.ID,
				domain.ErrConflict,
			)
		}
		return domain.Lesson{}, fmt.Errorf("can't update lesson: %w", err)
	}

	return lessonDomain, nil
}

func (r *LessonRepository) ExistsActiveLesson(
	ctx context.Context,
	studentID int64,
	dateTime time.Time,
) (bool, error) {
	query := `
		SELECT EXISTS (
		SELECT 1 FROM public.lessons
		WHERE student_id=$1 AND 
		date_time=$2 AND
		status IN ('scheduled', 'in_progress')
);
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		studentID,
		dateTime,
	)

	var exists bool
	if err := row.Scan(
		&exists,
	); err != nil {
		return false, fmt.Errorf("scan row: %w", err)
	}

	return exists, nil
}
