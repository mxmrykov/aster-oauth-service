package grpc_server

import (
	"context"
	"time"

	"github.com/mxmrykov/aster-oauth-service/internal/config"
	oauth "github.com/mxmrykov/aster-oauth-service/internal/proto/gen"
	"github.com/mxmrykov/aster-oauth-service/internal/store/postgres"
	"github.com/mxmrykov/aster-oauth-service/internal/store/redis"
	"github.com/mxmrykov/aster-oauth-service/pkg/clients/vault"
	"github.com/mxmrykov/aster-oauth-service/pkg/hashing"
	"github.com/mxmrykov/aster-oauth-service/pkg/sid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	oauth.UnimplementedOAuthServer
	IRedisDc     redis.IRedisDc
	IRedisTc     redis.IRedisTc
	IClientStore postgres.IClientStore
	IUserStore   postgres.IUserStore
	IVault       vault.IVault
	Cfg          *config.OAuth
	Logger       *zerolog.Logger
}

func (s *server) Authorize(ctx context.Context, in *oauth.AuthorizeRequest) (
	*oauth.AuthorizeResponse, error,
) {
	s.Logger.Info().Msg("New call")

	realCc, err := s.IVault.GetSecret(ctx, s.Cfg.Vault.TokenRepo.Path, s.Cfg.Vault.TokenRepo.OAuthJwtSecretName)

	if err != nil {
		s.Logger.Err(err).Send()
		return nil, status.Error(codes.Internal, err.Error())
	}

	switch {
	case in.ConfirmCode == "":
		err = status.Error(codes.InvalidArgument, "confirm code is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "confirm code is required",
		}, nil
	case in.ConfirmCode != realCc:
		err = status.Error(codes.InvalidArgument, "confirm code is incorrect")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "confirm code is incorrect",
		}, nil
	case in.Login == "":
		err = status.Error(codes.InvalidArgument, "login is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "login is required",
		}, nil
	case in.Password == "":
		err = status.Error(codes.InvalidArgument, "password is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "password is required",
		}, nil
	case in.IAID == "":
		err = status.Error(codes.InvalidArgument, "IAID is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "IAID is required",
		}, nil
	case in.ASID == "":
		err = status.Error(codes.InvalidArgument, "ASID is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "ASID is required",
		}, nil
	}

	innerIaid, err := s.IRedisDc.Get(ctx, in.ASID)

	if err != nil {
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "No such session",
		}, nil
	}

	iaid, err := sid.Validate(in.IAID)

	switch {
	case iaid.Subscriber != innerIaid,
		err != nil:
		return &oauth.AuthorizeResponse{
			Error: "Auth error",
		}, nil
	}

	isBanned, pwd, err := s.IUserStore.Authorize(ctx, iaid.Subscriber)

	switch {
	case err != nil:
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "No such user",
		}, nil
	case isBanned:
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "User is banned",
		}, nil
	case !hashing.Check(in.Password, pwd):
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "Incorrect password",
		}, nil
	}

	clientID, clientSecret, err := s.IClientStore.GetClient(ctx, innerIaid)

	if err != nil {
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "Cannot get client",
		}, nil
	}

	oac := sid.New(in.IAID, time.Minute)

	if err = s.IRedisDc.SetOAuthCode(ctx, oac, innerIaid); err != nil {
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "Cannot set oauth code",
		}, nil
	}

	return &oauth.AuthorizeResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		OAuthCode:    oac,
		Error:        "",
	}, nil
}
