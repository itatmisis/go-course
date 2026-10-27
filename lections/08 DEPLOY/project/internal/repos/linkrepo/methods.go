package linkrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("link not found")

type Link struct {
	ID        string
	CreatedAt time.Time
	ShortLink string
	RealLink  string
}

func (r *Repo) CreateLink(ctx context.Context, shortLink string, realLink string) error {
	_, err := r.pool.Exec(
		ctx,
		`INSERT INTO link (short_link, real_link) VALUES ($1, $2)`,
		shortLink,
		realLink,
	)
	if err != nil {
		return fmt.Errorf("insert link: %w", err)
	}

	return nil
}

func (r *Repo) GetLink(ctx context.Context, shortLink string) (Link, error) {
	var link Link

	err := r.pool.QueryRow(
		ctx,
		`SELECT id::text, created_at, short_link, real_link FROM link WHERE short_link = $1`,
		shortLink,
	).Scan(&link.ID, &link.CreatedAt, &link.ShortLink, &link.RealLink)
	if errors.Is(err, pgx.ErrNoRows) {
		return Link{}, ErrNotFound
	}

	if err != nil {
		return Link{}, fmt.Errorf("select link: %w", err)
	}

	return link, nil
}
