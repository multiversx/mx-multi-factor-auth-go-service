package redis

import (
	"context"
	"time"
)

// RateLimiter defines the behaviour of a rate limiter component
type RateLimiter interface {
	CheckAllowedAndIncreaseTrials(key string) (*RateLimiterResult, error)
	Reset(key string) error
	Period() time.Duration
	Rate() int
	IsInterfaceNil() bool
}

// RedisStorer defines the behaviour of a redis storer component
type RedisStorer interface {
	Increment(ctx context.Context, key string) (int64, error)
	SetExpireIfNotExisting(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Delete(ctx context.Context, key string) error
	ExpireTime(ctx context.Context, key string) (time.Duration, error)
	IsConnected(ctx context.Context) bool
	IsInterfaceNil() bool
}
