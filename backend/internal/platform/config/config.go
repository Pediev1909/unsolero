package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var cookieNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
var providerNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`)

const (
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 10 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
)

type Config struct {
	Environment string
	Version     string
	HTTP        HTTP
	Database    Database
	RateLimits  RateLimits
	Migrations  Migrations
	Seeds       Seeds
	Auth        Auth
	Passwords   Passwords
	Assets      Assets
	Site        Site
	AI          AI
	Commerce    Commerce
}

type Commerce struct {
	OfferMaximumAge time.Duration
}

type Site struct {
	PublicURL string
}

type AI struct {
	Provider         string
	Model            string
	APIKey           string
	Timeout          time.Duration
	MaxResponseBytes int64
}

type Assets struct {
	ProductImageDirectory string
}

type Auth struct {
	SessionCookieName string
	SessionTTL        time.Duration
	SessionIdleTTL    time.Duration
	CookieSecure      bool
}

type Passwords struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
}

type HTTP struct {
	Port              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

type Database struct {
	URL                   string
	MaxConnections        int32
	MinConnections        int32
	MaxConnectionLifetime time.Duration
	MaxConnectionIdleTime time.Duration
	HealthCheckPeriod     time.Duration
	ConnectTimeout        time.Duration
}

type RateLimits struct {
	AuthenticationPerMinute int
	RecommendationPerMinute int
	AnalyticsPerMinute      int
	AffiliatePerMinute      int
	MutationPerMinute       int
}

type Migrations struct {
	Directory string
}

type Seeds struct {
	Directory string
}

func Load() (Config, error) {
	environment := valueOrDefault("APP_ENV", "development")
	cookieSecure, err := boolValue("SESSION_COOKIE_SECURE", environment != "development")
	if err != nil {
		return Config{}, err
	}
	sessionTTL, err := durationValue("SESSION_TTL", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	sessionIdleTTL, err := durationValue("SESSION_IDLE_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	aiTimeout, err := durationValue("AI_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	aiMaxResponseBytes, err := int64Value("AI_MAX_RESPONSE_BYTES", 65_536)
	if err != nil {
		return Config{}, err
	}
	databaseMaxConnections, err := intValue("DATABASE_MAX_CONNECTIONS", 20)
	if err != nil {
		return Config{}, err
	}
	databaseMinConnections, err := intValue("DATABASE_MIN_CONNECTIONS", 2)
	if err != nil {
		return Config{}, err
	}
	databaseMaxConnectionLifetime, err := durationValue("DATABASE_MAX_CONNECTION_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}
	databaseMaxConnectionIdleTime, err := durationValue("DATABASE_MAX_CONNECTION_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	databaseHealthCheckPeriod, err := durationValue("DATABASE_HEALTH_CHECK_PERIOD", time.Minute)
	if err != nil {
		return Config{}, err
	}
	databaseConnectTimeout, err := durationValue("DATABASE_CONNECT_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	authenticationRateLimit, err := intValue("RATE_LIMIT_AUTH_PER_MINUTE", 10)
	if err != nil {
		return Config{}, err
	}
	recommendationRateLimit, err := intValue("RATE_LIMIT_RECOMMENDATION_PER_MINUTE", 20)
	if err != nil {
		return Config{}, err
	}
	analyticsRateLimit, err := intValue("RATE_LIMIT_ANALYTICS_PER_MINUTE", 120)
	if err != nil {
		return Config{}, err
	}
	affiliateRateLimit, err := intValue("RATE_LIMIT_AFFILIATE_PER_MINUTE", 120)
	if err != nil {
		return Config{}, err
	}
	mutationRateLimit, err := intValue("RATE_LIMIT_MUTATION_PER_MINUTE", 240)
	if err != nil {
		return Config{}, err
	}
	offerMaximumAge, err := durationValue("OFFER_MAXIMUM_AGE", 72*time.Hour)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment: environment,
		Version:     valueOrDefault("APP_VERSION", "development"),
		HTTP: HTTP{
			Port:              valueOrDefault("API_PORT", "8080"),
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			ReadTimeout:       defaultReadTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
			ShutdownTimeout:   defaultShutdownTimeout,
		},
		Database: Database{
			URL:                   os.Getenv("DATABASE_URL"),
			MaxConnections:        int32(databaseMaxConnections),
			MinConnections:        int32(databaseMinConnections),
			MaxConnectionLifetime: databaseMaxConnectionLifetime,
			MaxConnectionIdleTime: databaseMaxConnectionIdleTime,
			HealthCheckPeriod:     databaseHealthCheckPeriod,
			ConnectTimeout:        databaseConnectTimeout,
		},
		RateLimits: RateLimits{
			AuthenticationPerMinute: authenticationRateLimit,
			RecommendationPerMinute: recommendationRateLimit,
			AnalyticsPerMinute:      analyticsRateLimit,
			AffiliatePerMinute:      affiliateRateLimit,
			MutationPerMinute:       mutationRateLimit,
		},
		Migrations: Migrations{
			Directory: valueOrDefault("MIGRATIONS_DIR", "./migrations"),
		},
		Seeds: Seeds{
			Directory: valueOrDefault("SEEDS_DIR", "./seeds"),
		},
		Auth: Auth{
			SessionCookieName: valueOrDefault("SESSION_COOKIE_NAME", "rigmark_session"),
			SessionTTL:        sessionTTL,
			SessionIdleTTL:    sessionIdleTTL,
			CookieSecure:      cookieSecure,
		},
		Passwords: Passwords{
			MemoryKiB:   64 * 1024,
			Iterations:  3,
			Parallelism: 2,
			KeyLength:   32,
		},
		Assets: Assets{ProductImageDirectory: valueOrDefault("PRODUCT_IMAGE_UPLOAD_DIR", "./uploads/products")},
		Site:   Site{PublicURL: strings.TrimRight(valueOrDefault("PUBLIC_SITE_URL", "http://localhost:5173"), "/")},
		AI: AI{
			Provider:         valueOrDefault("AI_PROVIDER", "disabled"),
			Model:            os.Getenv("AI_MODEL"),
			APIKey:           os.Getenv("AI_API_KEY"),
			Timeout:          aiTimeout,
			MaxResponseBytes: aiMaxResponseBytes,
		},
		Commerce: Commerce{OfferMaximumAge: offerMaximumAge},
	}

	if cfg.Database.URL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	databaseURL, err := url.Parse(cfg.Database.URL)
	if err != nil || (databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql") {
		return Config{}, errors.New("DATABASE_URL must be a valid PostgreSQL URL")
	}
	if cfg.Environment == "production" {
		sslMode := databaseURL.Query().Get("sslmode")
		if sslMode != "require" && sslMode != "verify-ca" && sslMode != "verify-full" {
			return Config{}, errors.New("DATABASE_URL must enable PostgreSQL TLS in production")
		}
	}
	publicURL, err := url.Parse(cfg.Site.PublicURL)
	if err != nil || publicURL.Scheme == "" || publicURL.Host == "" || publicURL.Path != "" {
		return Config{}, errors.New("PUBLIC_SITE_URL must be an absolute origin without a path")
	}
	if cfg.Environment == "production" && publicURL.Scheme != "https" {
		return Config{}, errors.New("PUBLIC_SITE_URL must use HTTPS in production")
	}
	port, err := strconv.Atoi(cfg.HTTP.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("API_PORT must be an integer between 1 and 65535: %q", cfg.HTTP.Port)
	}
	if cfg.Auth.SessionIdleTTL <= 0 || cfg.Auth.SessionTTL <= 0 || cfg.Auth.SessionIdleTTL > cfg.Auth.SessionTTL {
		return Config{}, errors.New("SESSION_IDLE_TTL must be positive and no greater than SESSION_TTL")
	}
	if cfg.Database.MaxConnections < 1 || cfg.Database.MaxConnections > 500 ||
		cfg.Database.MinConnections < 0 || cfg.Database.MinConnections > cfg.Database.MaxConnections ||
		cfg.Database.MaxConnectionLifetime <= 0 || cfg.Database.MaxConnectionIdleTime <= 0 ||
		cfg.Database.HealthCheckPeriod <= 0 || cfg.Database.ConnectTimeout <= 0 {
		return Config{}, errors.New("database pool configuration is invalid")
	}
	for name, limit := range map[string]int{
		"RATE_LIMIT_AUTH_PER_MINUTE":           cfg.RateLimits.AuthenticationPerMinute,
		"RATE_LIMIT_RECOMMENDATION_PER_MINUTE": cfg.RateLimits.RecommendationPerMinute,
		"RATE_LIMIT_ANALYTICS_PER_MINUTE":      cfg.RateLimits.AnalyticsPerMinute,
		"RATE_LIMIT_AFFILIATE_PER_MINUTE":      cfg.RateLimits.AffiliatePerMinute,
		"RATE_LIMIT_MUTATION_PER_MINUTE":       cfg.RateLimits.MutationPerMinute,
	} {
		if limit < 1 || limit > 100_000 {
			return Config{}, fmt.Errorf("%s must be between 1 and 100000", name)
		}
	}
	if !cookieNamePattern.MatchString(cfg.Auth.SessionCookieName) {
		return Config{}, errors.New("SESSION_COOKIE_NAME must contain only letters, numbers, underscores, or hyphens")
	}
	if cfg.Environment == "production" && !cfg.Auth.CookieSecure {
		return Config{}, errors.New("SESSION_COOKIE_SECURE must be true in production")
	}
	if !providerNamePattern.MatchString(cfg.AI.Provider) || cfg.AI.Timeout <= 0 ||
		cfg.AI.MaxResponseBytes < 1_024 || cfg.AI.MaxResponseBytes > 1_048_576 {
		return Config{}, errors.New("AI provider, timeout, or maximum response bytes are invalid")
	}
	if cfg.Commerce.OfferMaximumAge < time.Hour || cfg.Commerce.OfferMaximumAge > 30*24*time.Hour {
		return Config{}, errors.New("OFFER_MAXIMUM_AGE must be between 1h and 720h")
	}
	if cfg.AI.Provider != "disabled" && (cfg.AI.Model == "" || cfg.AI.APIKey == "") {
		return Config{}, errors.New("AI_MODEL and AI_API_KEY are required when AI_PROVIDER is enabled")
	}

	return cfg, nil
}

func durationValue(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %q", key, value)
	}
	return parsed, nil
}

func boolValue(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false: %q", key, value)
	}
	return parsed, nil
}

func int64Value(key string, fallback int64) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %q", key, value)
	}
	return parsed, nil
}

func intValue(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %q", key, value)
	}
	return parsed, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
