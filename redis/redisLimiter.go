package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
)

const (
	minLimitPeriodInSec      = 1
	minMaxFailures           = 1
	minOperationTimeoutInSec = 1
)

// RateLimiterResult defines rate limiter result
type RateLimiterResult struct {
	// Allowed specifies if the request was allowed, 1 if allowed
	// and 0 if not allowed
	Allowed bool

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

// FailureConfig defines the configuration for the rate limiter failure configuration
type FailureConfig struct {
	MaxFailures      int64
	LimitPeriodInSec uint64
}

// ArgsRateLimiter defines the arguments needed for creating a rate limiter component
type ArgsRateLimiter struct {
	OperationTimeoutInSec     uint64
	FreezeFailureConfig       FailureConfig
	SecurityModeFailureConfig FailureConfig
	Storer                    RedisStorer
}

type failureConfig struct {
	maxFailures int64
	limitPeriod time.Duration
}

type rateLimiter struct {
	operationTimeout          time.Duration
	freezeFailureConfig       failureConfig
	securityModeFailureConfig failureConfig
	storer                    RedisStorer
	mutStorer                 sync.Mutex
}

// NewRateLimiter will create a new instance of rate limiter
func NewRateLimiter(args ArgsRateLimiter) (*rateLimiter, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &rateLimiter{
		operationTimeout: time.Duration(args.OperationTimeoutInSec) * time.Second,
		freezeFailureConfig: failureConfig{
			maxFailures: args.FreezeFailureConfig.MaxFailures,
			limitPeriod: time.Duration(args.FreezeFailureConfig.LimitPeriodInSec) * time.Second,
		},
		securityModeFailureConfig: failureConfig{
			maxFailures: args.SecurityModeFailureConfig.MaxFailures,
			limitPeriod: time.Duration(args.SecurityModeFailureConfig.LimitPeriodInSec) * time.Second,
		},
		storer: args.Storer,
	}, nil
}

func checkArgs(args ArgsRateLimiter) error {
	if args.OperationTimeoutInSec < minOperationTimeoutInSec {
		return fmt.Errorf("%w for OperationTimeoutInSec, received %d, min expected %d", core.ErrInvalidValue, args.OperationTimeoutInSec, minOperationTimeoutInSec)
	}
	if args.FreezeFailureConfig.LimitPeriodInSec < minLimitPeriodInSec {
		return fmt.Errorf("%w for LimitPeriodInSec, received %d, min expected %d", core.ErrInvalidValue, args.FreezeFailureConfig.LimitPeriodInSec, minLimitPeriodInSec)
	}
	if args.FreezeFailureConfig.MaxFailures < minMaxFailures {
		return fmt.Errorf("%w for FreezeMaxFailures, received %d, min expected %d", core.ErrInvalidValue, args.FreezeFailureConfig.MaxFailures, minMaxFailures)
	}

	if args.SecurityModeFailureConfig.MaxFailures < minMaxFailures {
		return fmt.Errorf("%w for SecurityModeMaxFailures, received %d, min expected %d", core.ErrInvalidValue, args.SecurityModeFailureConfig.MaxFailures, minMaxFailures)
	}
	if args.SecurityModeFailureConfig.LimitPeriodInSec < minLimitPeriodInSec {
		return fmt.Errorf("%w for SecurityModeLimitPeriod, received %d, min expected %d", core.ErrInvalidValue, args.SecurityModeFailureConfig.LimitPeriodInSec, minLimitPeriodInSec)
	}

	if check.IfNil(args.Storer) {
		return ErrNilRedisClientWrapper
	}

	return nil
}

// CheckAllowedAndIncreaseTrials will check the rate limits for the specified key, and it will increase the number of trials
// It will return number of remaining trials
func (rl *rateLimiter) CheckAllowedAndIncreaseTrials(key string, mode Mode) (*RateLimiterResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rl.operationTimeout)
	defer cancel()

	res, err := rl.rateLimit(ctx, key, mode)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (rl *rateLimiter) rateLimit(ctx context.Context, key string, mode Mode) (*RateLimiterResult, error) {
	rl.mutStorer.Lock()
	defer rl.mutStorer.Unlock()

	totalRetries, err := rl.storer.Increment(ctx, key)
	if err != nil {
		return nil, err
	}

	firstTry := totalRetries == 1
	if firstTry {
		return rl.setAndGetLimiterFirstTimeResult(ctx, key, mode)
	}

	err = rl.setExpireIfNotExistsInMode(ctx, key, mode)
	if err != nil {
		return nil, err
	}

	return rl.getLimiterResult(ctx, key, totalRetries, mode)
}

func (rl *rateLimiter) getLimiterResult(ctx context.Context, key string, totalRetries int64, mode Mode) (*RateLimiterResult, error) {
	allowed := true
	_, maxFailures := rl.getFailConfig(mode)
	remaining := maxFailures - totalRetries
	if totalRetries > maxFailures {
		remaining = 0
		allowed = false
	}

	expTime, err := rl.storer.ExpireTime(ctx, key)
	if err != nil {
		return nil, err
	}

	return &RateLimiterResult{
		Allowed:    allowed,
		Remaining:  int(remaining),
		ResetAfter: expTime,
	}, nil
}

func (rl *rateLimiter) setExpireIfNotExistsInMode(ctx context.Context, key string, mode Mode) error {
	limitPeriod, _ := rl.getFailConfig(mode)
	_, err := rl.storer.SetExpireIfNotExists(ctx, key, limitPeriod)
	return err
}

func (rl *rateLimiter) setAndGetLimiterFirstTimeResult(ctx context.Context, key string, mode Mode) (*RateLimiterResult, error) {
	limitPeriod, maxFailures := rl.getFailConfig(mode)

	_, err := rl.storer.SetExpire(ctx, key, limitPeriod)
	if err != nil {
		return nil, err
	}

	return &RateLimiterResult{
		Allowed:    true,
		Remaining:  int(maxFailures - 1),
		ResetAfter: limitPeriod,
	}, nil
}

func (rl *rateLimiter) getFailConfig(mode Mode) (time.Duration, int64) {
	if mode == SecurityMode {
		return rl.securityModeFailureConfig.limitPeriod, rl.securityModeFailureConfig.maxFailures
	}
	return rl.freezeFailureConfig.limitPeriod, rl.freezeFailureConfig.maxFailures
}

// Reset will reset the rate limits for the provided key
func (rl *rateLimiter) Reset(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rl.operationTimeout)
	defer cancel()

	rl.mutStorer.Lock()
	defer rl.mutStorer.Unlock()
	err := rl.storer.ResetCounterAndKeepTTL(ctx, key)
	if err != nil {
		log.Error("Delete", "key", key, "err", err.Error())
	}

	return err
}

// DecrementSecurityFailedTrials will decrement the number of security retrials for the specified key
func (rl *rateLimiter) DecrementSecurityFailedTrials(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), rl.operationTimeout)
	defer cancel()

	rl.mutStorer.Lock()
	defer rl.mutStorer.Unlock()

	_, err := rl.storer.Decrement(ctx, key)

	return err
}

// Period will return the limit period duration for the limiter
func (rl *rateLimiter) Period(mode Mode) time.Duration {
	if mode == SecurityMode {
		return rl.securityModeFailureConfig.limitPeriod
	}
	return rl.freezeFailureConfig.limitPeriod
}

// Rate will return the number of trials for the limiter
func (rl *rateLimiter) Rate(mode Mode) int {
	if mode == SecurityMode {
		return int(rl.securityModeFailureConfig.maxFailures)
	}
	return int(rl.freezeFailureConfig.maxFailures)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rl *rateLimiter) IsInterfaceNil() bool {
	return rl == nil
}
