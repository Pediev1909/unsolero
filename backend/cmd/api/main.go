package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	aiadapter "rigmark/internal/adapters/ai"
	authadapter "rigmark/internal/adapters/auth"
	emailadapter "rigmark/internal/adapters/email"
	merchantadapter "rigmark/internal/adapters/merchant"
	adminpostgres "rigmark/internal/adapters/postgres/admin"
	analyticspostgres "rigmark/internal/adapters/postgres/analytics"
	catalogpostgres "rigmark/internal/adapters/postgres/catalog"
	commercepostgres "rigmark/internal/adapters/postgres/commerce"
	contentpostgres "rigmark/internal/adapters/postgres/content"
	evidencepostgres "rigmark/internal/adapters/postgres/evidence"
	identitypostgres "rigmark/internal/adapters/postgres/identity"
	planningpostgres "rigmark/internal/adapters/postgres/planning"
	recommendationpostgres "rigmark/internal/adapters/postgres/recommendation"
	"rigmark/internal/adapters/storage/imagescan"
	app "rigmark/internal/app"
	admin "rigmark/internal/modules/admin/application"
	adminports "rigmark/internal/modules/admin/ports"
	ai "rigmark/internal/modules/ai/application"
	analytics "rigmark/internal/modules/analytics/application"
	catalog "rigmark/internal/modules/catalog/application"
	commerce "rigmark/internal/modules/commerce/application"
	content "rigmark/internal/modules/content/application"
	evidence "rigmark/internal/modules/evidence/application"
	health "rigmark/internal/modules/health/application"
	identity "rigmark/internal/modules/identity/application"
	identityports "rigmark/internal/modules/identity/ports"
	planning "rigmark/internal/modules/planning/application"
	recommendation "rigmark/internal/modules/recommendation/application"
	"rigmark/internal/platform/abuse"
	"rigmark/internal/platform/alerting"
	"rigmark/internal/platform/config"
	"rigmark/internal/platform/database"
	"rigmark/internal/platform/observability"
	"rigmark/internal/transport/httpapi"
	"rigmark/migrations"
)

