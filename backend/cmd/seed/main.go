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
		logger.Error("seed failed", "error", err)
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
		URL: cfg.Database.URL, ApplicationName: "unsolero-seed", MaxConnections: 1,
		MinConnections: 0, MaxConnectionLifetime: cfg.Database.MaxConnectionLifetime,
		MaxConnectionIdleTime: cfg.Database.MaxConnectionIdleTime, HealthCheckPeriod: cfg.Database.HealthCheckPeriod,
		ConnectTimeout: cfg.Database.ConnectTimeout, StatementTimeout: cfg.Database.MigrationTimeout,
		LockTimeout: cfg.Database.MigrationTimeout, IdleTransactionTimeout: cfg.Database.MigrationTimeout,
	})
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := database.ApplySeed(ctx, pool, os.DirFS(cfg.Seeds.Directory), "demo.sql"); err != nil {
		return fmt.Errorf("apply demo seed: %w", err)
	}

	logger.Info("fictional demo data is current")
	return nil
}
