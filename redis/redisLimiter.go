package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// ErrNilRedisClient signals that a nil redis client component has been provided
var ErrNilRedisClient = errors.New("nil redis client")

type rateLimiter struct {
	limiter *redis_rate.Limiter
	ctx     context.Context
}

// NewRateLimiter will create a new instance of rate limiter
func NewRateLimiter(client *redis.Client) (*rateLimiter, error) {
	if client == nil {
		return nil, ErrNilRedisClient
	}

	ctx := context.Background()
	limiter := redis_rate.NewLimiter(client)

	return &rateLimiter{
		limiter: limiter,
		ctx:     ctx,
	}, nil
}

// CheckAllowed will check if rate limits for the specified key
// It will return number of remaining trials
func (rl *rateLimiter) CheckAllowed(key string, maxFailures int, maxDuration time.Duration) (int, error) {
	limit := redis_rate.Limit{
		Rate:   maxFailures,
		Period: maxDuration,
		Burst:  maxFailures,
	}

	res, err := rl.limiter.Allow(rl.ctx, key, limit)
	if err != nil {
		return 0, err
	}

	return res.Remaining, nil
}

// Reset will reset the rate limits for the provided key
func (rl *rateLimiter) Reset(key string) error {
	return rl.limiter.Reset(rl.ctx, key)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rl *rateLimiter) IsInterfaceNil() bool {
	return rl == nil
}
