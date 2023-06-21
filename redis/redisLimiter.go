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

// ErrNilRedisLimiter signals that a nil redis limiter component has been provided
var ErrNilRedisLimiter = errors.New("nil redis limiter")

// ArgsRateLimiter defines the arguments needed for creating a rate limiter component
type ArgsRateLimiter struct {
	OperationTimeoutInSec uint64
	MaxFailures           uint64
	LimitPeriodInSec      uint64
	Storer                RedisClient
}

// RateLimiterResult defines redis rate limiter result
type RateLimiterResult struct {
	// Allowed is the number of events that may happen at time now.
	Allowed int

	// Remaining is the maximum number of requests that could be
	// permitted instantaneously for this key given the current
	// state. For example, if a rate limiter allows 10 requests per
	// second and has already received 6 requests for this key this
	// second, Remaining would be 4.
	Remaining int

	// RetryAfter is the time until the next request will be permitted.
	// It should be -1 unless the rate limit has been exceeded.
	RetryAfter time.Duration

	// ResetAfter is the time until the RateLimiter returns to its
	// initial state for a given key. For example, if a rate limiter
	// manages requests per second and received one request 200ms ago,
	// Reset would return 800ms. You can also think of this as the time
	// until Limit and Remaining will be equal.
	ResetAfter time.Duration
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
		return ErrNilRedisLimiter
	}

	return nil
}

// CheckAllowed will check if rate limits for the specified key
// It will return number of remaining trials
func (rl *rateLimiter) CheckAllowed(key string) (*RateLimiterResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rl.operationTimeout)
	defer cancel()

	res, err := rl.rateLimitV2(ctx, key)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (rl *rateLimiter) rateLimitV2(ctx context.Context, key string) (*RateLimiterResult, error) {
	wasSet, err := rl.storer.SetEntryWithTTL(ctx, key, int64(rl.maxFailures-1), rl.limitPeriod)
	if err != nil {
		return nil, err
	}

	if wasSet {
		return &RateLimiterResult{
			Allowed:    1,
			Remaining:  int(rl.maxFailures - 1),
			RetryAfter: -1,
			ResetAfter: rl.limitPeriod,
		}, nil
	}

	index, err := rl.storer.Decrement(ctx, key)
	if err != nil {
		return nil, err
	}

	expTime, err := rl.storer.ExpireTime(ctx, key)
	if err != nil {
		return nil, err
	}

	if index < 0 {
		return &RateLimiterResult{
			Allowed:    0,
			Remaining:  0,
			RetryAfter: -1,
			ResetAfter: expTime,
		}, nil
	}

	return &RateLimiterResult{
		Allowed:    1,
		Remaining:  int(index),
		RetryAfter: -1,
		ResetAfter: expTime,
	}, nil
}

func (rl *rateLimiter) rateLimit(ctx context.Context, key string) (*RateLimiterResult, error) {
	tatVal, now, err := rl.storer.GetWithTime(ctx, key)
	if err != nil {
		return nil, err
	}

	var tat time.Time
	if tatVal == -1 {
		tat = now
	} else {
		tat = time.Unix(0, tatVal)
	}

	if now.After(tat) {
		tat = now
	}

	period := int(rl.limitPeriod.Seconds()) / int(rl.maxFailures)

	increment := period
	burstOffset := period * int(rl.maxFailures)

	newTat := tat.Add(time.Second * time.Duration(increment))
	allowAt := newTat.Add(-time.Second * time.Duration(burstOffset))

	diff := now.Sub(allowAt)

	remaining := int(diff.Seconds()) / period
	if diff.Seconds() < 0 {
		remaining = -1
	}

	if remaining < 0 {
		resetAfter := tat.Sub(now)
		retryAfter := diff

		// _, err := rl.storer.SetEntryWithTTL(ctx, key, int64(newTat.UnixNano()), resetAfter)
		// if err != nil {
		// 	return nil, err
		// }

		return &RateLimiterResult{
			Allowed:    0,
			Remaining:  0,
			RetryAfter: retryAfter,
			ResetAfter: resetAfter,
		}, nil
	}

	resetAfter := newTat.Sub(now)
	if resetAfter.Seconds() > 0 {
		_, err := rl.storer.SetEntryWithTTL(ctx, key, int64(newTat.UnixNano()), resetAfter)
		if err != nil {
			return nil, err
		}
	}

	return &RateLimiterResult{
		Allowed:    1,
		Remaining:  remaining,
		RetryAfter: -1,
		ResetAfter: resetAfter,
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
