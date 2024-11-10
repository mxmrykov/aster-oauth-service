package grpc_server

import (
	"context"
	"errors"
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
	"time"
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
		}, err
	case in.ConfirmCode != realCc:
		err = status.Error(codes.InvalidArgument, "confirm code is incorrect")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "confirm code is incorrect",
		}, err
	case in.Login == "":
		err = status.Error(codes.InvalidArgument, "login is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "login is required",
		}, err
	case in.Password == "":
		err = status.Error(codes.InvalidArgument, "password is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "password is required",
		}, err
	case in.IAID == "":
		err = status.Error(codes.InvalidArgument, "IAID is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "IAID is required",
		}, err
	case in.ASID == "":
		err = status.Error(codes.InvalidArgument, "ASID is required")
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "ASID is required",
		}, err
	}

	innerIaid, err := s.IRedisDc.Get(ctx, in.ASID)

	if err != nil {
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "No such session",
		}, status.Error(codes.Internal, err.Error())
	}

	_, err = s.IRedisDc.Get(ctx, in.ASID)

	alive := errors.Is(err, redis.ErrorNotFound)

	switch {
	case !alive,
		innerIaid != in.IAID,
		sid.Validate(in.ASID) != nil:
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "Auth error",
		}, status.Error(codes.Unauthenticated, err.Error())
	}

	isBanned, pwd, err := s.IUserStore.Authorize(ctx, in.IAID)

	switch {
	case err != nil:
		s.Logger.Err(err).Send()
		return nil, status.Error(codes.Internal, err.Error())
	case isBanned:
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "User is banned",
		}, status.Error(codes.PermissionDenied, "User is banned")
	case !hashing.Check(in.Password, pwd):
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "Incorrect password",
		}, status.Error(codes.Unauthenticated, "Incorrect password")
	}

	clientID, clientSecret, err := s.IClientStore.GetClient(ctx, in.IAID)

	if err != nil {
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "Cannot get client",
		}, status.Error(codes.Internal, err.Error())
	}

	oac := sid.New(in.IAID, time.Minute)

	if err = s.IRedisDc.SetOAuthCode(ctx, oac, in.IAID); err != nil {
		s.Logger.Err(err).Send()
		return &oauth.AuthorizeResponse{
			Error: "Cannot set oauth code",
		}, status.Error(codes.Internal, "Cannot set oauth code")
	}

	return &oauth.AuthorizeResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		OAuthCode:    oac,
		Error:        "",
	}, nil
}
