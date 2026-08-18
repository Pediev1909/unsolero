package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var cookieNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
var providerNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`)

// verticalKeyPattern matches recommendation.policy_versions.vertical_key so a
// misconfigured vertical fails at startup instead of silently matching no
// active policy and failing every recommendation at request time.
var verticalKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

const (
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 10 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
)

type Config struct {
	Environment               string
	Version                   string
	AllowInsecureLocalStaging bool
	HTTP                      HTTP
	Database                  Database
	RateLimits                RateLimits
	Migrations                Migrations
	Seeds                     Seeds
	Auth                      Auth
	Passwords                 Passwords
	Assets                    Assets
	Site                      Site
	AI                        AI
	Commerce                  Commerce
	Analytics                 Analytics
	Security                  Security
	Operations                Operations
	Recommendation            Recommendation
}

// Recommendation selects which vertical this deployment serves. The schema
// permits one active recommendation policy per vertical, so changing this
// value repoints the product at a different catalog and policy without a code
// change.
type Recommendation struct {
	Vertical string
}

type Operations struct {
	AlertProvider     string
	AlertWebhookURL   string
	AlertWebhookToken string
	AlertTimeout      time.Duration
	MetricsEnabled    bool
	MetricsToken      string
}

type Analytics struct {
	SubjectCookieName      string
	AnonymousRetention     time.Duration
	AuthenticatedRetention time.Duration
	ReceiptRetention       time.Duration
	CleanupBatchSize       int
}

type Security struct {
	EmailProvider        string
	EmailSenderName      string
	EmailSenderAddress   string
	EmailSMTPAddress     string
	EmailSMTPUsername    string
	EmailSMTPPassword    string
	EmailSMTPRequireTLS  bool
	EmailSMTPTimeout     time.Duration
	MFAEncryptionKey     string
	VerificationTTL      time.Duration
	PasswordResetTTL     time.Duration
	MFAChallengeTTL      time.Duration
	MFAStepUpTTL         time.Duration
	EnforcePrivilegedMFA bool
}

type Commerce struct {
	OfferMaximumAge         time.Duration
	AffiliateClickRetention time.Duration
	WorkerPollInterval      time.Duration
	WorkerCycleTimeout      time.Duration
	WorkerLeaseTimeout      time.Duration
	WorkerMaxItemsPerCycle  int
	WorkerFailureThreshold  int
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
	StorageProvider       string
	ScanProvider          string
	S3Endpoint            string
	S3AccessKey           string
	S3SecretKey           string
	S3Bucket              string
	S3Region              string
	S3Secure              bool
	// ScanEndpoint and ScanTimeout configure the external malware scanner.
	// They are only read when ScanProvider is "external".
	ScanEndpoint string
	ScanTimeout  time.Duration
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
	HandlerTimeout    time.Duration
	MaxHeaderBytes    int
	TrustedProxyCIDRs []netip.Prefix
}

type Database struct {
	URL                    string
	MaxConnections         int32
	MinConnections         int32
	MaxConnectionLifetime  time.Duration
	MaxConnectionIdleTime  time.Duration
	HealthCheckPeriod      time.Duration
	ConnectTimeout         time.Duration
	StatementTimeout       time.Duration
	LockTimeout            time.Duration
	IdleTransactionTimeout time.Duration
	MigrationTimeout       time.Duration
}

type RateLimits struct {
	Provider                 string
	ReplicaCount             int
	KeySecret                string
	RedisURL                 string
	Namespace                string
	AuthenticationPerMinute  int
	RegistrationPerMinute    int
	PasswordResetPerMinute   int
	RecommendationPerMinute  int
	AnalyticsPerMinute       int
	AffiliatePerMinute       int
	AdminPerMinute           int
	MutationPerMinute        int
	RouteResolutionPerMinute int
}

type Migrations struct {
	Directory string
}

type Seeds struct {
	Directory string
}

func Load() (Config, error) {
	environment := valueOrDefault("APP_ENV", "development")
	deployedEnvironment := environment == "production" || environment == "staging"
	readHeaderTimeout, err := durationValue("HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := durationValue("HTTP_READ_TIMEOUT", defaultReadTimeout)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := durationValue("HTTP_WRITE_TIMEOUT", defaultWriteTimeout)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := durationValue("HTTP_IDLE_TIMEOUT", defaultIdleTimeout)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := durationValue("HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}
	handlerTimeout, err := durationValue("HTTP_HANDLER_TIMEOUT", 20*time.Second)
	if err != nil {
		return Config{}, err
	}
	maxHeaderBytes, err := intValue("HTTP_MAX_HEADER_BYTES", 32*1024)
	if err != nil {
		return Config{}, err
	}
	trustedProxyCIDRs, err := prefixListValue("TRUSTED_PROXY_CIDRS")
	if err != nil {
		return Config{}, err
	}
	cookieSecure, err := boolValue("SESSION_COOKIE_SECURE", deployedEnvironment || environment == "test")
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
	databaseStatementTimeout, err := durationValue("DATABASE_STATEMENT_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	databaseLockTimeout, err := durationValue("DATABASE_LOCK_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	databaseIdleTransactionTimeout, err := durationValue("DATABASE_IDLE_TRANSACTION_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	databaseMigrationTimeout, err := durationValue("DATABASE_MIGRATION_TIMEOUT", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	replicaCount, err := intValue("API_REPLICA_COUNT", 1)
	if err != nil {
		return Config{}, err
	}
	authenticationRateLimit, err := intValue("RATE_LIMIT_AUTH_PER_MINUTE", 10)
	if err != nil {
		return Config{}, err
	}
	registrationRateLimit, err := intValue("RATE_LIMIT_REGISTRATION_PER_MINUTE", 5)
	if err != nil {
		return Config{}, err
	}
	passwordResetRateLimit, err := intValue("RATE_LIMIT_PASSWORD_RESET_PER_MINUTE", 5)
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
	adminRateLimit, err := intValue("RATE_LIMIT_ADMIN_PER_MINUTE", 240)
	if err != nil {
		return Config{}, err
	}
	mutationRateLimit, err := intValue("RATE_LIMIT_MUTATION_PER_MINUTE", 240)
	if err != nil {
		return Config{}, err
	}
	routeResolutionRateLimit, err := intValue("RATE_LIMIT_ROUTE_PER_MINUTE", 600)
	if err != nil {
		return Config{}, err
	}
	offerMaximumAge, err := durationValue("OFFER_MAXIMUM_AGE", 72*time.Hour)
	if err != nil {
		return Config{}, err
	}
	affiliateClickRetention, err := durationValue("AFFILIATE_CLICK_RETENTION", 397*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	commerceWorkerPoll, err := durationValue("COMMERCE_WORKER_POLL_INTERVAL", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	commerceWorkerCycleTimeout, err := durationValue("WORKER_CYCLE_TIMEOUT", 2*time.Minute)
	if err != nil {
		return Config{}, err
	}
	commerceWorkerLeaseTimeout, err := durationValue("WORKER_LEASE_TIMEOUT", time.Hour)
	if err != nil {
		return Config{}, err
	}
	commerceWorkerMaxItems, err := intValue("WORKER_MAX_ITEMS_PER_CYCLE", 25)
	if err != nil {
		return Config{}, err
	}
	commerceWorkerFailureThreshold, err := intValue("WORKER_FAILURE_ALERT_THRESHOLD", 3)
	if err != nil {
		return Config{}, err
	}
	analyticsAnonymousRetention, err := durationValue("ANALYTICS_ANONYMOUS_RETENTION", 90*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	analyticsAuthenticatedRetention, err := durationValue("ANALYTICS_AUTHENTICATED_RETENTION", 397*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	analyticsReceiptRetention, err := durationValue("ANALYTICS_RECEIPT_RETENTION", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	analyticsCleanupBatch, err := intValue("ANALYTICS_CLEANUP_BATCH_SIZE", 1000)
	if err != nil {
		return Config{}, err
	}
	verificationTTL, err := durationValue("EMAIL_VERIFICATION_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	passwordResetTTL, err := durationValue("PASSWORD_RESET_TTL", time.Hour)
	if err != nil {
		return Config{}, err
	}
	mfaChallengeTTL, err := durationValue("MFA_CHALLENGE_TTL", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	mfaStepUpTTL, err := durationValue("MFA_STEP_UP_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	enforcePrivilegedMFA, err := boolValue("MFA_ENFORCE_PRIVILEGED", deployedEnvironment)
	if err != nil {
		return Config{}, err
	}
	metricsEnabled, err := boolValue("METRICS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	alertTimeout, err := durationValue("ALERT_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	mediaScanTimeout, err := durationValue("MEDIA_SCAN_TIMEOUT", 20*time.Second)
	if err != nil {
		return Config{}, err
	}
	allowInsecureLocalStaging, err := boolValue("ALLOW_INSECURE_LOCAL_STAGING", false)
	if err != nil {
		return Config{}, err
	}
	emailSMTPRequireTLS, err := boolValue("EMAIL_SMTP_REQUIRE_TLS", environment == "production")
	if err != nil {
		return Config{}, err
	}
	emailSMTPTimeout, err := durationValue("EMAIL_SMTP_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	s3Secure, err := boolValue("MEDIA_S3_SECURE", environment == "production")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:               environment,
		Version:                   valueOrDefault("APP_VERSION", "development"),
		AllowInsecureLocalStaging: allowInsecureLocalStaging,
		HTTP: HTTP{
			Port:              valueOrDefault("API_PORT", "8080"),
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
			ShutdownTimeout:   shutdownTimeout,
			HandlerTimeout:    handlerTimeout,
			MaxHeaderBytes:    maxHeaderBytes,
			TrustedProxyCIDRs: trustedProxyCIDRs,
		},
		Database: Database{
			URL:                    os.Getenv("DATABASE_URL"),
			MaxConnections:         int32(databaseMaxConnections),
			MinConnections:         int32(databaseMinConnections),
			MaxConnectionLifetime:  databaseMaxConnectionLifetime,
			MaxConnectionIdleTime:  databaseMaxConnectionIdleTime,
			HealthCheckPeriod:      databaseHealthCheckPeriod,
			ConnectTimeout:         databaseConnectTimeout,
			StatementTimeout:       databaseStatementTimeout,
			LockTimeout:            databaseLockTimeout,
			IdleTransactionTimeout: databaseIdleTransactionTimeout,
			MigrationTimeout:       databaseMigrationTimeout,
		},
		RateLimits: RateLimits{
			Provider:                 valueOrDefault("RATE_LIMIT_PROVIDER", "local"),
			ReplicaCount:             replicaCount,
			KeySecret:                os.Getenv("RATE_LIMIT_KEY_SECRET"),
			RedisURL:                 os.Getenv("RATE_LIMIT_REDIS_URL"),
			Namespace:                valueOrDefault("RATE_LIMIT_NAMESPACE", "unsolero:rate-limit"),
			AuthenticationPerMinute:  authenticationRateLimit,
			RegistrationPerMinute:    registrationRateLimit,
			PasswordResetPerMinute:   passwordResetRateLimit,
			RecommendationPerMinute:  recommendationRateLimit,
			AnalyticsPerMinute:       analyticsRateLimit,
			AffiliatePerMinute:       affiliateRateLimit,
			AdminPerMinute:           adminRateLimit,
			MutationPerMinute:        mutationRateLimit,
			RouteResolutionPerMinute: routeResolutionRateLimit,
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
		Assets: Assets{
			ProductImageDirectory: valueOrDefault("PRODUCT_IMAGE_UPLOAD_DIR", "./uploads/products"),
			StorageProvider:       valueOrDefault("MEDIA_STORAGE_PROVIDER", "local"),
			ScanProvider:          valueOrDefault("MEDIA_SCAN_PROVIDER", "development"),
			ScanEndpoint:          os.Getenv("MEDIA_SCAN_ENDPOINT"),
			ScanTimeout:           mediaScanTimeout,
			S3Endpoint:            os.Getenv("MEDIA_S3_ENDPOINT"),
			S3AccessKey:           os.Getenv("MEDIA_S3_ACCESS_KEY"),
			S3SecretKey:           os.Getenv("MEDIA_S3_SECRET_KEY"),
			S3Bucket:              os.Getenv("MEDIA_S3_BUCKET"),
			S3Region:              os.Getenv("MEDIA_S3_REGION"),
			S3Secure:              s3Secure,
		},
		Site:           Site{PublicURL: strings.TrimRight(valueOrDefault("PUBLIC_SITE_URL", "http://localhost:5173"), "/")},
		Recommendation: Recommendation{Vertical: valueOrDefault("RECOMMENDATION_VERTICAL", "fitness")},
		AI: AI{
			Provider:         valueOrDefault("AI_PROVIDER", "disabled"),
			Model:            os.Getenv("AI_MODEL"),
			APIKey:           os.Getenv("AI_API_KEY"),
			Timeout:          aiTimeout,
			MaxResponseBytes: aiMaxResponseBytes,
		},
		Commerce: Commerce{OfferMaximumAge: offerMaximumAge, AffiliateClickRetention: affiliateClickRetention,
			WorkerPollInterval: commerceWorkerPoll, WorkerCycleTimeout: commerceWorkerCycleTimeout,
			WorkerLeaseTimeout: commerceWorkerLeaseTimeout, WorkerMaxItemsPerCycle: commerceWorkerMaxItems,
			WorkerFailureThreshold: commerceWorkerFailureThreshold},
		Analytics: Analytics{
			SubjectCookieName:      valueOrDefault("ANALYTICS_SUBJECT_COOKIE_NAME", "unsolero_analytics_subject"),
			AnonymousRetention:     analyticsAnonymousRetention,
			AuthenticatedRetention: analyticsAuthenticatedRetention,
			ReceiptRetention:       analyticsReceiptRetention,
			CleanupBatchSize:       analyticsCleanupBatch,
		},
		Security: Security{EmailProvider: valueOrDefault("EMAIL_PROVIDER", "development"),
			EmailSenderName: valueOrDefault("EMAIL_SENDER_NAME", "UNSOLERO Security"), EmailSenderAddress: os.Getenv("EMAIL_SENDER_ADDRESS"),
			EmailSMTPAddress: os.Getenv("EMAIL_SMTP_ADDRESS"), EmailSMTPUsername: os.Getenv("EMAIL_SMTP_USERNAME"),
			EmailSMTPPassword: os.Getenv("EMAIL_SMTP_PASSWORD"), EmailSMTPRequireTLS: emailSMTPRequireTLS, EmailSMTPTimeout: emailSMTPTimeout,
			MFAEncryptionKey: os.Getenv("MFA_ENCRYPTION_KEY"), VerificationTTL: verificationTTL,
			PasswordResetTTL: passwordResetTTL, MFAChallengeTTL: mfaChallengeTTL,
			MFAStepUpTTL: mfaStepUpTTL, EnforcePrivilegedMFA: enforcePrivilegedMFA},
		Operations: Operations{AlertProvider: valueOrDefault("ALERT_PROVIDER", "disabled"),
			AlertWebhookURL: os.Getenv("ALERT_WEBHOOK_URL"), AlertWebhookToken: os.Getenv("ALERT_WEBHOOK_TOKEN"),
			AlertTimeout: alertTimeout, MetricsEnabled: metricsEnabled, MetricsToken: os.Getenv("METRICS_TOKEN")},
	}
	if cfg.Environment != "development" && cfg.Environment != "test" && cfg.Environment != "staging" && cfg.Environment != "production" {
		return Config{}, errors.New("APP_ENV must be development, test, staging, or production")
	}
	if cfg.AllowInsecureLocalStaging && cfg.Environment != "staging" {
		return Config{}, errors.New("ALLOW_INSECURE_LOCAL_STAGING is only valid for an isolated staging environment")
	}
	if deployedEnvironment && (cfg.Version == "" || cfg.Version == "development") {
		return Config{}, errors.New("APP_VERSION must identify the deployed release in staging and production")
	}

	if cfg.Database.URL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	databaseURL, err := url.Parse(cfg.Database.URL)
	if err != nil || (databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql") {
		return Config{}, errors.New("DATABASE_URL must be a valid PostgreSQL URL")
	}
	strictDeployedEnvironment := cfg.Environment == "production" || (cfg.Environment == "staging" && !cfg.AllowInsecureLocalStaging)
	if strictDeployedEnvironment {
		sslMode := databaseURL.Query().Get("sslmode")
		if sslMode != "require" && sslMode != "verify-ca" && sslMode != "verify-full" {
			return Config{}, errors.New("DATABASE_URL must enable PostgreSQL TLS in hosted staging and production")
		}
	}
	if cfg.Environment == "production" && localOnlyHostname(databaseURL.Hostname()) {
		return Config{}, errors.New("DATABASE_URL must not target a loopback service in production")
	}
	publicURL, err := url.Parse(cfg.Site.PublicURL)
	if err != nil || publicURL.Scheme == "" || publicURL.Host == "" || publicURL.Path != "" {
		return Config{}, errors.New("PUBLIC_SITE_URL must be an absolute origin without a path")
	}
	if deployedEnvironment && publicURL.Scheme != "https" {
		return Config{}, errors.New("PUBLIC_SITE_URL must use HTTPS in staging and production")
	}
	if cfg.Environment == "production" && !publicProductionHostname(publicURL.Hostname()) {
		return Config{}, errors.New("PUBLIC_SITE_URL must use a public DNS hostname in production")
	}
	if cfg.AllowInsecureLocalStaging && publicURL.Hostname() != "localhost" && publicURL.Hostname() != "127.0.0.1" && publicURL.Hostname() != "::1" {
		return Config{}, errors.New("ALLOW_INSECURE_LOCAL_STAGING requires a loopback PUBLIC_SITE_URL")
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
		cfg.Database.HealthCheckPeriod <= 0 || cfg.Database.ConnectTimeout <= 0 ||
		cfg.Database.StatementTimeout < time.Second || cfg.Database.StatementTimeout > 5*time.Minute ||
		cfg.Database.LockTimeout < 100*time.Millisecond || cfg.Database.LockTimeout > time.Minute ||
		cfg.Database.IdleTransactionTimeout < time.Second || cfg.Database.IdleTransactionTimeout > 10*time.Minute ||
		cfg.Database.MigrationTimeout < time.Minute || cfg.Database.MigrationTimeout > time.Hour {
		return Config{}, errors.New("database pool configuration is invalid")
	}
	for name, value := range map[string]time.Duration{
		"HTTP_READ_HEADER_TIMEOUT": cfg.HTTP.ReadHeaderTimeout, "HTTP_READ_TIMEOUT": cfg.HTTP.ReadTimeout,
		"HTTP_WRITE_TIMEOUT": cfg.HTTP.WriteTimeout, "HTTP_IDLE_TIMEOUT": cfg.HTTP.IdleTimeout,
		"HTTP_SHUTDOWN_TIMEOUT": cfg.HTTP.ShutdownTimeout, "HTTP_HANDLER_TIMEOUT": cfg.HTTP.HandlerTimeout,
	} {
		if value < time.Second || value > 5*time.Minute {
			return Config{}, fmt.Errorf("%s must be between 1s and 5m", name)
		}
	}
	if cfg.HTTP.WriteTimeout <= cfg.HTTP.HandlerTimeout {
		return Config{}, errors.New("HTTP_WRITE_TIMEOUT must be greater than HTTP_HANDLER_TIMEOUT")
	}
	if cfg.HTTP.MaxHeaderBytes < 8*1024 || cfg.HTTP.MaxHeaderBytes > 1024*1024 {
		return Config{}, errors.New("HTTP_MAX_HEADER_BYTES must be between 8192 and 1048576")
	}
	if cfg.RateLimits.Provider != "local" && cfg.RateLimits.Provider != "redis" && cfg.RateLimits.Provider != "external" {
		return Config{}, errors.New("RATE_LIMIT_PROVIDER must be local, redis, or external")
	}
	if cfg.RateLimits.ReplicaCount < 1 || cfg.RateLimits.ReplicaCount > 1000 {
		return Config{}, errors.New("API_REPLICA_COUNT must be between 1 and 1000")
	}
	if cfg.RateLimits.ReplicaCount > 1 && cfg.RateLimits.Provider != "redis" && cfg.RateLimits.Provider != "external" {
		return Config{}, errors.New("multiple API replicas require a distributed rate-limit provider")
	}
	if cfg.RateLimits.Provider == "redis" {
		redisURL, parseErr := url.Parse(cfg.RateLimits.RedisURL)
		if parseErr != nil || (redisURL.Scheme != "redis" && redisURL.Scheme != "rediss") || redisURL.Host == "" {
			return Config{}, errors.New("RATE_LIMIT_REDIS_URL must be a valid redis or rediss URL")
		}
		if strictDeployedEnvironment && redisURL.Scheme != "rediss" {
			return Config{}, errors.New("RATE_LIMIT_REDIS_URL must use TLS in hosted staging and production")
		}
		if cfg.Environment == "production" && localOnlyHostname(redisURL.Hostname()) {
			return Config{}, errors.New("RATE_LIMIT_REDIS_URL must not target a loopback service in production")
		}
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9:_-]{1,64}$`).MatchString(cfg.RateLimits.Namespace) {
		return Config{}, errors.New("RATE_LIMIT_NAMESPACE is invalid")
	}
	if deployedEnvironment {
		if _, err := decodeSecret32(cfg.RateLimits.KeySecret); err != nil {
			return Config{}, errors.New("RATE_LIMIT_KEY_SECRET must be raw-base64 for exactly 32 bytes in staging and production")
		}
	}
	for name, limit := range map[string]int{
		"RATE_LIMIT_AUTH_PER_MINUTE":           cfg.RateLimits.AuthenticationPerMinute,
		"RATE_LIMIT_REGISTRATION_PER_MINUTE":   cfg.RateLimits.RegistrationPerMinute,
		"RATE_LIMIT_PASSWORD_RESET_PER_MINUTE": cfg.RateLimits.PasswordResetPerMinute,
		"RATE_LIMIT_RECOMMENDATION_PER_MINUTE": cfg.RateLimits.RecommendationPerMinute,
		"RATE_LIMIT_ANALYTICS_PER_MINUTE":      cfg.RateLimits.AnalyticsPerMinute,
		"RATE_LIMIT_AFFILIATE_PER_MINUTE":      cfg.RateLimits.AffiliatePerMinute,
		"RATE_LIMIT_ADMIN_PER_MINUTE":          cfg.RateLimits.AdminPerMinute,
		"RATE_LIMIT_MUTATION_PER_MINUTE":       cfg.RateLimits.MutationPerMinute,
		"RATE_LIMIT_ROUTE_PER_MINUTE":          cfg.RateLimits.RouteResolutionPerMinute,
	} {
		if limit < 1 || limit > 100_000 {
			return Config{}, fmt.Errorf("%s must be between 1 and 100000", name)
		}
	}
	if !cookieNamePattern.MatchString(cfg.Auth.SessionCookieName) {
		return Config{}, errors.New("SESSION_COOKIE_NAME must contain only letters, numbers, underscores, or hyphens")
	}
	if deployedEnvironment && !cfg.Auth.CookieSecure {
		return Config{}, errors.New("SESSION_COOKIE_SECURE must be true in staging and production")
	}
	if cfg.Assets.StorageProvider != "local" && cfg.Assets.StorageProvider != "s3" && cfg.Assets.StorageProvider != "external" {
		return Config{}, errors.New("MEDIA_STORAGE_PROVIDER must be local, s3, or external")
	}
	if cfg.Assets.ScanProvider != "development" && cfg.Assets.ScanProvider != "disabled" && cfg.Assets.ScanProvider != "external" {
		return Config{}, errors.New("MEDIA_SCAN_PROVIDER must be development, disabled, or external")
	}
	if deployedEnvironment && cfg.Assets.StorageProvider != "s3" && cfg.Assets.StorageProvider != "external" {
		return Config{}, errors.New("staging and production require external durable media storage")
	}
	if cfg.Assets.StorageProvider == "s3" {
		if cfg.Assets.S3Endpoint == "" || strings.Contains(cfg.Assets.S3Endpoint, "://") ||
			cfg.Assets.S3AccessKey == "" || cfg.Assets.S3SecretKey == "" || cfg.Assets.S3Bucket == "" {
			return Config{}, errors.New("S3-compatible media storage configuration is incomplete")
		}
		if strictDeployedEnvironment && !cfg.Assets.S3Secure {
			return Config{}, errors.New("S3-compatible media storage must use TLS in hosted staging and production")
		}
		s3URL, s3ParseErr := url.Parse("//" + cfg.Assets.S3Endpoint)
		if s3ParseErr != nil || s3URL.Hostname() == "" {
			return Config{}, errors.New("MEDIA_S3_ENDPOINT must be a valid host and optional port")
		}
		if cfg.Environment == "production" && localOnlyHostname(s3URL.Hostname()) {
			return Config{}, errors.New("MEDIA_S3_ENDPOINT must not target a loopback service in production")
		}
	}
	if cfg.Assets.ScanProvider == "external" {
		if strings.TrimSpace(cfg.Assets.ScanEndpoint) == "" {
			return Config{}, errors.New("MEDIA_SCAN_ENDPOINT is required when MEDIA_SCAN_PROVIDER is external")
		}
		scanURL, scanErr := url.Parse("//" + cfg.Assets.ScanEndpoint)
		if scanErr != nil || scanURL.Hostname() == "" || scanURL.Port() == "" {
			return Config{}, errors.New("MEDIA_SCAN_ENDPOINT must be a host and port")
		}
		if cfg.Assets.ScanTimeout < time.Second || cfg.Assets.ScanTimeout > time.Minute {
			return Config{}, errors.New("MEDIA_SCAN_TIMEOUT must be between 1s and 1m")
		}
	}
	if strictDeployedEnvironment && cfg.Assets.ScanProvider != "external" {
		return Config{}, errors.New("hosted staging and production require an external media scanning adapter")
	}
	if !verticalKeyPattern.MatchString(cfg.Recommendation.Vertical) {
		return Config{}, errors.New("RECOMMENDATION_VERTICAL must be a lowercase key such as fitness or saas")
	}
	if !providerNamePattern.MatchString(cfg.AI.Provider) || cfg.AI.Timeout <= 0 ||
		cfg.AI.MaxResponseBytes < 1_024 || cfg.AI.MaxResponseBytes > 1_048_576 {
		return Config{}, errors.New("AI provider, timeout, or maximum response bytes are invalid")
	}
	if cfg.Commerce.OfferMaximumAge < time.Hour || cfg.Commerce.OfferMaximumAge > 30*24*time.Hour {
		return Config{}, errors.New("OFFER_MAXIMUM_AGE must be between 1h and 720h")
	}
	if cfg.Commerce.AffiliateClickRetention < 30*24*time.Hour || cfg.Commerce.AffiliateClickRetention > 3*365*24*time.Hour {
		return Config{}, errors.New("AFFILIATE_CLICK_RETENTION must be between 720h and 26280h")
	}
	if cfg.Commerce.WorkerPollInterval < time.Second || cfg.Commerce.WorkerPollInterval > 5*time.Minute {
		return Config{}, errors.New("COMMERCE_WORKER_POLL_INTERVAL must be between 1s and 5m")
	}
	if cfg.Commerce.WorkerCycleTimeout < 10*time.Second || cfg.Commerce.WorkerCycleTimeout > 30*time.Minute ||
		cfg.Commerce.WorkerLeaseTimeout < time.Minute || cfg.Commerce.WorkerLeaseTimeout > 24*time.Hour ||
		cfg.Commerce.WorkerMaxItemsPerCycle < 1 || cfg.Commerce.WorkerMaxItemsPerCycle > 1000 ||
		cfg.Commerce.WorkerFailureThreshold < 1 || cfg.Commerce.WorkerFailureThreshold > 100 {
		return Config{}, errors.New("worker reliability configuration is outside safe bounds")
	}
	if !cookieNamePattern.MatchString(cfg.Analytics.SubjectCookieName) {
		return Config{}, errors.New("ANALYTICS_SUBJECT_COOKIE_NAME must contain only letters, numbers, underscores, or hyphens")
	}
	if cfg.Analytics.AnonymousRetention < 24*time.Hour || cfg.Analytics.AnonymousRetention > 2*365*24*time.Hour ||
		cfg.Analytics.AuthenticatedRetention < 24*time.Hour || cfg.Analytics.AuthenticatedRetention > 3*365*24*time.Hour ||
		cfg.Analytics.ReceiptRetention < 24*time.Hour || cfg.Analytics.ReceiptRetention > 180*24*time.Hour ||
		cfg.Analytics.CleanupBatchSize < 1 || cfg.Analytics.CleanupBatchSize > 10_000 {
		return Config{}, errors.New("analytics retention or cleanup configuration is outside safe bounds")
	}
	if cfg.AI.Provider != "disabled" && (cfg.AI.Model == "" || cfg.AI.APIKey == "") {
		return Config{}, errors.New("AI_MODEL and AI_API_KEY are required when AI_PROVIDER is enabled")
	}
	if cfg.Security.VerificationTTL < 15*time.Minute || cfg.Security.VerificationTTL > 7*24*time.Hour ||
		cfg.Security.PasswordResetTTL < 10*time.Minute || cfg.Security.PasswordResetTTL > 24*time.Hour ||
		cfg.Security.MFAChallengeTTL < time.Minute || cfg.Security.MFAChallengeTTL > 15*time.Minute ||
		cfg.Security.MFAStepUpTTL < time.Minute || cfg.Security.MFAStepUpTTL > time.Hour {
		return Config{}, errors.New("account security token lifetimes are outside safe bounds")
	}
	if cfg.Security.EmailProvider != "development" && cfg.Security.EmailProvider != "disabled" && cfg.Security.EmailProvider != "smtp" && cfg.Security.EmailProvider != "external" {
		return Config{}, errors.New("EMAIL_PROVIDER must be development, disabled, smtp, or external")
	}
	if strictDeployedEnvironment && (cfg.Security.EmailProvider == "development" || cfg.Security.EmailProvider == "disabled") {
		return Config{}, errors.New("hosted staging and production require an external email delivery adapter")
	}
	if cfg.Security.EmailProvider == "smtp" {
		if cfg.Security.EmailSenderAddress == "" || cfg.Security.EmailSMTPAddress == "" ||
			(cfg.Security.EmailSMTPUsername == "") != (cfg.Security.EmailSMTPPassword == "") ||
			cfg.Security.EmailSMTPTimeout < time.Second || cfg.Security.EmailSMTPTimeout > time.Minute {
			return Config{}, errors.New("SMTP email delivery configuration is incomplete")
		}
		if strictDeployedEnvironment && !cfg.Security.EmailSMTPRequireTLS {
			return Config{}, errors.New("SMTP email delivery must require TLS in hosted staging and production")
		}
		smtpURL, smtpParseErr := url.Parse("//" + cfg.Security.EmailSMTPAddress)
		if smtpParseErr != nil || smtpURL.Hostname() == "" {
			return Config{}, errors.New("EMAIL_SMTP_ADDRESS must be a valid host and optional port")
		}
		if cfg.Environment == "production" && localOnlyHostname(smtpURL.Hostname()) {
			return Config{}, errors.New("EMAIL_SMTP_ADDRESS must not target a loopback service in production")
		}
	}
	if deployedEnvironment && cfg.Security.MFAEncryptionKey == "" {
		return Config{}, errors.New("MFA_ENCRYPTION_KEY is required in staging and production")
	}
	if cfg.Security.MFAEncryptionKey != "" {
		if _, err := decodeSecret32(cfg.Security.MFAEncryptionKey); err != nil {
			return Config{}, errors.New("MFA_ENCRYPTION_KEY must be raw-base64 for exactly 32 bytes")
		}
	}
	if cfg.Operations.AlertProvider != "disabled" && cfg.Operations.AlertProvider != "webhook" && cfg.Operations.AlertProvider != "external" {
		return Config{}, errors.New("ALERT_PROVIDER must be disabled, webhook, or external")
	}
	if cfg.Operations.AlertProvider == "webhook" {
		alertURL, parseErr := url.Parse(cfg.Operations.AlertWebhookURL)
		if parseErr != nil || alertURL.Host == "" || (alertURL.Scheme != "http" && alertURL.Scheme != "https") ||
			alertURL.User != nil || alertURL.RawQuery != "" || alertURL.Fragment != "" || len(cfg.Operations.AlertWebhookToken) < 32 {
			return Config{}, errors.New("alert webhook configuration is incomplete or unsafe")
		}
		if strictDeployedEnvironment && alertURL.Scheme != "https" {
			return Config{}, errors.New("ALERT_WEBHOOK_URL must use HTTPS in hosted staging and production")
		}
		if cfg.Environment == "production" && localOnlyHostname(alertURL.Hostname()) {
			return Config{}, errors.New("ALERT_WEBHOOK_URL must not target a loopback service in production")
		}
	}
	if cfg.Operations.AlertTimeout < time.Second || cfg.Operations.AlertTimeout > 30*time.Second {
		return Config{}, errors.New("ALERT_TIMEOUT must be between 1s and 30s")
	}
	if cfg.Operations.MetricsEnabled && len(cfg.Operations.MetricsToken) < 32 {
		return Config{}, errors.New("METRICS_TOKEN must contain at least 32 characters when metrics are enabled")
	}
	if strictDeployedEnvironment && cfg.Operations.AlertProvider == "disabled" {
		return Config{}, errors.New("hosted staging and production require an operational alert provider")
	}
	if strictDeployedEnvironment && !cfg.Operations.MetricsEnabled {
		return Config{}, errors.New("hosted staging and production require authenticated operational metrics")
	}

	return cfg, nil
}

