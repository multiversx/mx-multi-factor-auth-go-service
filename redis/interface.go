package redis

import (
	"time"
)

// RateLimiter defines the behaviour of a rate limiter component
type RateLimiter interface {
	CheckAllowed(key string, maxFailures int, maxDuration time.Duration) (int, error)
	Reset(key string) error
	IsInterfaceNil() bool
}
