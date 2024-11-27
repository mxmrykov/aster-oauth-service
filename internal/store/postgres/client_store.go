package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
)

type IClientStore interface {
	GetClient(ctx context.Context, iaid string) (string, string, error)
	PutClient(ctx context.Context, iaid, clientID, clientSecret string) error
	SetClient(ctx context.Context, tx pgx.Tx, cr model.ClientSignUpRequest) error
	CheckClient(ctx context.Context, cid, csecret, iaid string) error

	BeginTx(ctx context.Context) (pgx.Tx, error)
}

type ClientStore struct {
	pool            *pgxpool.Pool
	maxPoolInterval time.Duration
}

func (u *ClientStore) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return nil, err
	}

	return tx, nil
}
