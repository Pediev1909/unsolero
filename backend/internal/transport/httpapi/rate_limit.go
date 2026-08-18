package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"rigmark/internal/platform/abuse"
	"rigmark/internal/platform/alerting"
	"rigmark/internal/platform/observability"
)

const rateLimitWindow = time.Minute

type RateLimitConfig struct {
	AuthenticationPerMinute  int
	RegistrationPerMinute    int
	PasswordResetPerMinute   int
	RecommendationPerMinute  int
	AnalyticsPerMinute       int
	AffiliatePerMinute       int
	AdminPerMinute           int
	MutationPerMinute        int
	RouteResolutionPerMinute int
	TrustedProxyCIDRs        []netip.Prefix
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		AuthenticationPerMinute:  10,
		RegistrationPerMinute:    5,
		PasswordResetPerMinute:   5,
		RecommendationPerMinute:  20,
		AnalyticsPerMinute:       120,
		AffiliatePerMinute:       120,
		AdminPerMinute:           240,
		MutationPerMinute:        240,
		RouteResolutionPerMinute: 600,
	}
}

func (config RateLimitConfig) withDefaults() RateLimitConfig {
	defaults := DefaultRateLimitConfig()
	if config.AuthenticationPerMinute <= 0 {
		config.AuthenticationPerMinute = defaults.AuthenticationPerMinute
	}
	if config.RegistrationPerMinute <= 0 {
		config.RegistrationPerMinute = defaults.RegistrationPerMinute
	}
	if config.PasswordResetPerMinute <= 0 {
		config.PasswordResetPerMinute = defaults.PasswordResetPerMinute
	}
	if config.RecommendationPerMinute <= 0 {
		config.RecommendationPerMinute = defaults.RecommendationPerMinute
	}
	if config.AnalyticsPerMinute <= 0 {
		config.AnalyticsPerMinute = defaults.AnalyticsPerMinute
	}
	if config.AffiliatePerMinute <= 0 {
		config.AffiliatePerMinute = defaults.AffiliatePerMinute
	}
	if config.AdminPerMinute <= 0 {
		config.AdminPerMinute = defaults.AdminPerMinute
	}
	if config.MutationPerMinute <= 0 {
		config.MutationPerMinute = defaults.MutationPerMinute
	}
	if config.RouteResolutionPerMinute <= 0 {
		config.RouteResolutionPerMinute = defaults.RouteResolutionPerMinute
	}
	return config
}

func rateLimitRequests(next http.Handler, config RateLimitConfig, logger *slog.Logger) http.Handler {
	return rateLimitRequestsWithBackend(next, config, abuse.NewLocalLimiter(), []byte("local-test-rate-limit-key"), observability.DisabledRecorder{}, alerting.Disabled{}, logger)
}

func rateLimitRequestsWithBackend(next http.Handler, config RateLimitConfig, limiter abuse.Limiter, keySecret []byte, metrics observability.Recorder, alerts alerting.Notifier, logger *slog.Logger) http.Handler {
	config = config.withDefaults()
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		bucket, limit, protected := rateLimitRule(request, config)
		if !protected {
			next.ServeHTTP(response, request)
			return
		}
		key := opaqueRateLimitKey(keySecret, bucket, clientAddress(request, config.TrustedProxyCIDRs))
		startedAt := time.Now()
		decision, err := limiter.Allow(request.Context(), key, limit, rateLimitWindow)
		if backend, ok := limiter.(interface{ BackendName() string }); ok && backend.BackendName() == "redis" {
			metrics.SetGauge("redis_limiter_latency_milliseconds", float64(time.Since(startedAt).Microseconds())/1000)
			if err != nil {
				metrics.Increment(observability.MetricRedisUnavailable)
			}
		}
		if err != nil {
			metrics.Increment(observability.MetricRateLimitBackendFailure)
			logger.Error("rate-limit backend failed", "error", err, "bucket", bucket)
			_ = alerts.Notify(context.WithoutCancel(request.Context()), alerting.Alert{
				Category: alerting.RateLimitBackendFailure, Summary: "A protected request could not be evaluated.",
				OccurredAt: time.Now().UTC(), Severity: "critical",
			})
			writeAPIError(response, http.StatusServiceUnavailable, "abuse_protection_unavailable", "This operation is temporarily unavailable.", nil, logger)
			return
		}
		response.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		response.Header().Set("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
		if !decision.Allowed {
			metrics.Increment(observability.MetricRateLimited)
			seconds := int(decision.RetryAfter.Round(time.Second).Seconds())
			if seconds < 1 {
				seconds = 1
			}
			response.Header().Set("Retry-After", strconv.Itoa(seconds))
			writeAPIError(response, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests. Please try again shortly.", nil, logger)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func opaqueRateLimitKey(secret []byte, bucket, address string) string {
	digest := hmac.New(sha256.New, secret)
	_, _ = digest.Write([]byte(bucket))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(address))
	return bucket + ":" + hex.EncodeToString(digest.Sum(nil))
}

func rateLimitRule(request *http.Request, config RateLimitConfig) (string, int, bool) {
	path := request.URL.Path
	switch {
	case request.Method == http.MethodPost && path == "/api/auth/register":
		return "registration", config.RegistrationPerMinute, true
	case request.Method == http.MethodPost && strings.HasPrefix(path, "/api/auth/password-reset/"):
		return "password-reset", config.PasswordResetPerMinute, true
	case request.Method == http.MethodPost && (path == "/api/auth/login" ||
		strings.HasPrefix(path, "/api/auth/email-verification/") ||
		strings.HasPrefix(path, "/api/auth/mfa/")):
		return "authentication", config.AuthenticationPerMinute, true
	case (request.Method == http.MethodPost || request.Method == http.MethodDelete) &&
		(strings.HasPrefix(path, "/api/account/security/") || path == "/api/account"):
		return "account-security", config.AuthenticationPerMinute, true
	case request.Method == http.MethodPost && path == "/api/recommendations/generate":
		return "recommendation", config.RecommendationPerMinute, true
	case request.Method == http.MethodPost && path == "/api/analytics/events":
		return "analytics", config.AnalyticsPerMinute, true
	case request.Method == http.MethodGet &&
		(strings.HasPrefix(path, "/api/affiliate/click/") || strings.HasPrefix(path, "/api/out/")):
		return "affiliate", config.AffiliatePerMinute, true
	case request.Method == http.MethodGet && path == "/api/v1/public-route":
		return "public-route", config.RouteResolutionPerMinute, true
	case strings.HasPrefix(path, "/api/admin/"):
		return "admin", config.AdminPerMinute, true
	case request.Method != http.MethodGet && request.Method != http.MethodHead && request.Method != http.MethodOptions:
		return "mutation", config.MutationPerMinute, true
	default:
		return "", 0, false
	}
}

func clientAddress(request *http.Request, trustedProxyCIDRs []netip.Prefix) string {
	remote := remoteAddress(request.RemoteAddr)
	remoteIP, err := netip.ParseAddr(remote)
	if err == nil && addressInPrefixes(remoteIP, trustedProxyCIDRs) {
		forwarded := strings.TrimSpace(request.Header.Get("X-Forwarded-For"))
		if !strings.Contains(forwarded, ",") {
			if address, parseErr := netip.ParseAddr(forwarded); parseErr == nil {
				return address.String()
			}
		}
	}
	return remote
}

func addressInPrefixes(address netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func remoteAddress(value string) string {
	host, _, err := net.SplitHostPort(value)
	if err == nil && host != "" {
		return host
	}
	if value == "" {
		return "unknown"
	}
	return value
}
