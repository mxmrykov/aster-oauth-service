package postgres

import (
	"context"
	_ "embed"
	"github.com/jackc/pgx/v5"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
)

var (
	//go:embed queries/authorize.sql
	authorizeQuery string

	//go:embed queries/isPhoneInUse.sql
	isPhoneInUseQuery string

	//go:embed queries/isLoginInUse.sql
	isLoginInUseQuery string

	//go:embed queries/signupUser.sql
	signupQuery string
)

func (u *UserStore) Authorize(ctx context.Context, iaid string) (
	bool, string, error,
) {
	ctx, cancel := context.WithTimeout(ctx, u.maxPoolInterval)

	defer cancel()

	var (
		banned bool
		pass   string
	)

	if err := u.pool.QueryRow(ctx, authorizeQuery, iaid).Scan(&banned, &pass); err != nil {
		return false, "", err
	}

	return banned, pass, nil
}

func (u *UserStore) IsPhoneInUse(ctx context.Context, phone string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, u.maxPoolInterval)

	defer cancel()

	var e bool
	err := u.pool.QueryRow(ctx, isPhoneInUseQuery, phone).Scan(&e)

	return e, err
}

func (u *UserStore) IsLoginInUse(ctx context.Context, login string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, u.maxPoolInterval)

	defer cancel()

	var e bool
	err := u.pool.QueryRow(ctx, isLoginInUseQuery, login).Scan(&e)

	return e, err
}

func (u *UserStore) SignUpUser(ctx context.Context,
	tx pgx.Tx,
	e model.ExternalSignUpRequest,
	i model.InternalSignUpRequest,
) error {
	_, err := tx.Exec(ctx, signupQuery, e, i)

	return err
}
