package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"time"
)

type IClientStore interface {
	Authorize(ctx context.Context, iaid string) (
		bool, string, error,
	)
}

type ClientStore struct {
	pool            *pgxpool.Pool
	maxPoolInterval time.Duration
}

func NewClientStore(ctx context.Context, cfg *config.ClientPostgres, user, password string) (*ClientStore, error) {
	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		user,
		password,
		cfg.Host,
		cfg.Port,
		cfg.DataBaseName,
	))

	if err != nil {
		return nil, err
	}

	return &ClientStore{
		pool:            pool,
		maxPoolInterval: cfg.MaxPoolInterval,
	}, nil
}
