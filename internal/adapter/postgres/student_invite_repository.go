package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
)

type StudentInviteRepository struct {
	pool *Pool
}

func NewStudentInviteRepository(pool *Pool) *StudentInviteRepository {
	return &StudentInviteRepository{
		pool: pool,
	}
}

func (r *StudentInviteRepository) CreateInvite(
	ctx context.Context,
	studentInvite domain.StudentInvite,
) (domain.StudentInvite, error) {
	query := `
	INSERT INTO public.student_invites (token_hash, student_name, requester_telegram_id,
	status,expires_at, requested_at, decided_at)	
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, token_hash, student_name, requester_telegram_id, 
	status, expires_at, requested_at, decided_at, created_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		studentInvite.TokenHash,
		studentInvite.StudentName,
		studentInvite.RequesterTelegramID,
		studentInvite.Status,
		studentInvite.ExpiresAt,
		studentInvite.RequestedAt,
		studentInvite.DecidedAt,
	)

	var studentInviteModel domain.StudentInvite
	err := row.Scan(
		&studentInviteModel.ID,
		&studentInviteModel.TokenHash,
		&studentInviteModel.StudentName,
		&studentInviteModel.RequesterTelegramID,
		&studentInviteModel.Status,
		&studentInviteModel.ExpiresAt,
		&studentInviteModel.RequestedAt,
		&studentInviteModel.DecidedAt,
		&studentInviteModel.CreatedAt,
	)

	pgError := &pgconn.PgError{}
	if err != nil {
		if errors.As(err, &pgError) {
			if pgError.Code == "23505" {
				return domain.StudentInvite{}, fmt.Errorf(
					"create invite: %w", domain.ErrAlreadyExists)
			}
		}
		return domain.StudentInvite{}, fmt.Errorf("create invite: %w", err)
	}

	return studentInviteModel, nil
}

func (r *StudentInviteRepository) RequestRegistration(
	ctx context.Context,
	tokenHash string,
	telegramID int64,
) (domain.StudentInvite, error) {
	query := `
	UPDATE public.student_invites
	SET	requester_telegram_id=$1, status='pending', requested_at=CURRENT_TIMESTAMP
	WHERE token_hash=$2 AND status='issued' AND expires_at > CURRENT_TIMESTAMP
	RETURNING id, token_hash, student_name, requester_telegram_id, 
	status, expires_at, requested_at, decided_at, created_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		telegramID,
		tokenHash,
	)

	var studentInviteModel domain.StudentInvite
	err := row.Scan(
		&studentInviteModel.ID,
		&studentInviteModel.TokenHash,
		&studentInviteModel.StudentName,
		&studentInviteModel.RequesterTelegramID,
		&studentInviteModel.Status,
		&studentInviteModel.ExpiresAt,
		&studentInviteModel.RequestedAt,
		&studentInviteModel.DecidedAt,
		&studentInviteModel.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.StudentInvite{}, fmt.Errorf(
				"invite operation with telegram id='%d': %w",
				telegramID,
				domain.ErrInviteUnavailable,
			)
		}
		return domain.StudentInvite{}, fmt.Errorf("scan request registration: %w", err)
	}

	return studentInviteModel, nil
}

func (r *StudentInviteRepository) GetPendingInvites(
	ctx context.Context,
) ([]domain.StudentInvite, error) {
	query := `
	SELECT id, token_hash, student_name, requester_telegram_id, 
	status, expires_at, requested_at, decided_at, created_at
	FROM public.student_invites 
	WHERE status='pending'
	ORDER BY requested_at;
	`

	rows, err := r.pool.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, fmt.Errorf("query pending invites: %w", err)
	}
	defer rows.Close()

	var studentInvitesModel []domain.StudentInvite
	for rows.Next() {
		var studentInviteModel domain.StudentInvite
		err := rows.Scan(
			&studentInviteModel.ID,
			&studentInviteModel.TokenHash,
			&studentInviteModel.StudentName,
			&studentInviteModel.RequesterTelegramID,
			&studentInviteModel.Status,
			&studentInviteModel.ExpiresAt,
			&studentInviteModel.RequestedAt,
			&studentInviteModel.DecidedAt,
			&studentInviteModel.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan invites: %w", err)
		}

		studentInvitesModel = append(studentInvitesModel, studentInviteModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("invites rows: %w", err)
	}

	return studentInvitesModel, nil
}

