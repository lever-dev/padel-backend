package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lever-dev/padel-backend/internal/config"
	httpPkg "github.com/lever-dev/padel-backend/internal/controllers/http"
	courtRepo "github.com/lever-dev/padel-backend/internal/repositories/courts"
	reservationRepo "github.com/lever-dev/padel-backend/internal/repositories/reservation"
	"github.com/lever-dev/padel-backend/internal/repositories/users"
	"github.com/lever-dev/padel-backend/internal/services/auth"
	"github.com/lever-dev/padel-backend/internal/services/court"
	"github.com/lever-dev/padel-backend/internal/services/reservation"
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

		ctx := context.Background()

		reservationRepo := reservationRepo.NewRepository(cfg.Postgres.ConnectionURL)
		if err := reservationRepo.Connect(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to connect to postgres")
		}

		usersRepo := users.NewRepository(cfg.Postgres.ConnectionURL)
		if err := usersRepo.Connect(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to connect to postgres")
		}
		courtRepo := courtRepo.NewRepository(cfg.Postgres.ConnectionURL)
		if err := courtRepo.Connect(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to connect to postgres")
		}

		reservationService := reservation.NewService(reservationRepo, reservation.NewLocalLocker())
		courtService := court.NewService(courtRepo)
		authService := auth.NewService(usersRepo)

		reservationHandler := httpPkg.NewReservationHandler(reservationService)
		courtHandler := httpPkg.NewCourtHandler(courtService)
		authHandler := httpPkg.NewAuthHandler(authService)
		authMiddleware := httpPkg.NewAuthMiddleware(authService)

		router := httpPkg.NewRouter(reservationHandler, authHandler, courtHandler, authMiddleware)

		httpServer := http.Server{
			Addr:              cfg.HTTPServerAddr,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		}

		go func() {
			log.Info().Msg("started http server")

			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal().Err(err).Msg("failed on listen and serve")
			}
		}()

		// Graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigCh

		log.Info().Str("signal", sig.String()).Msg("application got signal")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		log.Info().Msg("closing all resources ...")

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed on shutdown server")
		}

		courtRepo.Close()
		reservationRepo.Close()
		usersRepo.Close()

		log.Info().Msg("Bye Bye !")

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
