package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type StudentRepository struct {
	pool *Pool
}

func NewStudentRepository(pool *Pool) *StudentRepository {
	return &StudentRepository{
		pool: pool,
	}
}

func (r *StudentRepository) CreateStudent(
	ctx context.Context,
	student domain.Student,
) (domain.Student, error) {
	query := `
		INSERT INTO public.students (telegram_id, name)	
		VALUES ($1, $2)
		RETURNING id, telegram_id, name, is_active;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		student.TelegramID,
		student.Name,
	)

	var studentDomain domain.Student
	err := row.Scan(
		&studentDomain.ID,
		&studentDomain.TelegramID,
		&studentDomain.Name,
		&studentDomain.IsActive,
	)
	pgError := &pgconn.PgError{}
	if err != nil {
		if errors.As(err, &pgError) {
			if pgError.Code == "23505" {
				return domain.Student{}, fmt.Errorf("create student: %w", domain.ErrAlreadyExists)
			}
		}
		return domain.Student{}, fmt.Errorf("create student: %w", err)
	}

	return studentDomain, nil
}

func (r *StudentRepository) GetStudentByID(
	ctx context.Context,
	studentID int64,
) (domain.Student, error) {
	query := `
		SELECT id, telegram_id, name, is_active
		FROM public.students
		WHERE id=$1;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		studentID,
	)

	var studentDomain domain.Student
	err := row.Scan(
		&studentDomain.ID,
		&studentDomain.TelegramID,
		&studentDomain.Name,
		&studentDomain.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Student{}, fmt.Errorf(
				"student with id='%d': %w",
				studentID, domain.ErrNotFound,
			)
		}
		return domain.Student{}, fmt.Errorf("get student by id: %w", err)
	}

	return studentDomain, nil
}

func (r *StudentRepository) GetStudentByTelegramID(
	ctx context.Context,
	telegramID int64,
) (domain.Student, error) {
	query := `
		SELECT id, telegram_id, name, is_active
		FROM public.students
		WHERE telegram_id=$1;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		telegramID,
	)

	var studentDomain domain.Student
	err := row.Scan(
		&studentDomain.ID,
		&studentDomain.TelegramID,
		&studentDomain.Name,
		&studentDomain.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Student{}, fmt.Errorf(
				"student with telegramID='%d': %w",
				telegramID, domain.ErrNotFound,
			)
		}
		return domain.Student{}, fmt.Errorf("get student by telegram ID: %w", err)
	}

	return studentDomain, nil
}

func (r *StudentRepository) GetStudents(
	ctx context.Context,
	isActive bool,
) ([]domain.Student, error) {
	query := `
		SELECT id, telegram_id, name, is_active
		FROM public.students
		WHERE is_active=$1;	
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		isActive,
	)
	if err != nil {
		return nil, fmt.Errorf("query students: %w", err)
	}
	defer rows.Close()

	var studentDomains []domain.Student
	for rows.Next() {
		var studentDomain domain.Student
		err := rows.Scan(
			&studentDomain.ID,
			&studentDomain.TelegramID,
			&studentDomain.Name,
			&studentDomain.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("scan students: %w", err)
		}

		studentDomains = append(studentDomains, studentDomain)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("student rows: %w", err)
	}

	return studentDomains, nil
}

func (r *StudentRepository) UpdateStudentName(
	ctx context.Context,
	name string,
	studentID int64,
) (domain.Student, error) {
	query := `
	UPDATE public.students
	SET name=$1
	WHERE id=$2
	RETURNING id, telegram_id, name, is_active;	
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		name,
		studentID,
	)

	var studentModel domain.Student
	err := row.Scan(
		&studentModel.ID,
		&studentModel.TelegramID,
		&studentModel.Name,
		&studentModel.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Student{}, fmt.Errorf(
				"student with id='%d': %w",
				studentID,
				domain.ErrNotFound,
			)
		}
		return domain.Student{}, fmt.Errorf("update student name: %w", err)
	}

	return studentModel, nil
}

func (r *StudentRepository) SetStudentActive(
	ctx context.Context,
	isActive bool,
	studentID int64,
) (domain.Student, error) {
	query := `
	UPDATE public.students
	SET is_active=$1
	WHERE id=$2
	RETURNING id, telegram_id, name, is_active;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		isActive,
		studentID,
	)

	var studentModel domain.Student
	err := row.Scan(
		&studentModel.ID,
		&studentModel.TelegramID,
		&studentModel.Name,
		&studentModel.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Student{}, fmt.Errorf(
				"student with id='%d': %w",
				studentID,
				domain.ErrNotFound,
			)
		}
		return domain.Student{}, fmt.Errorf("set student active: %w", err)
	}

	return studentModel, nil
}
