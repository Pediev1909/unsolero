package config

import (
	"net/netip"
	"testing"
	"time"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an error when DATABASE_URL is empty")
	}
}

func TestLoadParsesExplicitTrustedProxyCIDRs(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("TRUSTED_PROXY_CIDRS", "192.0.2.4/32, 2001:db8::/48")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}
	want := []netip.Prefix{netip.MustParsePrefix("192.0.2.4/32"), netip.MustParsePrefix("2001:db8::/48")}
	if len(cfg.HTTP.TrustedProxyCIDRs) != len(want) {
		t.Fatalf("trusted proxy prefixes = %#v", cfg.HTTP.TrustedProxyCIDRs)
	}
	for index := range want {
		if cfg.HTTP.TrustedProxyCIDRs[index] != want[index] {
			t.Fatalf("trusted proxy prefix[%d] = %s, want %s", index, cfg.HTTP.TrustedProxyCIDRs[index], want[index])
		}
	}
}

func TestLoadRejectsInvalidOrDangerouslyBroadTrustedProxyCIDRs(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	for _, value := range []string{"not-a-prefix", "0.0.0.0/0", "::/0"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("TRUSTED_PROXY_CIDRS", value)
			if _, err := Load(); err == nil {
				t.Fatalf("Load() accepted unsafe trusted proxy prefix %q", value)
			}
		})
	}
}

func TestLoadValidatesDistributedRateLimitConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("RATE_LIMIT_PROVIDER", "redis")
	t.Setenv("RATE_LIMIT_REDIS_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted Redis rate limiting without a URL")
	}
	t.Setenv("RATE_LIMIT_REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("API_REPLICA_COUNT", "3")
	if cfg, err := Load(); err != nil || cfg.RateLimits.Provider != "redis" {
		t.Fatalf("Load() Redis configuration = %#v, %v", cfg.RateLimits, err)
	}
}

func TestLoadValidatesS3CompatibleMediaConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("MEDIA_STORAGE_PROVIDER", "s3")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted incomplete S3-compatible configuration")
	}
	t.Setenv("MEDIA_S3_ENDPOINT", "minio:9000")
	t.Setenv("MEDIA_S3_ACCESS_KEY", "local-access")
	t.Setenv("MEDIA_S3_SECRET_KEY", "local-secret")
	t.Setenv("MEDIA_S3_BUCKET", "unsolero-media")
	if cfg, err := Load(); err != nil || cfg.Assets.StorageProvider != "s3" {
		t.Fatalf("Load() S3-compatible configuration = %#v, %v", cfg.Assets, err)
	}
}

func TestLoadRequiresEncryptedPhase10ProvidersInProduction(t *testing.T) {
	t.Run("Redis", func(t *testing.T) {
		setSecureProductionEnvironment(t)
		t.Setenv("RATE_LIMIT_PROVIDER", "redis")
		t.Setenv("RATE_LIMIT_REDIS_URL", "redis://redis.internal:6379/0")
		if _, err := Load(); err == nil {
			t.Fatal("Load() accepted plaintext production Redis transport")
		}
	})
	t.Run("S3", func(t *testing.T) {
		setSecureProductionEnvironment(t)
		t.Setenv("MEDIA_STORAGE_PROVIDER", "s3")
		t.Setenv("MEDIA_S3_ENDPOINT", "objects.internal:9000")
		t.Setenv("MEDIA_S3_ACCESS_KEY", "access")
		t.Setenv("MEDIA_S3_SECRET_KEY", "secret")
		t.Setenv("MEDIA_S3_BUCKET", "unsolero-products")
		t.Setenv("MEDIA_S3_SECURE", "false")
		if _, err := Load(); err == nil {
			t.Fatal("Load() accepted plaintext production object storage")
		}
	})
	t.Run("SMTP", func(t *testing.T) {
		setSecureProductionEnvironment(t)
		t.Setenv("EMAIL_PROVIDER", "smtp")
		t.Setenv("EMAIL_SENDER_ADDRESS", "security@unsolero.com")
		t.Setenv("EMAIL_SMTP_ADDRESS", "smtp.internal:587")
		t.Setenv("EMAIL_SMTP_REQUIRE_TLS", "false")
		if _, err := Load(); err == nil {
			t.Fatal("Load() accepted production SMTP without required TLS")
		}
	})
}

