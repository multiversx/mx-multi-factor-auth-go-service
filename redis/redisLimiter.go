package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/multiversx/multi-factor-auth-go-service/core"
)

const (
	minLimitPeriodInSec      = 1
	minMaxFailures           = 1
	minOperationTimeoutInSec = 1
)

// ErrNilRedisLimiter signals that a nil redis limiter component has been provided
var ErrNilRedisLimiter = errors.New("nil redis limiter")

// ArgsRateLimiter defines the arguments needed for creating a rate limiter component
type ArgsRateLimiter struct {
	OperationTimeoutInSec uint64
	MaxFailures           uint64
	LimitPeriodInSec      uint64
	Limiter               RedisLimiter
}

// RateLimiterResult defines an alias for redis rate limiter result
type RateLimiterResult = redis_rate.Result

type rateLimiter struct {
	operationTimeout time.Duration
	maxFailures      uint64
	limitPeriod      time.Duration
	limiter          RedisLimiter
}

// NewRateLimiter will create a new instance of rate limiter
func NewRateLimiter(args ArgsRateLimiter) (*rateLimiter, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &rateLimiter{
		operationTimeout: time.Duration(args.OperationTimeoutInSec) * time.Second,
		maxFailures:      args.MaxFailures,
		limitPeriod:      time.Duration(args.LimitPeriodInSec) * time.Second,
		limiter:          args.Limiter,
	}, nil
}

func checkArgs(args ArgsRateLimiter) error {
	if args.OperationTimeoutInSec < minOperationTimeoutInSec {
		return fmt.Errorf("%w for OperationTimeoutInSec, received %d, min expected %d", core.ErrInvalidValue, args.OperationTimeoutInSec, minOperationTimeoutInSec)
	}
	if args.LimitPeriodInSec < minLimitPeriodInSec {
		return fmt.Errorf("%w for LimitPeriodInSec, received %d, min expected %d", core.ErrInvalidValue, args.LimitPeriodInSec, minLimitPeriodInSec)
	}
	if args.MaxFailures < minMaxFailures {
		return fmt.Errorf("%w for MaxFailures, received %d, min expected %d", core.ErrInvalidValue, args.MaxFailures, minMaxFailures)
	}
	if args.Limiter == nil {
		return ErrNilRedisLimiter
	}

	return nil
}

// CheckAllowed will check if rate limits for the specified key
// It will return number of remaining trials
func (rl *rateLimiter) CheckAllowed(key string) (*RateLimiterResult, error) {
	limit := redis_rate.Limit{
		Rate:   int(rl.maxFailures),
		Period: rl.limitPeriod,
		Burst:  int(rl.maxFailures + 1),
	}

	ctx, cancel := context.WithTimeout(context.Background(), rl.operationTimeout)
	defer cancel()

	return rl.limiter.Allow(ctx, key, limit)
}

// Reset will reset the rate limits for the provided key
func (rl *rateLimiter) Reset(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rl.operationTimeout)
	defer cancel()

	return rl.limiter.Reset(ctx, key)
}

// Period will return the limit period duration for the limiter
func (rl *rateLimiter) Period() time.Duration {
	return rl.limitPeriod
}

// Rate will return the number of trials for the limiter
func (rl *rateLimiter) Rate() int {
	return int(rl.maxFailures)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rl *rateLimiter) IsInterfaceNil() bool {
	return rl == nil
}
