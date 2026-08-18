package abuse

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisLimiterSharesAtomicStateAcrossReplicas(t *testing.T) {
	rawURL := os.Getenv("TEST_REDIS_URL")
	if rawURL == "" {
		t.Skip("TEST_REDIS_URL is not configured")
	}
	namespace := fmt.Sprintf("unsolero:test:%d", time.Now().UnixNano())
	first, err := NewRedisLimiterFromURL(rawURL, namespace)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := NewRedisLimiterFromURL(rawURL, namespace)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	const requests = 40
	const limit = 13
	var allowed atomic.Int64
	var wait sync.WaitGroup
	for index := 0; index < requests; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			limiter := first
			if index%2 == 1 {
				limiter = second
			}
			decision, allowErr := limiter.Allow(context.Background(), "authentication:opaque", limit, time.Minute)
			if allowErr != nil {
				t.Errorf("Allow() error = %v", allowErr)
				return
			}
			if decision.Allowed {
				allowed.Add(1)
			}
		}(index)
	}
	wait.Wait()
	if got := allowed.Load(); got != limit {
		t.Fatalf("allowed = %d, want %d", got, limit)
	}
}

func TestRedisLimiterExpiresAndSeparatesRoutes(t *testing.T) {
	rawURL := os.Getenv("TEST_REDIS_URL")
	if rawURL == "" {
		t.Skip("TEST_REDIS_URL is not configured")
	}
	limiter, err := NewRedisLimiterFromURL(rawURL, fmt.Sprintf("unsolero:test:%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()
	ctx := context.Background()
	if decision, err := limiter.Allow(ctx, "registration:opaque", 1, time.Second); err != nil || !decision.Allowed {
		t.Fatalf("first registration = %#v, %v", decision, err)
	}
	if decision, err := limiter.Allow(ctx, "registration:opaque", 1, time.Second); err != nil || decision.Allowed {
		t.Fatalf("duplicate registration = %#v, %v", decision, err)
	}
	if decision, err := limiter.Allow(ctx, "password-reset:opaque", 1, time.Second); err != nil || !decision.Allowed {
		t.Fatalf("separate route = %#v, %v", decision, err)
	}
	time.Sleep(1100 * time.Millisecond)
	if decision, err := limiter.Allow(ctx, "registration:opaque", 1, time.Second); err != nil || !decision.Allowed {
		t.Fatalf("after TTL = %#v, %v", decision, err)
	}
}

func TestRedisLimiterFailsClosedWhenBackendUnavailable(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:1", DialTimeout: 20 * time.Millisecond,
		ReadTimeout: 20 * time.Millisecond, WriteTimeout: 20 * time.Millisecond,
		MaxRetries: 0,
	})
	limiter, err := NewRedisLimiter(client, "unsolero:test:unavailable")
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	decision, err := limiter.Allow(ctx, "authentication:opaque", 1, time.Minute)
	if !errors.Is(err, ErrBackendUnavailable) || decision.Allowed {
		t.Fatalf("decision = %#v, error = %v", decision, err)
	}
}
