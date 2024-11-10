package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type IClientStore interface {
	GetClient(ctx context.Context, iaid string) (string, string, error)
	PutClient(ctx context.Context, iaid, clientID, clientSecret string) error
}

type ClientStore struct {
	pool            *pgxpool.Pool
	maxPoolInterval time.Duration
}
