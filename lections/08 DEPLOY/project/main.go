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

	"github.com/jackc/pgx/v5/pgxpool"

	"linkshortener/internal/apis/httpapi"
	"linkshortener/internal/repos/linkrepo"
	"linkshortener/internal/services/linkservice"
	"linkshortener/internal/settings"
)

const shutdownTimeout = 5 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := run(context.Background(), logger)
	if err != nil {
		logger.Error("app.failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	appSettings, err := settings.Load()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}

	logger.InfoContext(ctx, "settings.inited", "http", appSettings.HTTP, "pg_host", appSettings.Postgres.Host)

	pool, err := pgxpool.New(ctx, appSettings.Postgres.URL())
	if err != nil {
		return fmt.Errorf("create pg pool: %w", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping pg: %w", err)
	}

	linkRepo, err := linkrepo.New(ctx, pool)
	if err != nil {
		return fmt.Errorf("create link repo: %w", err)
	}

	server := httpapi.New(linkservice.New(linkRepo), appSettings.HTTP.BaseURL)

	go func() {
		<-ctx.Done()

		logger.WarnContext(ctx, "server.shutdown.begin")

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()

		shutdownErr := server.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			logger.ErrorContext(ctx, "server.shutdown.failed", "error", shutdownErr)
		}
	}()

	err = server.Start(appSettings.HTTP.Address())
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}