func TestLoadAcceptsEncryptedPhase10ProvidersInProduction(t *testing.T) {
	setSecureProductionEnvironment(t)
	t.Setenv("RATE_LIMIT_PROVIDER", "redis")
	t.Setenv("RATE_LIMIT_REDIS_URL", "rediss://service:secret@redis.internal:6379/0")
	t.Setenv("API_REPLICA_COUNT", "3")
	t.Setenv("MEDIA_STORAGE_PROVIDER", "s3")
	t.Setenv("MEDIA_S3_ENDPOINT", "objects.internal:443")
	t.Setenv("MEDIA_S3_ACCESS_KEY", "access")
	t.Setenv("MEDIA_S3_SECRET_KEY", "secret")
	t.Setenv("MEDIA_S3_BUCKET", "unsolero-products")
	t.Setenv("MEDIA_S3_SECURE", "true")
	t.Setenv("EMAIL_PROVIDER", "smtp")
	t.Setenv("EMAIL_SENDER_ADDRESS", "security@unsolero.com")
	t.Setenv("EMAIL_SMTP_ADDRESS", "smtp.internal:587")
	t.Setenv("EMAIL_SMTP_REQUIRE_TLS", "true")
	if _, err := Load(); err != nil {
		t.Fatalf("Load() rejected secure Phase 10 provider configuration: %v", err)
	}
}

func TestLoadUsesSafeDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_VERSION", "")
	t.Setenv("API_PORT", "")
	t.Setenv("MIGRATIONS_DIR", "")
	t.Setenv("SEEDS_DIR", "")
	t.Setenv("SESSION_COOKIE_SECURE", "")
	t.Setenv("SESSION_TTL", "")
	t.Setenv("SESSION_IDLE_TTL", "")
	t.Setenv("AI_PROVIDER", "")
	t.Setenv("AI_MODEL", "")
	t.Setenv("AI_API_KEY", "")
	t.Setenv("AI_TIMEOUT", "")
	t.Setenv("AI_MAX_RESPONSE_BYTES", "")
	t.Setenv("PUBLIC_SITE_URL", "")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "")
	t.Setenv("DATABASE_MIN_CONNECTIONS", "")
	t.Setenv("DATABASE_MAX_CONNECTION_LIFETIME", "")
	t.Setenv("DATABASE_MAX_CONNECTION_IDLE_TIME", "")
	t.Setenv("DATABASE_HEALTH_CHECK_PERIOD", "")
	t.Setenv("DATABASE_CONNECT_TIMEOUT", "")
	t.Setenv("RATE_LIMIT_AUTH_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_RECOMMENDATION_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_ANALYTICS_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_AFFILIATE_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_MUTATION_PER_MINUTE", "")
	t.Setenv("OFFER_MAXIMUM_AGE", "")
	t.Setenv("AFFILIATE_CLICK_RETENTION", "")
	t.Setenv("COMMERCE_WORKER_POLL_INTERVAL", "")
	t.Setenv("ANALYTICS_SUBJECT_COOKIE_NAME", "")
	t.Setenv("ANALYTICS_ANONYMOUS_RETENTION", "")
	t.Setenv("ANALYTICS_AUTHENTICATED_RETENTION", "")
	t.Setenv("ANALYTICS_RECEIPT_RETENTION", "")
	t.Setenv("ANALYTICS_CLEANUP_BATCH_SIZE", "")
	t.Setenv("MEDIA_STORAGE_PROVIDER", "")
	t.Setenv("MEDIA_SCAN_PROVIDER", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}

	if cfg.Environment != "development" {
		t.Errorf("Environment = %q, want development", cfg.Environment)
	}
	if cfg.Version != "development" {
		t.Errorf("Version = %q, want development", cfg.Version)
	}
	if cfg.HTTP.Port != "8080" {
		t.Errorf("HTTP.Port = %q, want 8080", cfg.HTTP.Port)
	}
	if cfg.Migrations.Directory != "./migrations" {
		t.Errorf("Migrations.Directory = %q, want ./migrations", cfg.Migrations.Directory)
	}
	if cfg.Seeds.Directory != "./seeds" {
		t.Errorf("Seeds.Directory = %q, want ./seeds", cfg.Seeds.Directory)
	}
	if cfg.Auth.SessionCookieName != "rigmark_session" || cfg.Auth.CookieSecure {
		t.Errorf("Auth defaults = %#v", cfg.Auth)
	}
	if cfg.Auth.SessionTTL != 30*24*time.Hour ||
		cfg.Auth.SessionIdleTTL != 7*24*time.Hour {
		t.Errorf("Auth session lifetimes = %#v", cfg.Auth)
	}
	if cfg.AI.Provider != "disabled" || cfg.AI.Timeout != 15*time.Second || cfg.AI.MaxResponseBytes != 65_536 {
		t.Errorf("AI defaults = %#v", cfg.AI)
	}
	if cfg.Site.PublicURL != "http://localhost:5173" {
		t.Errorf("Site.PublicURL = %q, want local frontend origin", cfg.Site.PublicURL)
	}
	if cfg.Assets.StorageProvider != "local" || cfg.Assets.ScanProvider != "development" {
		t.Errorf("Assets defaults = %#v", cfg.Assets)
	}
	if cfg.Database.MaxConnections != 20 || cfg.Database.MinConnections != 2 || cfg.Database.ConnectTimeout != 10*time.Second {
		t.Errorf("Database defaults = %#v", cfg.Database)
	}
	if cfg.RateLimits.AuthenticationPerMinute != 10 || cfg.RateLimits.MutationPerMinute != 240 {
		t.Errorf("Rate limit defaults = %#v", cfg.RateLimits)
	}
	if cfg.RateLimits.RegistrationPerMinute != 5 || cfg.RateLimits.PasswordResetPerMinute != 5 ||
		cfg.RateLimits.AdminPerMinute != 240 || cfg.RateLimits.Namespace != "unsolero:rate-limit" {
		t.Errorf("Rate limit policy defaults = %#v", cfg.RateLimits)
	}
	if cfg.Commerce.OfferMaximumAge != 72*time.Hour ||
		cfg.Commerce.AffiliateClickRetention != 397*24*time.Hour ||
		cfg.Commerce.WorkerPollInterval != 15*time.Second {
		t.Errorf("Commerce defaults = %#v", cfg.Commerce)
	}
	if cfg.Analytics.SubjectCookieName != "unsolero_analytics_subject" ||
		cfg.Analytics.AnonymousRetention != 90*24*time.Hour ||
		cfg.Analytics.AuthenticatedRetention != 397*24*time.Hour ||
		cfg.Analytics.ReceiptRetention != 30*24*time.Hour || cfg.Analytics.CleanupBatchSize != 1000 {
		t.Errorf("Analytics defaults = %#v", cfg.Analytics)
	}
}

