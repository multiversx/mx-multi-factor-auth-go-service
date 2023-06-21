package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core"
)

const (
	minLimitPeriodInSec      = 1
	minMaxFailures           = 1
	minOperationTimeoutInSec = 1
)

// ErrNilRedisClient signals that a nil redis client component has been provided
var ErrNilRedisClient = errors.New("nil redis client")

// RateLimiterResult defines rate limiter result
type RateLimiterResult struct {
	// Allowed specifies if the request was allowed, 1 if allowed
	// and 0 if not allowed
	Allowed int

	// Remaining is the maximum number of requests that could be
	// permitted instantaneously for this key given the current
	// state. For example, if a rate limiter allows 10 requests per
	// second and has already received 6 requests for this key this
	// second, Remaining would be 4.
	Remaining int

	// ResetAfter is the time until the expiration time is reached
	// for a given key. For example, if a rate limiter
	// manages requests per minute and received one request 20s ago,
	// reset after would return 40s
	ResetAfter time.Duration
}

// ArgsRateLimiter defines the arguments needed for creating a rate limiter component
type ArgsRateLimiter struct {
	OperationTimeoutInSec uint64
	MaxFailures           uint64
	LimitPeriodInSec      uint64
	Storer                RedisClient
}

type rateLimiter struct {
	operationTimeout time.Duration
	maxFailures      uint64
	limitPeriod      time.Duration
	storer           RedisClient
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
		storer:           args.Storer,
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
	if args.Storer == nil {
		return ErrNilRedisClient
	}

	return nil
}

// CheckAllowed will check if rate limits for the specified key
// It will return number of remaining trials
func (rl *rateLimiter) CheckAllowed(key string) (*RateLimiterResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rl.operationTimeout)
	defer cancel()

	res, err := rl.rateLimit(ctx, key)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (rl *rateLimiter) rateLimit(ctx context.Context, key string) (*RateLimiterResult, error) {
	initRemaining := int64(rl.maxFailures - 1)
	allowed := 1

	firstTry, err := rl.storer.SetEntryIfNotExisting(ctx, key, initRemaining, rl.limitPeriod)
	if err != nil {
		return nil, err
	}

	if firstTry {
		return &RateLimiterResult{
			Allowed:    1,
			Remaining:  int(initRemaining),
			ResetAfter: rl.limitPeriod,
		}, nil
	}

	remaining, err := rl.storer.Decrement(ctx, key)
	if err != nil {
		return nil, err
	}

	expTime, err := rl.storer.ExpireTime(ctx, key)
	if err != nil {
		return nil, err
	}

	if remaining < 0 {
		remaining = 0
		allowed = 0
	}

	return &RateLimiterResult{
		Allowed:    allowed,
		Remaining:  int(remaining),
		ResetAfter: expTime,
	}, nil
}

// Reset will reset the rate limits for the provided key
func (rl *rateLimiter) Reset(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rl.operationTimeout)
	defer cancel()

	err := rl.storer.Delete(ctx, key)
	if err != nil {
		log.Error("Delete", "key", key, "err", err.Error())
	}

	return err
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
