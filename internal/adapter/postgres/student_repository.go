package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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

func (r *StudentRepository) GetStudentByTelegramID(
	ctx context.Context,
	telegramID int64,
) (domain.Student, error) {
	query := `
		SELECT id, telegram_id, is_active
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
		&studentDomain.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Student{}, fmt.Errorf(
				"student with telegramID='%d': %w",
				telegramID, domain.ErrNotFound,
			)
		}
		return domain.Student{}, fmt.Errorf("scan row: %w", err)
	}

	return studentDomain, nil
}
