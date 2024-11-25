package postgres

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
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

	//go:embed queries/signUpInserts/*
	signUpUserInserts embed.FS

	//go:embed queries/exit.sql
	exitQuery string

	inserts = [4]string{
		"signUpUserInsert1.sql",
		"signUpUserInsert2.sql",
		"signUpUserInsert3.sql",
		"signUpUserInsert4.sql",
	}
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
	for _, f := range inserts {
		q, err := signUpUserInserts.ReadFile(fmt.Sprintf("queries/signUpInserts/%s", f))

		if err != nil {
			return err
		}

		if _, err = tx.Exec(
			ctx,
			string(q),
			[1]model.ExternalSignUpRequest{e},
			[1]model.InternalSignUpRequest{i},
		); err != nil {
			return err
		}
	}

	return nil
}

func (u *UserStore) Exit(ctx context.Context, iaid string, id int) error {
	ctx, cancel := context.WithTimeout(ctx, u.maxPoolInterval)

	defer cancel()

	_, err := u.pool.Exec(ctx, exitQuery, iaid, id)

	return err
}
