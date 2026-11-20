package postgres

import (
	"context"
	"fmt"

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
	RETURNING id, student_id, title, description, due_at, created_at;
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

func (r *HomeworkRepository) GetHomeworksByStudentID(
	ctx context.Context,
	studentID int64,
) ([]domain.Homework, error) {
	query := `
	SELECT id, student_id, title, description, due_at, created_at
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
