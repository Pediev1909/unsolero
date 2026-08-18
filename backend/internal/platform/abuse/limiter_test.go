package abuse

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLocalLimiterEnforcesSharedBackendContract(t *testing.T) {
	limiter := NewLocalLimiter()
	first, err := limiter.Allow(context.Background(), "opaque-key", 1, time.Minute)
	if err != nil || !first.Allowed {
		t.Fatalf("first=%#v err=%v", first, err)
	}
	second, err := limiter.Allow(context.Background(), "opaque-key", 1, time.Minute)
	if err != nil || second.Allowed || second.RetryAfter <= 0 {
		t.Fatalf("second=%#v err=%v", second, err)
	}
}

func TestUnavailableLimiterFailsClosed(t *testing.T) {
	decision, err := (UnavailableLimiter{}).Allow(context.Background(), "opaque-key", 1, time.Minute)
	if !errors.Is(err, ErrBackendUnavailable) || decision.Allowed {
		t.Fatalf("decision=%#v err=%v", decision, err)
	}
}
