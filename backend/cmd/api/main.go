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

	"github.com/jackc/pgx/v5/pgxpool"

	aiadapter "rigmark/internal/adapters/ai"
	authadapter "rigmark/internal/adapters/auth"
	adminpostgres "rigmark/internal/adapters/postgres/admin"
	analyticspostgres "rigmark/internal/adapters/postgres/analytics"
	catalogpostgres "rigmark/internal/adapters/postgres/catalog"
	commercepostgres "rigmark/internal/adapters/postgres/commerce"
	contentpostgres "rigmark/internal/adapters/postgres/content"
	evidencepostgres "rigmark/internal/adapters/postgres/evidence"
	identitypostgres "rigmark/internal/adapters/postgres/identity"
	planningpostgres "rigmark/internal/adapters/postgres/planning"
	recommendationpostgres "rigmark/internal/adapters/postgres/recommendation"
	"rigmark/internal/adapters/storage/localimages"
	admin "rigmark/internal/modules/admin/application"
	ai "rigmark/internal/modules/ai/application"
	analytics "rigmark/internal/modules/analytics/application"
	catalog "rigmark/internal/modules/catalog/application"
	commerce "rigmark/internal/modules/commerce/application"
	content "rigmark/internal/modules/content/application"
	evidence "rigmark/internal/modules/evidence/application"
	health "rigmark/internal/modules/health/application"
	identity "rigmark/internal/modules/identity/application"
	planning "rigmark/internal/modules/planning/application"
	recommendation "rigmark/internal/modules/recommendation/application"
	"rigmark/internal/platform/config"
	"rigmark/internal/transport/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("API stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	ctx := context.Background()
	databaseConfig, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("parse database configuration: %w", err)
	}
	databaseConfig.MaxConns = cfg.Database.MaxConnections
	databaseConfig.MinConns = cfg.Database.MinConnections
	databaseConfig.MaxConnLifetime = cfg.Database.MaxConnectionLifetime
	databaseConfig.MaxConnIdleTime = cfg.Database.MaxConnectionIdleTime
	databaseConfig.HealthCheckPeriod = cfg.Database.HealthCheckPeriod
	db, err := pgxpool.NewWithConfig(ctx, databaseConfig)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer db.Close()
	connectContext, cancelConnect := context.WithTimeout(ctx, cfg.Database.ConnectTimeout)
	defer cancelConnect()
	if err := db.Ping(connectContext); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	healthService := health.NewService(db, cfg.Version)
	passwordHasher, err := authadapter.NewPasswordHasher(authadapter.PasswordParams{
		MemoryKiB:   cfg.Passwords.MemoryKiB,
		Iterations:  cfg.Passwords.Iterations,
		Parallelism: cfg.Passwords.Parallelism,
		KeyLength:   cfg.Passwords.KeyLength,
	})
	if err != nil {
		return fmt.Errorf("configure password hashing: %w", err)
	}
	identityRepository := identitypostgres.New(db)
	catalogRepository := catalogpostgres.New(db)
	commerceRepository := commercepostgres.New(db, cfg.Commerce.OfferMaximumAge)
	contentRepository := contentpostgres.New(db)
	analyticsRepository := analyticspostgres.New(db)
	adminRepository := adminpostgres.New(db)
	evidenceRepository := evidencepostgres.New(db)
	imageStore, err := localimages.New(cfg.Assets.ProductImageDirectory)
	if err != nil {
		return fmt.Errorf("configure product image storage: %w", err)
	}
	wishlistRepository := planningpostgres.NewWishlistRepository(db)
	comparisonRepository := planningpostgres.NewComparisonRepository(db)
	recommendationRepository := recommendationpostgres.New(db)
	aiProvider, err := aiadapter.NewRegistry().Select(aiadapter.Config{
		Provider: cfg.AI.Provider, Model: cfg.AI.Model, APIKey: cfg.AI.APIKey,
		Timeout: cfg.AI.Timeout, MaxResponseBytes: cfg.AI.MaxResponseBytes,
	})
	if err != nil {
		return fmt.Errorf("configure AI provider: %w", err)
	}
	aiService, err := ai.NewService(aiProvider)
	if err != nil {
		return fmt.Errorf("configure AI service: %w", err)
	}
	contentService, err := content.NewService(contentRepository, catalogRepository, cfg.Site.PublicURL)
	if err != nil {
		return fmt.Errorf("configure editorial content service: %w", err)
	}
	authService, err := identity.NewService(
		identityRepository,
		passwordHasher,
		authadapter.NewSessionTokenManager(),
		cfg.Auth.SessionTTL,
		cfg.Auth.SessionIdleTTL,
	)
	if err != nil {
		return fmt.Errorf("configure authentication: %w", err)
	}
	server := &http.Server{
		Addr: ":" + cfg.HTTP.Port,
		Handler: httpapi.NewRouter(
			healthService,
			authService,
			httpapi.AuthCookieConfig{
				Name:          cfg.Auth.SessionCookieName,
				Secure:        cfg.Auth.CookieSecure,
				MaxAge:        int(cfg.Auth.SessionTTL.Seconds()),
				AllowedOrigin: cfg.Site.PublicURL,
			},
			logger,
			httpapi.PublicServices{
				Catalog:              catalog.NewService(catalogRepository),
				Commerce:             commerce.NewService(commerceRepository, commerceRepository),
				Wishlist:             planning.NewWishlistService(wishlistRepository),
				Recommendations:      recommendation.NewService(recommendationRepository, catalogRepository, recommendationRepository),
				Comparison:           planning.NewComparisonService(comparisonRepository),
				Analytics:            analytics.NewService(analyticsRepository),
				AnalyticsReporting:   analytics.NewReportingService(analyticsRepository),
				Admin:                admin.NewService(adminRepository, imageStore),
				AI:                   aiService,
				Content:              contentService,
				Evidence:             evidence.NewService(evidenceRepository),
				RecommendationPolicy: recommendation.NewPolicyService(recommendationRepository),
				RateLimits: httpapi.RateLimitConfig{
					AuthenticationPerMinute: cfg.RateLimits.AuthenticationPerMinute,
					RecommendationPerMinute: cfg.RateLimits.RecommendationPerMinute,
					AnalyticsPerMinute:      cfg.RateLimits.AnalyticsPerMinute,
					AffiliatePerMinute:      cfg.RateLimits.AffiliatePerMinute,
					MutationPerMinute:       cfg.RateLimits.MutationPerMinute,
				},
			},
		),
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	shutdownSignals, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API listening", "address", server.Addr, "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-shutdownSignals.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	return nil
}
