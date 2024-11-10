package postgres

import (
	"context"
)

func (c *ClientStore) Authorize(ctx context.Context, iaid string) (
	bool, string, error,
) {
	ctx, cancel := context.WithTimeout(ctx, c.maxPoolInterval)

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

	if err := c.pool.QueryRow(ctx, query, iaid).Scan(&banned, &pass); err != nil {
		return false, "", err
	}

	return banned, pass, nil
}
