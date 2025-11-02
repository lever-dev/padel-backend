package main

import (
	"fmt"

	"github.com/lever-dev/padel-backend/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the gRPC server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load config")
		}

		if err := initLogger(cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to init logger")
		}

		return nil
	},
}

func initLogger(cfg config.Config) error {
	logLvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parse level: %w", err)

	}

	zerolog.SetGlobalLevel(logLvl)

	return nil
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
