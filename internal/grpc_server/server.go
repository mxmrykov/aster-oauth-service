package grpc_server

import (
	"context"
	"fmt"
	"github.com/mxmrykov/aster-oauth-service/internal/cache"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	oauth "github.com/mxmrykov/aster-oauth-service/internal/proto/gen"
	"github.com/mxmrykov/aster-oauth-service/internal/store/postgres"
	"github.com/mxmrykov/aster-oauth-service/internal/store/redis"
	"github.com/mxmrykov/aster-oauth-service/pkg/clients/vault"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"net"
	"time"
)

type IGrpcServer interface {
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

type GrpcServer struct {
	svc         IGrpcServer
	lis         net.Listener
	S           *grpc.Server
	MaxPollTime time.Duration
}

func NewServer(svc IGrpcServer) (*GrpcServer, error) {
	s := grpc.NewServer()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", svc.OAuth().GrpcServer.Port))

	if err != nil {
		return nil, err
	}

	oauth.RegisterOAuthServer(s, &server{
		IRedisDc:     svc.RedisDc(),
		IRedisTc:     svc.RedisTc(),
		IVault:       svc.IVault(),
		Cfg:          svc.OAuth(),
		Logger:       svc.Logger(),
		IClientStore: svc.ClientStore(),
		IUserStore:   svc.UserStore(),
	})

	return &GrpcServer{
		lis:         lis,
		S:           s,
		MaxPollTime: svc.OAuth().GrpcServer.MaxPollTime,
	}, nil
}

func (s *GrpcServer) Start(_ context.Context) error {
	return s.S.Serve(s.lis)
}

func (s *GrpcServer) Stop(_ context.Context) error {
	s.S.GracefulStop()
	return nil
}
