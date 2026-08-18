package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	adminpostgres "rigmark/internal/adapters/postgres/admin"
	"rigmark/internal/app"
	admin "rigmark/internal/modules/admin/application"
	"rigmark/internal/platform/config"
	"rigmark/internal/platform/database"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var apply bool
	var batchSize int
	var objectCursor, referenceCursor string
	var orphanGrace, deletionLease time.Duration
	flag.BoolVar(&apply, "apply-safe-orphans", false, "enqueue deletion only for validated, aged, unreferenced objects")
	flag.IntVar(&batchSize, "batch-size", 100, "objects and references to inspect (1-500)")
	flag.StringVar(&objectCursor, "object-cursor", "", "opaque storage cursor returned by the prior page")
	flag.StringVar(&referenceCursor, "reference-cursor", "", "database reference cursor returned by the prior page")
	flag.DurationVar(&orphanGrace, "orphan-grace", 24*time.Hour, "minimum object age before apply mode may enqueue deletion")
	flag.DurationVar(&deletionLease, "deletion-lease", 10*time.Minute, "age after which an incomplete deletion is stale")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	pool, err := database.OpenPool(ctx, database.PoolConfig{
		URL: cfg.Database.URL, ApplicationName: "unsolero-media-reconcile",
		MaxConnections: cfg.Database.MaxConnections, MinConnections: cfg.Database.MinConnections,
		MaxConnectionLifetime: cfg.Database.MaxConnectionLifetime, MaxConnectionIdleTime: cfg.Database.MaxConnectionIdleTime,
		HealthCheckPeriod: cfg.Database.HealthCheckPeriod, ConnectTimeout: cfg.Database.ConnectTimeout,
		StatementTimeout: cfg.Database.StatementTimeout, LockTimeout: cfg.Database.LockTimeout,
		IdleTransactionTimeout: cfg.Database.IdleTransactionTimeout,
	})
	if err != nil {
		return err
	}
	defer pool.Close()
	storage, err := app.NewImageStorage(cfg.Assets)
	if err != nil {
		return err
	}
	service, err := admin.NewMediaReconciliationService(adminpostgres.New(pool), storage)
	if err != nil {
		return err
	}
	mode := admin.MediaReconciliationDryRun
	if apply {
		mode = admin.MediaReconciliationApply
	}
	result, err := service.Reconcile(ctx, admin.MediaReconciliationRequest{
		Mode: mode, BatchSize: batchSize, ObjectCursor: objectCursor, ReferenceCursor: referenceCursor,
		OrphanGrace: orphanGrace, DeletionLease: deletionLease,
	})
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
