package abuse

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisNamespacePattern = regexp.MustCompile(`^[a-zA-Z0-9:_-]{1,64}$`)

const incrementWindowScript = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
return {count, ttl}
`

// RedisLimiter provides one atomic fixed-window counter shared by every API
// replica. Callers supply only already-pseudonymized bounded keys.
type RedisLimiter struct {
	client    redis.UniversalClient
	namespace string
	script    *redis.Script
}

func NewRedisLimiter(client redis.UniversalClient, namespace string) (*RedisLimiter, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	if !redisNamespacePattern.MatchString(namespace) {
		return nil, errors.New("redis rate-limit namespace is invalid")
	}
	return &RedisLimiter{client: client, namespace: namespace, script: redis.NewScript(incrementWindowScript)}, nil
}

func NewRedisLimiterFromURL(rawURL, namespace string) (*RedisLimiter, error) {
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}
	options.MaxRetries = 1
	options.DialTimeout = 2 * time.Second
	options.ReadTimeout = 2 * time.Second
	options.WriteTimeout = 2 * time.Second
	client := redis.NewClient(options)
	limiter, err := NewRedisLimiter(client, namespace)
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	return limiter, nil
}

func (limiter *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (Decision, error) {
	if len(key) < 3 || len(key) > 160 || limit < 1 || window < time.Second || window > 24*time.Hour {
		return Decision{}, errors.New("invalid rate-limit operation")
	}
	result, err := limiter.script.Run(ctx, limiter.client, []string{limiter.namespace + ":" + key}, window.Milliseconds()).Slice()
	if err != nil {
		return Decision{}, fmt.Errorf("%w: %v", ErrBackendUnavailable, err)
	}
	if len(result) != 2 {
		return Decision{}, ErrBackendUnavailable
	}
	count, countOK := result[0].(int64)
	ttlMS, ttlOK := result[1].(int64)
	if !countOK || !ttlOK || ttlMS < 0 {
		return Decision{}, ErrBackendUnavailable
	}
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}
	return Decision{
		Allowed:    count <= int64(limit),
		Remaining:  remaining,
		RetryAfter: time.Duration(ttlMS) * time.Millisecond,
	}, nil
}

func (limiter *RedisLimiter) Ready(ctx context.Context) error {
	if err := limiter.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrBackendUnavailable, err)
	}
	return nil
}

func (limiter *RedisLimiter) Close() error { return limiter.client.Close() }

func (*RedisLimiter) BackendName() string { return "redis" }