func TestLoadRejectsUnsafeAnalyticsRetention(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("ANALYTICS_ANONYMOUS_RETENTION", "1h")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted unsafe anonymous retention")
	}
	t.Setenv("ANALYTICS_ANONYMOUS_RETENTION", "2160h")
	t.Setenv("ANALYTICS_CLEANUP_BATCH_SIZE", "10001")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an unbounded analytics cleanup batch")
	}
}

func TestLoadRejectsUnsafeCommerceOperationalWindows(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AFFILIATE_CLICK_RETENTION", "1h")
	if _, err := Load(); err == nil {
		t.Fatal("Load() expected unsafe affiliate retention to fail")
	}
	t.Setenv("AFFILIATE_CLICK_RETENTION", "9528h")
	t.Setenv("COMMERCE_WORKER_POLL_INTERVAL", "500ms")
	if _, err := Load(); err == nil {
		t.Fatal("Load() expected unsafe commerce poll interval to fail")
	}
}

func TestLoadRejectsInsecureProductionSiteURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("PUBLIC_SITE_URL", "http://rigmark.example")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected production HTTP public URL to fail")
	}
}

func TestLoadRequiresServerOnlyCredentialsForEnabledAI(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AI_PROVIDER", "openai")
	t.Setenv("AI_MODEL", "")
	t.Setenv("AI_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected enabled AI without credentials to fail")
	}
}

func TestLoadAcceptsConfiguredAIProvider(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AI_PROVIDER", "custom")
	t.Setenv("AI_MODEL", "model-v1")
	t.Setenv("AI_API_KEY", "server-secret")
	t.Setenv("AI_TIMEOUT", "5s")
	t.Setenv("AI_MAX_RESPONSE_BYTES", "32768")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AI.Provider != "custom" || cfg.AI.Model != "model-v1" || cfg.AI.APIKey != "server-secret" ||
		cfg.AI.Timeout != 5*time.Second || cfg.AI.MaxResponseBytes != 32768 {
		t.Fatalf("AI config = %#v", cfg.AI)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("API_PORT", "not-a-port")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an error for an invalid API_PORT")
	}
}

func TestLoadRequiresSecureCookiesInProduction(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_COOKIE_SECURE", "false")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected production to reject insecure session cookies")
	}
}

func TestLoadRequiresDatabaseTLSInProduction(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://rigmark:secret@database.example/rigmark?sslmode=disable")
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("PUBLIC_SITE_URL", "https://rigmark.example")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected production PostgreSQL without TLS to fail")
	}
}

func TestLoadAcceptsSecureProductionConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://rigmark:secret@database.example/rigmark?sslmode=verify-full")
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_VERSION", "2026.08.17")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("PUBLIC_SITE_URL", "https://unsolero.com")
	t.Setenv("EMAIL_PROVIDER", "external")
	t.Setenv("MEDIA_STORAGE_PROVIDER", "external")
	t.Setenv("MEDIA_SCAN_PROVIDER", "external")
	t.Setenv("ALERT_PROVIDER", "external")
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("METRICS_TOKEN", "12345678901234567890123456789012")
	t.Setenv("MFA_ENCRYPTION_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("RATE_LIMIT_KEY_SECRET", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	if _, err := Load(); err != nil {
		t.Fatalf("Load() rejected a secure production configuration: %v", err)
	}
}

func TestLoadAcceptsExplicitProductionShapedLocalStaging(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("APP_VERSION", "phase11-staging")
	t.Setenv("DATABASE_URL", "postgres://staging:staging@postgres:5432/staging?sslmode=disable")
	t.Setenv("PUBLIC_SITE_URL", "https://localhost:8443")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("API_REPLICA_COUNT", "2")
	t.Setenv("RATE_LIMIT_PROVIDER", "redis")
	t.Setenv("RATE_LIMIT_REDIS_URL", "redis://:staging@redis-staging:6379/0")
	t.Setenv("RATE_LIMIT_KEY_SECRET", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("MFA_ENCRYPTION_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("MEDIA_STORAGE_PROVIDER", "s3")
	t.Setenv("MEDIA_S3_ENDPOINT", "minio-staging:9000")
	t.Setenv("MEDIA_S3_ACCESS_KEY", "staging")
	t.Setenv("MEDIA_S3_SECRET_KEY", "staging-secret")
	t.Setenv("MEDIA_S3_BUCKET", "unsolero-staging")
	t.Setenv("MEDIA_S3_SECURE", "false")
	t.Setenv("MEDIA_SCAN_PROVIDER", "development")
	t.Setenv("EMAIL_PROVIDER", "disabled")
	t.Setenv("ALLOW_INSECURE_LOCAL_STAGING", "true")
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() rejected explicit local staging topology: %v", err)
	}
	if config.Environment != "staging" || !config.Auth.CookieSecure || config.RateLimits.Provider != "redis" || config.Assets.StorageProvider != "s3" {
		t.Fatalf("staging config=%+v", config)
	}
}

func TestLoadRejectsPrivateOrReservedProductionPublicURL(t *testing.T) {
	for _, siteURL := range []string{
		"https://localhost", "https://127.0.0.1", "https://10.10.0.1",
		"https://192.168.1.20", "https://service.internal", "https://unsolero.example",
	} {
		t.Run(siteURL, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://rigmark:secret@database.example/rigmark?sslmode=verify-full")
			t.Setenv("APP_ENV", "production")
			t.Setenv("APP_VERSION", "release-test")
			t.Setenv("SESSION_COOKIE_SECURE", "true")
			t.Setenv("PUBLIC_SITE_URL", siteURL)
			t.Setenv("EMAIL_PROVIDER", "external")
			t.Setenv("MFA_ENCRYPTION_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			t.Setenv("RATE_LIMIT_KEY_SECRET", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			t.Setenv("MEDIA_STORAGE_PROVIDER", "external")
			t.Setenv("MEDIA_SCAN_PROVIDER", "external")
			if _, err := Load(); err == nil {
				t.Fatalf("Load() accepted non-public production origin %q", siteURL)
			}
		})
	}
}

func TestLoadRejectsLocalOrUnscannedProductionMedia(t *testing.T) {
	setSecureProductionEnvironment(t)
	t.Setenv("MEDIA_STORAGE_PROVIDER", "local")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted local production media storage")
	}

	setSecureProductionEnvironment(t)
	t.Setenv("MEDIA_SCAN_PROVIDER", "disabled")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted production media without external scanning")
	}
}

func setSecureProductionEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://rigmark:secret@database.example/rigmark?sslmode=verify-full")
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_VERSION", "release-test")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("PUBLIC_SITE_URL", "https://unsolero.com")
	t.Setenv("EMAIL_PROVIDER", "external")
	t.Setenv("MFA_ENCRYPTION_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("RATE_LIMIT_KEY_SECRET", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("MEDIA_STORAGE_PROVIDER", "external")
	t.Setenv("MEDIA_SCAN_PROVIDER", "external")
	t.Setenv("ALERT_PROVIDER", "external")
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("METRICS_TOKEN", "12345678901234567890123456789012")
}

