package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(
	ctx context.Context,
	databaseURL string,
) (*Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &Pool{pool}, nil
}
