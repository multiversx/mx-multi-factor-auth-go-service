package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
)

// RateLimiter defines the behaviour of a rate limiter component
type RateLimiter interface {
	CheckAllowed(key string) (*RateLimiterResult, error)
	Reset(key string) error
	Period() time.Duration
	IsInterfaceNil() bool
}

// RedisLimiter defines the behaviour of the redis rate limiter component
type RedisLimiter interface {
	Allow(ctx context.Context, key string, limit redis_rate.Limit) (*redis_rate.Result, error)
	Reset(ctx context.Context, key string) error
}
