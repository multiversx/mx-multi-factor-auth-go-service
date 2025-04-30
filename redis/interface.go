package redis

import (
	"context"
	"time"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
)

type Mode int

const (
	// NormalMode is the mode for normal
	NormalMode Mode = iota
	// SecurityMode is the mode for security
	SecurityMode
)

// RateLimiter defines the behaviour of a rate limiter component
type RateLimiter interface {
	CheckAllowedAndIncreaseTrials(key string, mode Mode) (*RateLimiterResult, error)
	Reset(key string) error
	SetSecurityModeNoExpire(key string) error
	UnsetSecurityModeNoExpire(key string) error
	GetSecurityStatus(key string) core.Status
	DecrementSecurityFailedTrials(key string) error
	Period(mode Mode) time.Duration
	Rate(mode Mode) int
	ExtendSecurityMode(key string) error
	IsInterfaceNil() bool
}

// RedisStorer defines the behaviour of a redis storer component
type RedisStorer interface {
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	SetExpire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	SetExpireIfNotExists(ctx context.Context, key string, ttl time.Duration) (bool, error)
	SetPersist(ctx context.Context, key string) (bool, error)
	SetGreaterExpireTTL(ctx context.Context, key string, ttl time.Duration) (bool, error)
	ResetCounterAndKeepTTL(ctx context.Context, key string) error
	ExpireTime(ctx context.Context, key string) (time.Duration, error)
	IsConnected(ctx context.Context) bool
	IsInterfaceNil() bool
}
