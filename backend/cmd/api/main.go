package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samfiallos/owenone/backend/internal/app"
	"github.com/samfiallos/owenone/backend/internal/avatars"
	"github.com/samfiallos/owenone/backend/internal/config"
	"github.com/samfiallos/owenone/backend/internal/httpapi"
	"github.com/samfiallos/owenone/backend/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	config, err := config.Load()
	if err != nil {
		return err
	}

	rootContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repository, err := store.New(rootContext, config.DatabaseURL, config.MaxDBConnections)
	if err != nil {
		return err
	}
	defer repository.Close()

	if config.AutoMigrate {
		if err := repository.Migrate(rootContext); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
	}

	service := app.NewService(repository)
	avatarStorage, err := avatars.NewStorage(config.AvatarDir)
	if err != nil {
		return fmt.Errorf("avatar storage: %w", err)
	}
	server := &http.Server{
		Addr:              config.Address,
		Handler:           httpapi.New(service, logger, config.FrontendOrigin, avatarStorage),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API listening", "address", config.Address)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-rootContext.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		return nil
	}
}
