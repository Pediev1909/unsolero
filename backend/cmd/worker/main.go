package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	merchantadapter "rigmark/internal/adapters/merchant"
	adminpostgres "rigmark/internal/adapters/postgres/admin"
	analyticspostgres "rigmark/internal/adapters/postgres/analytics"
	commercepostgres "rigmark/internal/adapters/postgres/commerce"
	identitypostgres "rigmark/internal/adapters/postgres/identity"
	app "rigmark/internal/app"
	admin "rigmark/internal/modules/admin/application"
	analytics "rigmark/internal/modules/analytics/application"
	analyticsdomain "rigmark/internal/modules/analytics/domain"
	commerce "rigmark/internal/modules/commerce/application"
	"rigmark/internal/platform/alerting"
	"rigmark/internal/platform/config"
	"rigmark/internal/platform/database"
	"rigmark/internal/platform/observability"
)

func main() {
	logger := observability.NewJSONLogger(os.Stdout, slog.LevelInfo)
	if err := run(logger); err != nil {
		logger.Error("commerce worker stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, err := database.OpenPool(ctx, database.PoolConfig{
		URL: cfg.Database.URL, ApplicationName: "unsolero-worker", MaxConnections: cfg.Database.MaxConnections,
		MinConnections: cfg.Database.MinConnections, MaxConnectionLifetime: cfg.Database.MaxConnectionLifetime,
		MaxConnectionIdleTime: cfg.Database.MaxConnectionIdleTime, HealthCheckPeriod: cfg.Database.HealthCheckPeriod,
		ConnectTimeout: cfg.Database.ConnectTimeout, StatementTimeout: cfg.Database.StatementTimeout,
		LockTimeout: cfg.Database.LockTimeout, IdleTransactionTimeout: cfg.Database.IdleTransactionTimeout,
	})
	if err != nil {
		return err
	}
	defer db.Close()
	notifier, err := alerting.Select(alerting.Config{Provider: cfg.Operations.AlertProvider,
		Endpoint: cfg.Operations.AlertWebhookURL, Token: cfg.Operations.AlertWebhookToken, Timeout: cfg.Operations.AlertTimeout})
	if err != nil {
		return fmt.Errorf("configure alerting: %w", err)
	}
	repository := commercepostgres.New(db, cfg.Commerce.OfferMaximumAge)
	// No live provider is registered without a reviewed adapter and externally
	// supplied credentials. Unknown adapters resolve to the fail-closed adapter.
	service := commerce.NewImportService(repository, merchantadapter.NewRegistry())
	conversionService := commerce.NewConversionService(repository, merchantadapter.NewConversionRegistry())
	identityRepository := identitypostgres.New(db)
	analyticsService := analytics.NewServiceWithConfig(analyticspostgres.New(db), analytics.Config{
		AnonymousRetention:     cfg.Analytics.AnonymousRetention,
		AuthenticatedRetention: cfg.Analytics.AuthenticatedRetention,
		ReceiptRetention:       cfg.Analytics.ReceiptRetention,
	})
	imageStore, err := app.NewImageStorage(cfg.Assets)
	if err != nil {
		return err
	}
	mediaCleanup := admin.NewMediaCleanupService(adminpostgres.New(db), imageStore)
	logger.Info("commerce worker started", "poll_interval", cfg.Commerce.WorkerPollInterval.String())
	if err := runWorkerLoop(ctx, cfg.Commerce.WorkerPollInterval, cfg.Commerce.WorkerCycleTimeout,
		cfg.Commerce.WorkerFailureThreshold, notifier, logger, func(cycleCtx context.Context) error {
			cycleErr := workCycle(cycleCtx, service, conversionService, identityRepository, analyticsService, mediaCleanup,
				cfg.Analytics.CleanupBatchSize, cfg.Commerce.WorkerLeaseTimeout,
				cfg.Commerce.WorkerMaxItemsPerCycle, logger)
			checkpointCtx, cancel := context.WithTimeout(context.WithoutCancel(cycleCtx), cfg.Database.ConnectTimeout)
			defer cancel()
			if checkpointErr := observability.RecordWorkerCheckpoint(checkpointCtx, db, cycleErr == nil); checkpointErr != nil && cycleErr == nil {
				return fmt.Errorf("record worker checkpoint: %w", checkpointErr)
			}
			return cycleErr
		}); err != nil {
		return err
	}
	logger.Info("commerce worker shutdown completed")
	return nil
}

