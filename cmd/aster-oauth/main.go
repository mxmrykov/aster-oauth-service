package main

import (
	"context"
	"os"

	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"github.com/mxmrykov/aster-oauth-service/internal/service"
	"github.com/mxmrykov/aster-oauth-service/pkg/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func main() {
	os.Setenv("BUILD_ENV", "local")
	os.Setenv("VAULT_AUTH_TOKEN", "hvs.CAESIGTYh1nuNhYQNytin2slQdg7vDEHExB9_R2VCbwlqNR-Gh4KHGh2cy4yWFNDQkROUFh1VU8zU0ExU29TNHVqbzg")
	_, ctx := errgroup.WithContext(context.Background())
	cfg, logger, err := config.InitConfig()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize config")
	}

	logger.Info().Msg("config created")
	logger.Info().Msg("initializing service")

	s, err := service.NewService(ctx, cfg, logger)

	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize service")
	}

	logger.Info().Msg("starting service")

	go func() {
		if err = s.Start(ctx); err != nil {
			logger.Fatal().Err(err).Msg("failed to start service")
		}
	}()

	<-utils.GracefulShutDown()

	logger.Info().Msg("graceful shutdown")

	_ = s.Stop(ctx)

	logger.Info().Msg("service stopped")
}
