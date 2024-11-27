package service

import (
	"context"
	"github.com/mxmrykov/aster-oauth-service/internal/cache"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"github.com/mxmrykov/aster-oauth-service/internal/grpc_server"
	"github.com/mxmrykov/aster-oauth-service/internal/http/external_server"
	"github.com/mxmrykov/aster-oauth-service/internal/store/postgres"
	"github.com/mxmrykov/aster-oauth-service/internal/store/redis"
	"github.com/mxmrykov/aster-oauth-service/pkg/clients/vault"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type IService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	IVault() vault.IVault
	ICache() cache.ICache
	OAuth() *config.OAuth
	Logger() *zerolog.Logger
	ClientStore() postgres.IClientStore
	UserStore() postgres.IUserStore
	RedisDc() redis.IRedisDc
	RedisTc() redis.IRedisTc
}

type Service struct {
	Zerolog *zerolog.Logger
	Cfg     *config.OAuth

	Vault    vault.IVault
	IRedisDc redis.IRedisDc
	IRedisTc redis.IRedisTc
	Cache    cache.ICache

	GrpcServer *grpc_server.GrpcServer
	HttpServer *external_server.Server

	IClientStore postgres.IClientStore
	IUserStore   postgres.IUserStore
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

	pgSecrets, err := v.GetSecretRepo(ctx, cfg.Vault.PostgresSecret.Path)

	if err != nil {
		logger.Fatal().Err(err).Msg("error getting dc-tc-redis credentials")
	}

	userStorage, err := postgres.NewStore[config.UserPostgres, postgres.UserStore](ctx, cfg.UserPostgres,
		pgSecrets[cfg.Vault.PostgresSecret.UPostgresUserName],
		pgSecrets[cfg.Vault.PostgresSecret.UPostgresSecretName],
	)

	if err != nil {
		logger.Fatal().Err(err).Msg("error getting user storage credentials")
	}

	clientStorage, err := postgres.NewStore[config.ClientPostgres, postgres.ClientStore](ctx, cfg.ClientPostgres,
		pgSecrets[cfg.Vault.PostgresSecret.ClPostgresUserName],
		pgSecrets[cfg.Vault.PostgresSecret.ClPostgresSecretName],
	)

	if err != nil {
		logger.Fatal().Err(err).Msg("error getting user storage credentials")
	}

	svc := &Service{
		Zerolog:      logger,
		Cfg:          cfg,
		Vault:        v,
		IRedisDc:     dc,
		IRedisTc:     tc,
		IClientStore: clientStorage,
		IUserStore:   userStorage,
	}

	grpcServer, err := grpc_server.NewServer(svc)

	if err != nil {
		logger.Fatal().Err(err).Msg("error starting grpc server")
	}

	svc.HttpServer = external_server.NewServer(logger, svc)
	svc.GrpcServer = grpcServer

	return svc, nil
}

func (s *Service) Start(ctx context.Context) error {
	erg, ctx := errgroup.WithContext(ctx)
	erg.Go(func() error {
		return s.HttpServer.Start(ctx)
	})
	erg.Go(func() error {
		return s.GrpcServer.Start(ctx)
	})

	return erg.Wait()
}
func (s *Service) Stop(ctx context.Context) error {
	_ = s.GrpcServer.Stop(ctx)
	return s.HttpServer.Stop(ctx)
}

func (s *Service) IVault() vault.IVault {
	return s.Vault
}
func (s *Service) ICache() cache.ICache               { return s.Cache }
func (s *Service) OAuth() *config.OAuth               { return s.Cfg }
func (s *Service) Logger() *zerolog.Logger            { return s.Zerolog }
func (s *Service) ClientStore() postgres.IClientStore { return s.IClientStore }
func (s *Service) UserStore() postgres.IUserStore     { return s.IUserStore }
func (s *Service) RedisDc() redis.IRedisDc            { return s.IRedisDc }
func (s *Service) RedisTc() redis.IRedisTc            { return s.IRedisTc }
