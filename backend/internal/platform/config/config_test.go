package config

import (
	"testing"
	"time"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an error when DATABASE_URL is empty")
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
	if cfg.Database.MaxConnections != 20 || cfg.Database.MinConnections != 2 || cfg.Database.ConnectTimeout != 10*time.Second {
		t.Errorf("Database defaults = %#v", cfg.Database)
	}
	if cfg.RateLimits.AuthenticationPerMinute != 10 || cfg.RateLimits.MutationPerMinute != 240 {
		t.Errorf("Rate limit defaults = %#v", cfg.RateLimits)
	}
	if cfg.Commerce.OfferMaximumAge != 72*time.Hour {
		t.Errorf("Commerce defaults = %#v", cfg.Commerce)
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
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("PUBLIC_SITE_URL", "https://rigmark.example")

	if _, err := Load(); err != nil {
		t.Fatalf("Load() rejected a secure production configuration: %v", err)
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
