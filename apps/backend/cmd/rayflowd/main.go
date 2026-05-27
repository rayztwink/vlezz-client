package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rayflow/rayflow-client/apps/backend/internal/app"
	"github.com/rayflow/rayflow-client/apps/backend/internal/config"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: app.Stdout()}).With().Timestamp().Logger()

	cfg := config.LoadAppConfig()
	rayflow, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize rayflowd")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := rayflow.Run(ctx); err != nil {
		logger.Fatal().Err(err).Msg("rayflowd stopped with error")
	}
}