type securityCleaner interface {
	CleanupExpiredSecurityArtifacts(context.Context, time.Time) error
}

type analyticsCleaner interface {
	Cleanup(context.Context, int) (analyticsdomain.CleanupResult, error)
}

type mediaCleaner interface {
	Process(context.Context, int) (int, error)
}

func runWorkerLoop(ctx context.Context, pollInterval, cycleTimeout time.Duration, alertThreshold int, notifier alerting.Notifier, logger *slog.Logger, cycle func(context.Context) error) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	consecutiveFailures := 0
	for {
		cycleCtx, cancel := context.WithTimeout(ctx, cycleTimeout)
		err := cycle(cycleCtx)
		cancel()
		if err != nil && !errors.Is(err, context.Canceled) {
			consecutiveFailures++
			logger.Error("commerce worker cycle failed", "error", err, "consecutive_failures", consecutiveFailures)
			if consecutiveFailures == alertThreshold {
				alertErr := notifier.Notify(context.WithoutCancel(ctx), alerting.Alert{Category: alerting.WorkerRepeatedFailure,
					Summary: "The commerce worker has failed repeatedly.", OccurredAt: time.Now().UTC(), Severity: "critical"})
				if alertErr != nil {
					logger.Warn("operational alert was not delivered", "category", alerting.WorkerRepeatedFailure, "error", alertErr)
				}
			}
		} else if err == nil {
			consecutiveFailures = 0
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func workCycle(ctx context.Context, service *commerce.ImportService, conversions *commerce.ConversionService, security securityCleaner, analyticsCleanup analyticsCleaner, mediaCleanup mediaCleaner, analyticsBatch int, lease time.Duration, maximumItems int, logger *slog.Logger) error {
	recovered, err := service.RecoverStalled(ctx, lease, maximumItems)
	if err != nil {
		return err
	}
	queued, err := service.QueueScheduled(ctx, maximumItems)
	if err != nil {
		return err
	}
	processed := 0
	for processed < maximumItems {
		didWork, processErr := service.ProcessNext(ctx)
		if processErr != nil {
			return processErr
		}
		if !didWork {
			break
		}
		processed++
	}
	anonymized, err := service.AnonymizeExpiredClicks(ctx, 1000)
	if err != nil {
		return err
	}
	conversionRecovered, err := conversions.RecoverStalled(ctx, lease, maximumItems)
	if err != nil {
		return err
	}
	conversionQueued, err := conversions.QueueScheduled(ctx, maximumItems)
	if err != nil {
		return err
	}
	conversionProcessed := 0
	for conversionProcessed < maximumItems {
		didWork, processErr := conversions.ProcessNext(ctx)
		if processErr != nil {
			return processErr
		}
		if !didWork {
			break
		}
		conversionProcessed++
	}
	if err := security.CleanupExpiredSecurityArtifacts(ctx, time.Now().UTC()); err != nil {
		return err
	}
	analyticsResult, err := analyticsCleanup.Cleanup(ctx, analyticsBatch)
	if err != nil {
		return err
	}
	mediaDeleted, err := mediaCleanup.Process(ctx, maximumItems)
	if err != nil {
		return err
	}
	if recovered > 0 || queued > 0 || processed > 0 || anonymized > 0 ||
		conversionRecovered > 0 || conversionQueued > 0 || conversionProcessed > 0 ||
		analyticsResult.EventsDeleted > 0 || analyticsResult.ReceiptsDeleted > 0 || mediaDeleted > 0 {
		logger.Info("commerce worker cycle completed", "recovered", recovered, "queued", queued, "processed", processed,
			"clicks_anonymized", anonymized, "conversion_recovered", conversionRecovered,
			"conversion_queued", conversionQueued, "conversion_processed", conversionProcessed,
			"analytics_events_deleted", analyticsResult.EventsDeleted,
			"analytics_receipts_deleted", analyticsResult.ReceiptsDeleted)
		logger.Info("media cleanup completed", "objects_deleted", mediaDeleted)
	}
	return nil
}