func (r *StudentInviteRepository) ApproveInvite(
	ctx context.Context,
	inviteID int64,
) (domain.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Student{}, fmt.Errorf("start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	querySelect := `
	SELECT student_name, requester_telegram_id
	FROM public.student_invites
	WHERE id=$1 AND status='pending'
	FOR UPDATE;
	`
	rowSelect := tx.QueryRow(
		ctx,
		querySelect,
		inviteID,
	)

	var studentInviteModel domain.StudentInvite
	errSelect := rowSelect.Scan(
		&studentInviteModel.StudentName,
		&studentInviteModel.RequesterTelegramID,
	)
	if errSelect != nil {
		if errors.Is(errSelect, pgx.ErrNoRows) {
			return domain.Student{}, fmt.Errorf(
				"invite operation with invite id='%d': %w",
				inviteID,
				domain.ErrInviteUnavailable,
			)
		}
		return domain.Student{}, fmt.Errorf("select studentInvite: %w", errSelect)
	}

	if studentInviteModel.RequesterTelegramID == nil {
		return domain.Student{}, domain.ErrConflict
	}

	queryInsert := `
	INSERT INTO public.students (telegram_id, name)
	VALUES ($1, $2)
	RETURNING id, telegram_id, name, is_active;	
	`

	rowInsert := tx.QueryRow(
		ctx,
		queryInsert,
		*studentInviteModel.RequesterTelegramID,
		studentInviteModel.StudentName,
	)

	var studentModel domain.Student
	errInsert := rowInsert.Scan(
		&studentModel.ID,
		&studentModel.TelegramID,
		&studentModel.Name,
		&studentModel.IsActive,
	)
	pgError := &pgconn.PgError{}
	if errInsert != nil {
		if errors.As(errInsert, &pgError) {
			if pgError.Code == "23505" {
				return domain.Student{}, fmt.Errorf("create student in invite: %w", domain.ErrAlreadyExists)
			}
		}
		return domain.Student{}, fmt.Errorf("create student approve: %w", errInsert)
	}

	queryUpdate := `
	UPDATE public.student_invites	
	SET status='approved', decided_at=CURRENT_TIMESTAMP
	WHERE id=$1 AND status='pending';
	`
	cmdTag, errUpdate := tx.Exec(
		ctx,
		queryUpdate,
		inviteID,
	)
	if errUpdate != nil {
		return domain.Student{}, fmt.Errorf("exec invite: %w", errUpdate)
	}

	if cmdTag.RowsAffected() != 1 {
		return domain.Student{}, fmt.Errorf(
			"invite with id='%d': %w",
			inviteID,
			domain.ErrInviteUnavailable,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Student{}, fmt.Errorf("commit invite: %w", err)
	}

	return studentModel, nil
}

func (r *StudentInviteRepository) RejectInvite(
	ctx context.Context,
	inviteID int64,
) (domain.StudentInvite, error) {
	query := `
	UPDATE public.student_invites 
	SET status='rejected', decided_at=CURRENT_TIMESTAMP
	WHERE id=$1 AND status='pending'
	RETURNING id, token_hash, student_name, requester_telegram_id, 
	status, expires_at, requested_at, decided_at, created_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		inviteID,
	)

	var studentInviteModel domain.StudentInvite
	err := row.Scan(
		&studentInviteModel.ID,
		&studentInviteModel.TokenHash,
		&studentInviteModel.StudentName,
		&studentInviteModel.RequesterTelegramID,
		&studentInviteModel.Status,
		&studentInviteModel.ExpiresAt,
		&studentInviteModel.RequestedAt,
		&studentInviteModel.DecidedAt,
		&studentInviteModel.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.StudentInvite{}, fmt.Errorf(
				"reject invite with id='%d': %w",
				inviteID,
				domain.ErrInviteUnavailable,
			)
		}
		return domain.StudentInvite{}, fmt.Errorf("reject invite: %w", err)
	}

	return studentInviteModel, nil
}
