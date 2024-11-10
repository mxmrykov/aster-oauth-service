package service

import (
	"context"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"github.com/mxmrykov/aster-oauth-service/internal/store/redis"
	"github.com/mxmrykov/aster-oauth-service/pkg/clients/vault"
	"github.com/rs/zerolog"
)

type IService interface {
}

type Service struct {
	Zerolog *zerolog.Logger
	Cfg     *config.OAuth

	Vault    vault.IVault
	IRedisDc redis.IRedisDc
	IRedisTc redis.IRedisTc
}

func NewService(ctx context.Context, cfg *config.OAuth, logger *zerolog.Logger) (IService, error) {
	v, err := vault.NewVault(&cfg.Vault)

	if err != nil {
		logger.Fatal().Err(err).Msg("error initializing vault client")
	}

	secrets, err := v.GetSecretRepo(ctx, cfg.Vault.RedisSecret.Path)

	if err != nil {
		logger.Fatal().Err(err).Msg("error getting dc-tc-redis credentials")
	}

	dc, tc := redis.NewRedisDc(
		&cfg.DcRedis,
		secrets[cfg.Vault.RedisSecret.DcRedisUserName],
		secrets[cfg.Vault.RedisSecret.DcRedisSecretName],
	), redis.NewRedisTc(
		&cfg.TcRedis,
		secrets[cfg.Vault.RedisSecret.TcRedisUserName],
		secrets[cfg.Vault.RedisSecret.TcRedisSecretName],
	)

	if err != nil {
		logger.Fatal().Err(err).Msg("error getting dc-tc-redis password")
	}

	return &Service{
		Zerolog:  logger,
		Cfg:      cfg,
		Vault:    v,
		IRedisDc: dc,
		IRedisTc: tc,
	}, nil
}
