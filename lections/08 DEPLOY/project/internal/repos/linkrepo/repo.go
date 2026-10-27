package linkrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS link (
	id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	short_link TEXT        NOT NULL UNIQUE,
	real_link  TEXT        NOT NULL
)`

type Repo struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, pool *pgxpool.Pool) (*Repo, error) {
	_, err := pool.Exec(ctx, createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("create link table: %w", err)
	}

	return &Repo{pool: pool}, nil
}
