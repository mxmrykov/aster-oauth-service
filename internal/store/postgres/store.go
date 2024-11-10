package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"time"
)

func NewStore[C config.ClientPostgres | config.UserPostgres, S ClientStore | UserStore](
	ctx context.Context, cfg C, user, password string,
) (*S, error) {
	var (
		host, dbname    string
		port            int
		maxPollInterval time.Duration
	)

	switch localCfg := any(cfg).(type) {
	case config.ClientPostgres:
		host = localCfg.Host
		port = localCfg.Port
		dbname = localCfg.DataBaseName
		maxPollInterval = localCfg.MaxPoolInterval
	case config.UserPostgres:
		host = localCfg.Host
		port = localCfg.Port
		dbname = localCfg.DataBaseName
		maxPollInterval = localCfg.MaxPoolInterval
	}

	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		user,
		password,
		host,
		port,
		dbname,
	))

	return &S{pool: pool, maxPoolInterval: maxPollInterval}, err
}