func TestLoadRequiresOperationalVisibilityInProduction(t *testing.T) {
	setSecureProductionEnvironment(t)
	t.Setenv("ALERT_PROVIDER", "disabled")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted production without operational alert delivery")
	}

	setSecureProductionEnvironment(t)
	t.Setenv("METRICS_ENABLED", "false")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted production without operational metrics")
	}
}

func TestLoadRejectsLoopbackProductionDependencies(t *testing.T) {
	setSecureProductionEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://user:secret@localhost/unsolero?sslmode=verify-full")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a loopback production database")
	}

	setSecureProductionEnvironment(t)
	t.Setenv("RATE_LIMIT_PROVIDER", "redis")
	t.Setenv("RATE_LIMIT_REDIS_URL", "rediss://user:secret@127.0.0.1:6379/0")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a loopback production Redis service")
	}
}

func TestLoadValidatesWebhookAlertTransport(t *testing.T) {
	setSecureProductionEnvironment(t)
	t.Setenv("ALERT_PROVIDER", "webhook")
	t.Setenv("ALERT_WEBHOOK_URL", "http://alerts.example.test/unsolero")
	t.Setenv("ALERT_WEBHOOK_TOKEN", "12345678901234567890123456789012")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted plaintext production alert delivery")
	}

	setSecureProductionEnvironment(t)
	t.Setenv("ALERT_PROVIDER", "webhook")
	t.Setenv("ALERT_WEBHOOK_URL", "https://alerts.example.test/unsolero")
	t.Setenv("ALERT_WEBHOOK_TOKEN", "12345678901234567890123456789012")
	if _, err := Load(); err != nil {
		t.Fatalf("Load() rejected secure webhook alert delivery: %v", err)
	}
}

func TestLoadConfinesInsecureStagingOverrideToLoopback(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("APP_VERSION", "local-staging-test")
	t.Setenv("ALLOW_INSECURE_LOCAL_STAGING", "true")
	t.Setenv("DATABASE_URL", "postgres://staging:secret@postgres/staging?sslmode=disable")
	t.Setenv("PUBLIC_SITE_URL", "https://staging.example.test")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("RATE_LIMIT_KEY_SECRET", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("MFA_ENCRYPTION_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	t.Setenv("MEDIA_STORAGE_PROVIDER", "external")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted the insecure staging override on a non-loopback origin")
	}
}

func TestLoadRejectsMultipleReplicasWithLocalRateLimiter(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("API_REPLICA_COUNT", "2")
	t.Setenv("RATE_LIMIT_PROVIDER", "local")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted multiple replicas with process-local abuse protection")
	}
}

func TestLoadRequiresMetricsCredentialWhenEnabled(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("METRICS_TOKEN", "short")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an unsafe metrics credential")
	}
}

func TestLoadRejectsUnsafeHTTPAndDatabaseTimeouts(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("HTTP_HANDLER_TIMEOUT", "30s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "20s")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a write timeout shorter than the handler deadline")
	}
	t.Setenv("HTTP_HANDLER_TIMEOUT", "10s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "20s")
	t.Setenv("DATABASE_STATEMENT_TIMEOUT", "500ms")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an unsafe database statement timeout")
	}
}

func TestLoadRejectsDevelopmentDeliveryAndMissingMFAKeyInProduction(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://rigmark:secret@database.example/rigmark?sslmode=verify-full")
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("PUBLIC_SITE_URL", "https://unsolero.example")
	t.Setenv("EMAIL_PROVIDER", "development")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted the development email sink in production")
	}
	t.Setenv("EMAIL_PROVIDER", "external")
	t.Setenv("MFA_ENCRYPTION_KEY", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted production without an MFA encryption key")
	}
}

func TestLoadRejectsInvalidOperationalLimits(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "1")
	t.Setenv("DATABASE_MIN_CONNECTIONS", "2")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an invalid database pool to fail")
	}

	t.Setenv("DATABASE_MIN_CONNECTIONS", "0")
	t.Setenv("RATE_LIMIT_AUTH_PER_MINUTE", "0")
	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an invalid rate limit to fail")
	}
}

func TestLoadRejectsIdleSessionLongerThanAbsoluteSession(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SESSION_TTL", "1h")
	t.Setenv("SESSION_IDLE_TTL", "2h")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected invalid session lifetimes to fail")
	}
}

func TestLoadRejectsInvalidSessionCookieName(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SESSION_COOKIE_NAME", "invalid cookie")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an invalid session cookie name to fail")
	}
}

func TestLoadRejectsUnsafeOfferFreshnessWindow(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("OFFER_MAXIMUM_AGE", "721h")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an excessive offer freshness window to fail")
	}
}
