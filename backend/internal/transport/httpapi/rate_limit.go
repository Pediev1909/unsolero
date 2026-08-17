package httpapi

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const rateLimitWindow = time.Minute

type RateLimitConfig struct {
	AuthenticationPerMinute int
	RecommendationPerMinute int
	AnalyticsPerMinute      int
	AffiliatePerMinute      int
	MutationPerMinute       int
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		AuthenticationPerMinute: 10,
		RecommendationPerMinute: 20,
		AnalyticsPerMinute:      120,
		AffiliatePerMinute:      120,
		MutationPerMinute:       240,
	}
}

func (config RateLimitConfig) withDefaults() RateLimitConfig {
	defaults := DefaultRateLimitConfig()
	if config.AuthenticationPerMinute <= 0 {
		config.AuthenticationPerMinute = defaults.AuthenticationPerMinute
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
	if config.MutationPerMinute <= 0 {
		config.MutationPerMinute = defaults.MutationPerMinute
	}
	return config
}

type rateLimitEntry struct {
	count    int
	resetAt  time.Time
	lastSeen time.Time
}

type fixedWindowLimiter struct {
	mu      sync.Mutex
	entries map[string]rateLimitEntry
	now     func() time.Time
}

func newFixedWindowLimiter() *fixedWindowLimiter {
	return &fixedWindowLimiter{entries: make(map[string]rateLimitEntry), now: time.Now}
}

func (limiter *fixedWindowLimiter) allow(key string, limit int) (bool, int, time.Duration) {
	now := limiter.now()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	entry, exists := limiter.entries[key]
	if !exists || !now.Before(entry.resetAt) {
		entry = rateLimitEntry{resetAt: now.Add(rateLimitWindow)}
	}
	entry.lastSeen = now
	if entry.count >= limit {
		limiter.entries[key] = entry
		return false, 0, entry.resetAt.Sub(now)
	}
	entry.count++
	limiter.entries[key] = entry
	if len(limiter.entries) > 10_000 {
		limiter.prune(now)
	}
	return true, limit - entry.count, entry.resetAt.Sub(now)
}

func (limiter *fixedWindowLimiter) prune(now time.Time) {
	for key, entry := range limiter.entries {
		if !now.Before(entry.resetAt) || now.Sub(entry.lastSeen) > 2*rateLimitWindow {
			delete(limiter.entries, key)
		}
	}
}

func rateLimitRequests(next http.Handler, config RateLimitConfig, logger *slog.Logger) http.Handler {
	config = config.withDefaults()
	limiter := newFixedWindowLimiter()
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		bucket, limit, protected := rateLimitRule(request, config)
		if !protected {
			next.ServeHTTP(response, request)
			return
		}
		key := bucket + ":" + clientAddress(request)
		allowed, remaining, retryAfter := limiter.allow(key, limit)
		response.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		response.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if !allowed {
			seconds := int(retryAfter.Round(time.Second).Seconds())
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

func rateLimitRule(request *http.Request, config RateLimitConfig) (string, int, bool) {
	path := request.URL.Path
	switch {
	case request.Method == http.MethodPost && (path == "/api/auth/login" || path == "/api/auth/register"):
		return "authentication", config.AuthenticationPerMinute, true
	case request.Method == http.MethodPost && path == "/api/recommendations/generate":
		return "recommendation", config.RecommendationPerMinute, true
	case request.Method == http.MethodPost && path == "/api/analytics/events":
		return "analytics", config.AnalyticsPerMinute, true
	case request.Method == http.MethodGet &&
		(strings.HasPrefix(path, "/api/affiliate/click/") || strings.HasPrefix(path, "/api/out/")):
		return "affiliate", config.AffiliatePerMinute, true
	case request.Method != http.MethodGet && request.Method != http.MethodHead && request.Method != http.MethodOptions:
		return "mutation", config.MutationPerMinute, true
	default:
		return "", 0, false
	}
}

func clientAddress(request *http.Request) string {
	remote := remoteAddress(request.RemoteAddr)
	remoteIP := net.ParseIP(remote)
	if remoteIP != nil && (remoteIP.IsPrivate() || remoteIP.IsLoopback()) {
		forwarded := strings.TrimSpace(request.Header.Get("X-Forwarded-For"))
		if !strings.Contains(forwarded, ",") && net.ParseIP(forwarded) != nil {
			return forwarded
		}
	}
	return remote
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
