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

// Locker defines the behaviour of a locker component which is able to create new distributed mutexes
type Locker interface {
	NewMutex(name string) Mutex
	IsInterfaceNil() bool
}

// RedLockMutex defines the behaviour of a redlock mutex component
type RedLockMutex interface {
	Lock() error
	LockContext(ctx context.Context) error
	Unlock() (bool, error)
	UnlockContext(ctx context.Context) (bool, error)
}

// Mutex defines the behaviour of a distributed mutex component
type Mutex interface {
	Lock()
	LockContext(ctx context.Context)
	Unlock()
	UnlockContext(ctx context.Context)
}
