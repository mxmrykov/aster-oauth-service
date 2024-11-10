package postgres

import "context"

func (c *ClientStore) GetClient(ctx context.Context, iaid string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.maxPoolInterval)

	defer cancel()

	const query = `
	select clientID, clientSecret from profiles.secrets where iaid = $1;
	`

	var clientID, clientSecret string

	if err := c.pool.QueryRow(ctx, query, iaid).Scan(&clientID, &clientSecret); err != nil {
		return "", "", err
	}

	return clientID, clientSecret, nil
}

func (c *ClientStore) PutClient(ctx context.Context, iaid, clientID, clientSecret string) error {
	ctx, cancel := context.WithTimeout(ctx, c.maxPoolInterval)

	defer cancel()

	const query = `
	update profiles.secrets set clientID = $1, clientSecret = $2 where iaid = $3;
	`

	_, err := c.pool.Exec(ctx, query, clientID, clientSecret, iaid)

	return err
}
