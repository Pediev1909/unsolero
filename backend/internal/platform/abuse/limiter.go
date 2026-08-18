package abuse

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrBackendUnavailable = errors.New("rate-limit backend unavailable")

type Decision struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

type Limiter interface {
	Allow(context.Context, string, int, time.Duration) (Decision, error)
	Ready(context.Context) error
}

type entry struct {
	count    int
	resetAt  time.Time
	lastSeen time.Time
}

type LocalLimiter struct {
	mu      sync.Mutex
	entries map[string]entry
	now     func() time.Time
}

func NewLocalLimiter() *LocalLimiter {
	return &LocalLimiter{entries: map[string]entry{}, now: time.Now}
}

func (limiter *LocalLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (Decision, error) {
	now := limiter.now()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	value, exists := limiter.entries[key]
	if !exists || !now.Before(value.resetAt) {
		value = entry{resetAt: now.Add(window)}
	}
	value.lastSeen = now
	if value.count >= limit {
		limiter.entries[key] = value
		return Decision{Allowed: false, RetryAfter: value.resetAt.Sub(now)}, nil
	}
	value.count++
	limiter.entries[key] = value
	if len(limiter.entries) > 10_000 {
		limiter.prune(now, window)
	}
	return Decision{Allowed: true, Remaining: limit - value.count, RetryAfter: value.resetAt.Sub(now)}, nil
}

func (limiter *LocalLimiter) Ready(context.Context) error { return nil }
func (*LocalLimiter) BackendName() string                 { return "local" }

func (limiter *LocalLimiter) prune(now time.Time, window time.Duration) {
	for key, value := range limiter.entries {
		if !now.Before(value.resetAt) || now.Sub(value.lastSeen) > 2*window {
			delete(limiter.entries, key)
		}
	}
}

type UnavailableLimiter struct{}

func (UnavailableLimiter) Allow(context.Context, string, int, time.Duration) (Decision, error) {
	return Decision{}, ErrBackendUnavailable
}
func (UnavailableLimiter) Ready(context.Context) error { return ErrBackendUnavailable }
func (UnavailableLimiter) BackendName() string         { return "unavailable" }
