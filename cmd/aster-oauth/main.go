package main

import (
	"context"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func main() {
	erg, ctx := errgroup.WithContext(context.Background())

	log.Info().Timestamp().Msg("starting oauth service...")
	log.Info().Timestamp().Msg("initializing config...")

	cfg, logger, err := config.InitConfig()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize config")
	}

	log.Info().Timestamp().Msg("initializing service...")

}
