package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"rigmark/internal/platform/config"
	"rigmark/internal/platform/database"
	"rigmark/internal/platform/observability"
)

var errRefusedDeployedSeed = errors.New(
	"fictional demo data must not be loaded into a deployed environment")

// fixtureLoadAllowed reports whether the fictional fixture may be loaded into an
// environment. The fixture publishes invented products at the same status the
// public catalog serves, so a deployed environment would show invented prices to
// real visitors. Removing the fixture stays allowed everywhere.
func fixtureLoadAllowed(environment string) bool {
	return environment != "production" && environment != "staging"
}

func main() {
	purge := flag.Bool("purge", false,
		"remove the fictional demo fixture from the database instead of loading it")
	flag.Parse()

	logger := observability.NewJSONLogger(os.Stdout, slog.LevelInfo)
	if err := run(logger, *purge); err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, purge bool) error {
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
	// Each vertical has its own fictional fixture. Seeding every vertical would
	// leave a deployment holding catalogs it does not serve. The unprefixed
	// demo.sql was the home-gym fixture and is gone with that vertical.
	seedFile := cfg.Recommendation.Vertical + "_demo.sql"

	// Removing the fixture is always allowed: a deployment that already holds it
	// needs the cleanup most, and the purge only touches fixture rows.
	if purge {
		purgeFile := cfg.Recommendation.Vertical + "_demo_purge.sql"
		if err := database.ApplySeed(ctx, pool, os.DirFS(cfg.Seeds.Directory), purgeFile); err != nil {
			return fmt.Errorf("apply %s purge: %w", cfg.Recommendation.Vertical, err)
		}
		logger.Info("fictional demo data removed", "vertical", cfg.Recommendation.Vertical, "seed", purgeFile)
		return nil
	}

	// The fixture publishes invented products at the same status the public
	// catalog serves, so loading it into a deployed environment puts invented
	// prices in front of real visitors. Refuse rather than rely on discipline.
	if !fixtureLoadAllowed(cfg.Environment) {
		// The JSON logger reduces an unrecognised error to its Go type, so the
		// reason has to be logged here or the operator sees only *fmt.wrapError.
		logger.Error("refusing to load fictional demo data into a deployed environment",
			"environment", cfg.Environment, "seed", seedFile,
			"remedy", "run 'make seed-purge' to remove the fixture from this database")
		return errRefusedDeployedSeed
	}

	if err := database.ApplySeed(ctx, pool, os.DirFS(cfg.Seeds.Directory), seedFile); err != nil {
		return fmt.Errorf("apply %s seed: %w", cfg.Recommendation.Vertical, err)
	}

	logger.Info("fictional demo data is current", "vertical", cfg.Recommendation.Vertical, "seed", seedFile)
	return nil
}
