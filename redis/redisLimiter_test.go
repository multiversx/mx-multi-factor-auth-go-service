package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
)

func createMockRateLimiterArgs() redis.ArgsRateLimiter {
	return redis.ArgsRateLimiter{
		OperationTimeoutInSec: 10,
		MaxFailures:           3,
		LimitPeriodInSec:      60,
		Limiter:               &testscommon.RedisLimiterStub{},
	}
}

func TestNewRateLimiter(t *testing.T) {
	t.Parallel()

	t.Run("invalid max failures", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.MaxFailures = 0

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
	})

	t.Run("invalid limit period in seconds", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.LimitPeriodInSec = 0

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
	})

	t.Run("invalid operation timeout in seconds", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.OperationTimeoutInSec = 0

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
	})

	t.Run("nil redis rate limiter", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.Limiter = nil

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, redis.ErrNilRedisLimiter))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()

		rl, err := redis.NewRateLimiter(args)
		require.NotNil(t, rl)
		require.False(t, rl.IsInterfaceNil())
		require.Nil(t, err)

		require.Equal(t, time.Duration(args.LimitPeriodInSec)*time.Second, rl.Period())
	})
}

func TestCheckAllowed(t *testing.T) {
	t.Parallel()

	t.Run("returns err on limiter allow fail", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockRateLimiterArgs()
		redisLimiter := &testscommon.RedisLimiterStub{
			AllowCalled: func(ctx context.Context, key string, limit redis_rate.Limit) (*redis_rate.Result, error) {
				return nil, expectedErr
			},
		}
		args.Limiter = redisLimiter

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowed("key")
		require.Equal(t, expectedErr, err)
		require.Nil(t, res)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		maxFailures := 4
		maxDuration := 10

		expRemaining := 3

		args := createMockRateLimiterArgs()
		args.MaxFailures = uint64(maxFailures)
		args.LimitPeriodInSec = uint64(maxDuration)
		redisLimiter := &testscommon.RedisLimiterStub{
			AllowCalled: func(ctx context.Context, key string, limit redis_rate.Limit) (*redis_rate.Result, error) {
				require.Equal(t, maxFailures, limit.Rate)
				require.Equal(t, maxFailures, limit.Burst)
				require.Equal(t, time.Duration(maxDuration)*time.Second, limit.Period)

				return &redis_rate.Result{
					Remaining: expRemaining,
				}, nil
			},
		}
		args.Limiter = redisLimiter

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowed("key")
		require.Nil(t, err)

		require.Equal(t, expRemaining, res.Remaining)
	})
}

func TestReset(t *testing.T) {
	t.Parallel()

	args := createMockRateLimiterArgs()

	wasCalled := false
	redisLimiter := &testscommon.RedisLimiterStub{
		ResetCalled: func(ctx context.Context, key string) error {
			wasCalled = true
			return nil
		},
	}
	args.Limiter = redisLimiter

	rl, err := redis.NewRateLimiter(args)
	require.Nil(t, err)

	err = rl.Reset("key")
	require.Nil(t, err)
	require.True(t, wasCalled)
}
