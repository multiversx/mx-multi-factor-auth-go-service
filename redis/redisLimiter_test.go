package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
		Storer:                &testscommon.RedisClientStub{},
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
		args.Storer = nil

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, redis.ErrNilRedisClientWrapper))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()

		rl, err := redis.NewRateLimiter(args)
		require.NotNil(t, rl)
		require.False(t, rl.IsInterfaceNil())
		require.Nil(t, err)

		require.Equal(t, time.Duration(args.LimitPeriodInSec)*time.Second, rl.Period())
		require.Equal(t, int(args.MaxFailures), rl.Rate())
	})
}

func TestCheckAllowed(t *testing.T) {
	t.Parallel()

	t.Run("returns err on storer set entry fail", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockRateLimiterArgs()
		redisClient := &testscommon.RedisClientStub{
			SetEntryIfNotExistingCalled: func(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
				return false, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowed("key")
		require.Equal(t, expectedErr, err)
		require.Nil(t, res)
	})

	t.Run("returns err on storer decrement fail", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockRateLimiterArgs()
		redisClient := &testscommon.RedisClientStub{
			SetEntryIfNotExistingCalled: func(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
				return false, nil
			},
			DecrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 1, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowed("key")
		require.Equal(t, expectedErr, err)
		require.Nil(t, res)
	})

	t.Run("returns err on storer decrement fail", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockRateLimiterArgs()
		redisClient := &testscommon.RedisClientStub{
			SetEntryIfNotExistingCalled: func(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
				return false, nil
			},
			DecrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 1, nil
			},
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return 0, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowed("key")
		require.Equal(t, expectedErr, err)
		require.Nil(t, res)
	})

	t.Run("should work on first try, when key was not previously set", func(t *testing.T) {
		t.Parallel()

		maxFailures := 3
		maxDuration := 10

		expRemaining := 2

		args := createMockRateLimiterArgs()
		args.MaxFailures = uint64(maxFailures)
		args.LimitPeriodInSec = uint64(maxDuration)

		redisClient := &testscommon.RedisClientStub{
			SetEntryIfNotExistingCalled: func(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
				return true, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowed("key")
		require.Nil(t, err)

		require.Equal(t, 1, res.Allowed)
		require.Equal(t, expRemaining, res.Remaining)
	})

	t.Run("should work on second try", func(t *testing.T) {
		t.Parallel()

		maxFailures := 3
		maxDuration := 10

		expRemaining := 1
		expRetryAfter := time.Second * time.Duration(6)

		args := createMockRateLimiterArgs()
		args.MaxFailures = uint64(maxFailures)
		args.LimitPeriodInSec = uint64(maxDuration)

		redisClient := &testscommon.RedisClientStub{
			SetEntryIfNotExistingCalled: func(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
				require.Equal(t, int64(maxFailures-1), value)
				return false, nil
			},
			DecrementCalled: func(ctx context.Context, key string) (int64, error) {
				return int64(maxFailures - 2), nil
			},
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return expRetryAfter, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowed("key")
		require.Nil(t, err)

		require.Equal(t, 1, res.Allowed)
		require.Equal(t, expRemaining, res.Remaining)
		require.Equal(t, expRetryAfter, res.ResetAfter)
	})

	t.Run("should block on exceeded trials, on forth try", func(t *testing.T) {
		t.Parallel()

		maxFailures := 3
		maxDuration := 9

		expRetryAfter := time.Second * time.Duration(9)

		args := createMockRateLimiterArgs()
		args.MaxFailures = uint64(maxFailures)
		args.LimitPeriodInSec = uint64(maxDuration)

		redisClient := &testscommon.RedisClientStub{
			SetEntryIfNotExistingCalled: func(ctx context.Context, key string, value int64, ttl time.Duration) (bool, error) {
				return false, nil
			},
			DecrementCalled: func(ctx context.Context, key string) (int64, error) {
				return int64(maxFailures - 4), nil
			},
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return expRetryAfter, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowed("key")
		require.Nil(t, err)

		require.Equal(t, 0, res.Allowed)
		require.Equal(t, 0, res.Remaining)
		require.Equal(t, expRetryAfter, res.ResetAfter)
	})
}

func TestReset(t *testing.T) {
	t.Parallel()

	args := createMockRateLimiterArgs()

	wasCalled := false
	redisClient := &testscommon.RedisClientStub{
		DeleteCalled: func(ctx context.Context, key string) error {
			wasCalled = true
			return nil
		},
	}
	args.Storer = redisClient

	rl, err := redis.NewRateLimiter(args)
	require.Nil(t, err)

	err = rl.Reset("key")
	require.Nil(t, err)
	require.True(t, wasCalled)
}