func main() {
	logger := observability.NewJSONLogger(os.Stdout, slog.LevelInfo)
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
	db, err := database.OpenPool(ctx, database.PoolConfig{
		URL: cfg.Database.URL, ApplicationName: "unsolero-api", MaxConnections: cfg.Database.MaxConnections,
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
	limiter, err := selectRateLimiter(cfg.RateLimits)
	if err != nil {
		return fmt.Errorf("configure abuse protection: %w", err)
	}
	if closer, ok := limiter.(interface{ Close() error }); ok {
		defer closer.Close()
	}
	rateKey, err := operationalSecret(cfg.RateLimits.KeySecret, cfg.Environment)
	if err != nil {
		return fmt.Errorf("configure rate-limit key: %w", err)
	}
	imageStore, err := app.NewImageStorage(cfg.Assets)
	if err != nil {
		return err
	}
	var metrics observability.Recorder = observability.DisabledRecorder{}
	if cfg.Operations.MetricsEnabled {
		metrics = observability.NewMemoryRecorder()
	}

	healthService := health.NewServiceWithDependencies(db, cfg.Version, []health.Dependency{
		{Name: "schema", Critical: true, Checker: database.NewSchemaChecker(db, migrations.Files)},
		{Name: "rate_limit", Critical: true, Checker: limiter},
		{Name: "media_storage", Critical: true, Checker: imageStore},
		{Name: "alerting", Critical: false, Checker: notifier},
	})
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
	tokenManager := authadapter.NewSessionTokenManager()
	mfaKey, err := mfaEncryptionKey(cfg.Security.MFAEncryptionKey, cfg.Environment)
	if err != nil {
		return fmt.Errorf("configure MFA encryption: %w", err)
	}
	secretBox, err := authadapter.NewAESGCMSecretBox(mfaKey)
	if err != nil {
		return fmt.Errorf("configure MFA secret storage: %w", err)
	}
	var emailDelivery identityports.EmailDelivery
	var developmentEmail httpapi.DevelopmentEmailMessages
	switch cfg.Security.EmailProvider {
	case "development":
		sink := emailadapter.NewDevelopmentSink()
		emailDelivery = sink
		developmentEmail = sink
	case "disabled":
		emailDelivery = emailadapter.Disabled{}
	case "smtp":
		emailDelivery, err = emailadapter.NewSMTPDelivery(emailadapter.SMTPConfig{
			Address: cfg.Security.EmailSMTPAddress, Username: cfg.Security.EmailSMTPUsername,
			Password: cfg.Security.EmailSMTPPassword, SenderName: cfg.Security.EmailSenderName,
			SenderAddress: cfg.Security.EmailSenderAddress, PublicSiteURL: cfg.Site.PublicURL,
			RequireTLS: cfg.Security.EmailSMTPRequireTLS, Timeout: cfg.Security.EmailSMTPTimeout,
		})
		if err != nil {
			return fmt.Errorf("configure SMTP email delivery: %w", err)
		}
	case "external":
		return errors.New("EMAIL_PROVIDER=external requires a reviewed delivery adapter that is not linked in this repository")
	default:
		return errors.New("unsupported email delivery provider")
	}
	securityService, err := identity.NewSecurityService(identityRepository, passwordHasher, tokenManager,
		emailDelivery, secretBox, authadapter.TOTP{}, identity.SecurityConfig{
			VerificationTTL: cfg.Security.VerificationTTL, PasswordResetTTL: cfg.Security.PasswordResetTTL,
			MFAChallengeTTL: cfg.Security.MFAChallengeTTL, StepUpTTL: cfg.Security.MFAStepUpTTL,
			SessionTTL: cfg.Auth.SessionTTL, SessionIdleTTL: cfg.Auth.SessionIdleTTL, Issuer: "UNSOLERO",
		})
	if err != nil {
		return fmt.Errorf("configure account security: %w", err)
	}
	catalogRepository := catalogpostgres.NewForVertical(db, cfg.Recommendation.Vertical)
	commerceRepository := commercepostgres.New(db, cfg.Commerce.OfferMaximumAge)
	commerceProviders := merchantadapter.NewRegistry()
	commerceImportService := commerce.NewImportService(commerceRepository, commerceProviders)
	conversionService := commerce.NewConversionService(commerceRepository, merchantadapter.NewConversionRegistry())
	contentRepository := contentpostgres.New(db)
	analyticsRepository := analyticspostgres.New(db)
	adminRepository := adminpostgres.New(db)
	evidenceRepository := evidencepostgres.New(db)
	var imageScanner adminports.ImageScanner
	switch cfg.Assets.ScanProvider {
	case "development":
		imageScanner = imagescan.Development{}
	case "disabled":
		imageScanner = imagescan.Unavailable{}
	case "external":
		// clamd is reachable only on the private network. Readiness is checked
		// at startup so a misconfigured scanner is found now rather than by the
		// first administrator who tries to upload an image.
		clamav := imagescan.NewClamAV(cfg.Assets.ScanEndpoint, cfg.Assets.ScanTimeout, 0)
		if scanErr := clamav.Ready(ctx); scanErr != nil {
			return fmt.Errorf("media scanner is not ready: %w", scanErr)
		}
		imageScanner = clamav
	default:
		return errors.New("unsupported media scanning provider")
	}
	wishlistRepository := planningpostgres.NewWishlistRepository(db)
	comparisonRepository := planningpostgres.NewComparisonRepository(db)
	recommendationRepository := recommendationpostgres.NewForVertical(db, cfg.Recommendation.Vertical)
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
	authService, err := identity.NewServiceWithMFA(
		identityRepository,
		passwordHasher,
		tokenManager,
		securityService,
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
				Name:                   cfg.Auth.SessionCookieName,
				Secure:                 cfg.Auth.CookieSecure,
				MaxAge:                 int(cfg.Auth.SessionTTL.Seconds()),
				AllowedOrigin:          cfg.Site.PublicURL,
				AnalyticsSubjectName:   cfg.Analytics.SubjectCookieName,
				AnalyticsSubjectMaxAge: int(cfg.Analytics.AuthenticatedRetention.Seconds()),
			},
			logger,
			httpapi.PublicServices{
				Catalog:              catalog.NewService(catalogRepository),
				Commerce:             commerce.NewService(commerceRepository, commerceRepository, cfg.Commerce.AffiliateClickRetention),
				CommerceOperations:   commerceImportService,
				ConversionOperations: conversionService,
				Wishlist:             planning.NewWishlistService(wishlistRepository),
				Recommendations:      recommendation.NewService(recommendationRepository, catalogRepository, recommendationRepository),
				Comparison:           planning.NewComparisonService(comparisonRepository),
				Analytics: analytics.NewServiceWithConfig(analyticsRepository, analytics.Config{
					AnonymousRetention:     cfg.Analytics.AnonymousRetention,
					AuthenticatedRetention: cfg.Analytics.AuthenticatedRetention,
					ReceiptRetention:       cfg.Analytics.ReceiptRetention,
				}),
				AnalyticsReporting:   analytics.NewReportingService(analyticsRepository),
				Admin:                admin.NewServiceWithMedia(adminRepository, imageStore, imageScanner),
				AI:                   aiService,
				Content:              contentService,
				Evidence:             evidence.NewService(evidenceRepository),
				RecommendationPolicy: recommendation.NewPolicyService(recommendationRepository),
				RateLimits: httpapi.RateLimitConfig{
					AuthenticationPerMinute:  cfg.RateLimits.AuthenticationPerMinute,
					RegistrationPerMinute:    cfg.RateLimits.RegistrationPerMinute,
					PasswordResetPerMinute:   cfg.RateLimits.PasswordResetPerMinute,
					RecommendationPerMinute:  cfg.RateLimits.RecommendationPerMinute,
					AnalyticsPerMinute:       cfg.RateLimits.AnalyticsPerMinute,
					AffiliatePerMinute:       cfg.RateLimits.AffiliatePerMinute,
					AdminPerMinute:           cfg.RateLimits.AdminPerMinute,
					MutationPerMinute:        cfg.RateLimits.MutationPerMinute,
					RouteResolutionPerMinute: cfg.RateLimits.RouteResolutionPerMinute,
					TrustedProxyCIDRs:        cfg.HTTP.TrustedProxyCIDRs,
				},
				RateLimiter: limiter, RateLimitKey: rateKey, Metrics: metrics,
				MetricsToken: cfg.Operations.MetricsToken, Alerts: notifier,
				OperationalMetrics: observability.NewPostgresSource(db),
				HandlerTimeout:     cfg.HTTP.HandlerTimeout,
				Security:           securityService,
				DevelopmentEmail:   developmentEmail,
				SecurityPolicy: httpapi.SecurityPolicyConfig{EnforcePrivilegedMFA: cfg.Security.EnforcePrivilegedMFA,
					DevelopmentEmailAccess: cfg.Environment == "development" && cfg.Security.EmailProvider == "development"},
			},
		),
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
	}

	shutdownSignals, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("listen HTTP: %w", err)
	}
	logger.Info("API listening", "address", server.Addr, "environment", cfg.Environment, "version", cfg.Version)
	if err := serveHTTP(shutdownSignals, server, listener, cfg.HTTP.ShutdownTimeout); err != nil {
		return err
	}
	logger.Info("API shutdown completed")
	return nil
}

