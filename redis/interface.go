package redis

import (
	"context"
	"time"
)

// RateLimiter defines the behaviour of a rate limiter component
type RateLimiter interface {
	CheckAllowed(key string) (*RateLimiterResult, error)
	Reset(key string) error
	Period() time.Duration
	Rate() int
	IsInterfaceNil() bool
}

// RedisStorer defines the behaviour of a redis storer component
type RedisStorer interface {
	SetEntryIfNotExisting(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error)
	Delete(ctx context.Context, key string) error
	Decrement(ctx context.Context, key string) (int64, error)
	ExpireTime(ctx context.Context, key string) (time.Duration, error)
	IsInterfaceNil() bool
}
