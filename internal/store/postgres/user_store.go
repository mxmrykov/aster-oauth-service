package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
)

type IUserStore interface {
	Authorize(ctx context.Context, iaid string) (
		bool, string, error,
	)
	IsPhoneInUse(ctx context.Context, phone string) (bool, error)
	IsLoginInUse(ctx context.Context, login string) (bool, error)
	SignUpUser(ctx context.Context,
		tx pgx.Tx,
		e model.ExternalSignUpRequest,
		i model.InternalSignUpRequest,
	) error
	Exit(ctx context.Context, iaid, signature string) error
	EnterSession(ctx context.Context, i model.EnterSession) error
	ExtractEaid(ctx context.Context, iaid string) (int, string, error)

	BeginTx(ctx context.Context) (pgx.Tx, error)
}

type UserStore struct {
	pool            *pgxpool.Pool
	maxPoolInterval time.Duration
}

func (u *UserStore) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return nil, err
	}

	return tx, nil
}
