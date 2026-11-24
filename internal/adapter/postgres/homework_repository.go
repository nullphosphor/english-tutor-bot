package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type HomeworkRepository struct {
	pool *Pool
}

func NewHomeworkRepository(pool *Pool) *HomeworkRepository {
	return &HomeworkRepository{
		pool: pool,
	}
}

func (r *HomeworkRepository) CreateHomework(
	ctx context.Context,
	homework domain.Homework,
) (domain.Homework, error) {
	query := `
	INSERT INTO public.homeworks (student_id, title,
	description, due_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id, student_id, version, title, description, due_at, created_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		homework.StudentID,
		homework.Title,
		homework.Description,
		homework.DueAt,
	)

	var homeworkDomain domain.Homework
	err := row.Scan(
		&homeworkDomain.ID,
		&homeworkDomain.StudentID,
		&homeworkDomain.Version,
		&homeworkDomain.Title,
		&homeworkDomain.Description,
		&homeworkDomain.DueAt,
		&homeworkDomain.CreatedAt,
	)
	if err != nil {
		return domain.Homework{}, fmt.Errorf("can't create homework: %w", err)
	}

	return homeworkDomain, nil
}

func (r *HomeworkRepository) GetHomeworkByID(
	ctx context.Context,
	homeworkID int64,
) (domain.Homework, error) {
	query := `
	SELECT id, student_id, version, title, description, due_at, created_at
	FROM public.homeworks
	WHERE id=$1;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		homeworkID,
	)

	var homeworkDomain domain.Homework
	err := row.Scan(
		&homeworkDomain.ID,
		&homeworkDomain.StudentID,
		&homeworkDomain.Version,
		&homeworkDomain.Title,
		&homeworkDomain.Description,
		&homeworkDomain.DueAt,
		&homeworkDomain.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Homework{}, fmt.Errorf(
				"homework with id='%d': %w",
				homeworkID, domain.ErrNotFound,
			)
		}
		return domain.Homework{}, fmt.Errorf("can't get homework: %w", err)
	}

	return homeworkDomain, nil
}

func (r *HomeworkRepository) GetHomeworksByStudentID(
	ctx context.Context,
	studentID int64,
) ([]domain.Homework, error) {
	query := `
	SELECT id, student_id, version, title, description, due_at, created_at
	FROM public.homeworks
	WHERE student_id=$1
	ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("select rows: %w", err)
	}
	defer rows.Close()

	var homeworksDomain []domain.Homework
	for rows.Next() {
		var homeworkDomain domain.Homework
		err := rows.Scan(
			&homeworkDomain.ID,
			&homeworkDomain.StudentID,
			&homeworkDomain.Version,
			&homeworkDomain.Title,
			&homeworkDomain.Description,
			&homeworkDomain.DueAt,
			&homeworkDomain.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan rows: %w", err)
		}
		homeworksDomain = append(homeworksDomain, homeworkDomain)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return homeworksDomain, nil
}

func (r *HomeworkRepository) UpdateHomework(
	ctx context.Context,
	homework domain.Homework,
) (domain.Homework, error) {
	query := `
	UPDATE public.homeworks
	SET title=$1, description=$2, due_at=$3, version=version+1
	WHERE id=$4 AND version=$5
	RETURNING id, student_id, version, title, description, due_at, created_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		homework.Title,
		homework.Description,
		homework.DueAt,
		homework.ID,
		homework.Version,
	)

	var homeworkDomain domain.Homework
	err := row.Scan(
		&homeworkDomain.ID,
		&homeworkDomain.StudentID,
		&homeworkDomain.Version,
		&homeworkDomain.Title,
		&homeworkDomain.Description,
		&homeworkDomain.DueAt,
		&homeworkDomain.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Homework{}, fmt.Errorf(
				"homework with id='%d': %w",
				homework.ID, domain.ErrConflict,
			)
		}
		return domain.Homework{}, fmt.Errorf("can't update homework: %w", err)
	}

	return homeworkDomain, nil
}

func (r *HomeworkRepository) DeleteHomework(
	ctx context.Context,
	homeworkID int64,
	version int64,
) error {
	query := `
	DELETE FROM public.homeworks
	WHERE id=$1 AND version=$2;
	`

	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		homeworkID,
		version,
	)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf(
			"homework with id='%d': %w",
			homeworkID, domain.ErrConflict,
		)
	}

	return nil
}
