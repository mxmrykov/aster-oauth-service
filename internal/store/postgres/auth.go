package postgres

import (
	"context"
)

func (u *UserStore) Authorize(ctx context.Context, iaid string) (
	bool, string, error,
) {
	ctx, cancel := context.WithTimeout(ctx, u.maxPoolInterval)

	defer cancel()

	const query = `
	select s.is_banned,
       p.value
	from users.signature s
         left join secrets.passwords p on s.iaid = p.iaid
	where s.iaid = $1;
	`

	var (
		banned bool
		pass   string
	)

	if err := u.pool.QueryRow(ctx, query, iaid).Scan(&banned, &pass); err != nil {
		return false, "", err
	}

	return banned, pass, nil
}

func (u *UserStore) IsPhoneInUse(ctx context.Context, phone string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, u.maxPoolInterval)

	defer cancel()

	const query = `
	select exists(select * from users.signature where phone = $1) as e;
	`

	var e bool
	err := u.pool.QueryRow(ctx, query, phone).Scan(&e)

	return e, err
}

func (u *UserStore) IsLoginInUse(ctx context.Context, login string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, u.maxPoolInterval)

	defer cancel()

	const query = `
	select exists(select * from users.signature where login = $1) as e;
	`

	var e bool
	err := u.pool.QueryRow(ctx, query, login).Scan(&e)

	return e, err
}
