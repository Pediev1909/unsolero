package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"rigmark/internal/platform/config"
	"rigmark/internal/platform/database"
	"rigmark/internal/platform/observability"
)

func main() {
	logger := observability.NewJSONLogger(os.Stdout, slog.LevelInfo)
	if err := run(logger); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	pool, err := database.OpenPool(ctx, database.PoolConfig{
		URL: cfg.Database.URL, ApplicationName: "unsolero-migrate", MaxConnections: 1,
		MinConnections: 0, MaxConnectionLifetime: cfg.Database.MaxConnectionLifetime,
		MaxConnectionIdleTime: cfg.Database.MaxConnectionIdleTime, HealthCheckPeriod: cfg.Database.HealthCheckPeriod,
		ConnectTimeout: cfg.Database.ConnectTimeout, StatementTimeout: cfg.Database.MigrationTimeout,
		LockTimeout: cfg.Database.MigrationTimeout, IdleTransactionTimeout: cfg.Database.MigrationTimeout,
	})
	if err != nil {
		return err
	}
	defer pool.Close()
	migrationCtx, cancel := context.WithTimeout(ctx, cfg.Database.MigrationTimeout)
	defer cancel()
	if err := database.ApplyMigrations(migrationCtx, pool, os.DirFS(cfg.Migrations.Directory)); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	logger.Info("database migrations are current")
	return nil
}
