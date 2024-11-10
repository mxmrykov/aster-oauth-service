package grpc_server

import (
	"fmt"
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
	Serve() error
	GracefulStop()
}

type GrpcServer struct {
	lis         net.Listener
	S           *grpc.Server
	MaxPollTime time.Duration
}

func NewServer(
	dc redis.IRedisDc,
	tc redis.IRedisTc,
	vault vault.IVault,
	cfg *config.OAuth,
	l *zerolog.Logger,
	IClientStore postgres.IClientStore,
) (*GrpcServer, error) {
	s := grpc.NewServer()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GrpcServer.Port))

	if err != nil {
		return nil, err
	}

	oauth.RegisterOAuthServer(s, &server{
		IRedisDc:     dc,
		IRedisTc:     tc,
		IVault:       vault,
		Cfg:          cfg,
		Logger:       l,
		IClientStore: IClientStore,
	})

	return &GrpcServer{
		lis:         lis,
		S:           s,
		MaxPollTime: cfg.GrpcServer.MaxPollTime,
	}, nil
}

func (s *GrpcServer) Serve() error {
	return s.S.Serve(s.lis)
}

func (s *GrpcServer) GracefulStop() {
	s.S.GracefulStop()
}