func decodeSecret32(value string) ([]byte, error) {
	decoded, err := base64.RawStdEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return nil, errors.New("secret must be raw-base64 for exactly 32 bytes")
	}
	return decoded, nil
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

func prefixListValue(key string) ([]netip.Prefix, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	prefixes := make([]netip.Prefix, 0, len(parts))
	for _, part := range parts {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("%s must contain only valid IP CIDR prefixes: %q", key, part)
		}
		minimumBits := 8
		if prefix.Addr().Is6() {
			minimumBits = 32
		}
		if prefix.Bits() < minimumBits {
			return nil, fmt.Errorf("%s contains an unreasonably broad trusted proxy prefix: %q", key, part)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func publicProductionHostname(value string) bool {
	hostname := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
	if hostname == "" || hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") ||
		strings.HasSuffix(hostname, ".local") || strings.HasSuffix(hostname, ".internal") ||
		strings.HasSuffix(hostname, ".invalid") || strings.HasSuffix(hostname, ".test") ||
		strings.HasSuffix(hostname, ".example") || !strings.Contains(hostname, ".") {
		return false
	}
	if _, err := netip.ParseAddr(hostname); err == nil {
		return false
	}
	return true
}

func localOnlyHostname(value string) bool {
	hostname := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
	if hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") {
		return true
	}
	address, err := netip.ParseAddr(hostname)
	return err == nil && (address.IsLoopback() || address.IsUnspecified())
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
