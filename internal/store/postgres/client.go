package postgres

import (
	"context"
	_ "embed"
	"github.com/jackc/pgx/v5"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
)

var (
	//go:embed queries/getClientData.sql
	getClientDataQuery string

	//go:embed queries/putClientData.sql
	putClientDataQuery string

	//go:embed queries/signupClient.sql
	signUpClient string
)

func (c *ClientStore) GetClient(ctx context.Context, iaid string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.maxPoolInterval)

	defer cancel()

	var clientID, clientSecret string

	if err := c.pool.QueryRow(ctx, getClientDataQuery, iaid).Scan(&clientID, &clientSecret); err != nil {
		return "", "", err
	}

	return clientID, clientSecret, nil
}

func (c *ClientStore) PutClient(ctx context.Context, iaid, clientID, clientSecret string) error {
	ctx, cancel := context.WithTimeout(ctx, c.maxPoolInterval)

	defer cancel()

	_, err := c.pool.Exec(ctx, putClientDataQuery, clientID, clientSecret, iaid)

	return err
}

func (c *ClientStore) SetClient(ctx context.Context, tx pgx.Tx, cr model.ClientSignUpRequest) error {
	_, err := tx.Exec(ctx, signUpClient, [1]model.ClientSignUpRequest{cr})

	return err
}