func serveHTTP(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(listener) }()
	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	return nil
}

func selectRateLimiter(config config.RateLimits) (abuse.Limiter, error) {
	if config.Provider == "local" {
		return abuse.NewLocalLimiter(), nil
	}
	if config.Provider == "redis" {
		return abuse.NewRedisLimiterFromURL(config.RedisURL, config.Namespace)
	}
	if config.Provider == "external" {
		return nil, errors.New("RATE_LIMIT_PROVIDER=external requires a reviewed distributed adapter that is not linked in this repository")
	}
	return nil, errors.New("unsupported rate-limit provider")
}

func operationalSecret(encoded, environment string) ([]byte, error) {
	if encoded != "" {
		value, err := base64.RawStdEncoding.Strict().DecodeString(encoded)
		if err != nil || len(value) != 32 {
			return nil, errors.New("secret must be raw-base64 for exactly 32 bytes")
		}
		return value, nil
	}
	if environment == "production" {
		return nil, errors.New("secret is required")
	}
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return nil, errors.New("generate development secret")
	}
	return value, nil
}

func mfaEncryptionKey(encoded, environment string) ([]byte, error) {
	if encoded != "" {
		value, err := base64.RawStdEncoding.Strict().DecodeString(encoded)
		if err != nil || len(value) != 32 {
			return nil, errors.New("MFA_ENCRYPTION_KEY must be raw-base64 for exactly 32 bytes")
		}
		return value, nil
	}
	if environment == "production" {
		return nil, errors.New("MFA_ENCRYPTION_KEY is required")
	}
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return nil, fmt.Errorf("generate development MFA key: %w", err)
	}
	return value, nil
}
