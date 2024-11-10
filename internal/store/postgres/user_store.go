package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type IUserStore interface {
	Authorize(ctx context.Context, iaid string) (
		bool, string, error,
	)
	IsPhoneInUse(ctx context.Context, phone string) (bool, error)
	IsLoginInUse(ctx context.Context, login string) (bool, error)
}

type UserStore struct {
	pool            *pgxpool.Pool
	maxPoolInterval time.Duration
}
