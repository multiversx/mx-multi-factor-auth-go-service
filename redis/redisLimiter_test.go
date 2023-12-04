package redis_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected err")

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

	t.Run("returns err on storer increment fail", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		redisClient := &testscommon.RedisClientStub{
			IncrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 0, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowedAndIncreaseTrials("key")
		require.Equal(t, expectedErr, err)
		require.Nil(t, res)
	})

	t.Run("returns err on storer set expire if not exists fail", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		redisClient := &testscommon.RedisClientStub{
			SetExpireIfNotExistingCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				return false, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowedAndIncreaseTrials("key")
		require.Equal(t, expectedErr, err)
		require.Nil(t, res)
	})

	t.Run("returns err on storer get expire fail", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		redisClient := &testscommon.RedisClientStub{
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return 0, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowedAndIncreaseTrials("key")
		require.Equal(t, expectedErr, err)
		require.Nil(t, res)
	})

	t.Run("should work on first try, when key was not previously set", func(t *testing.T) {
		t.Parallel()

		maxFailures := 3
		maxDuration := 10

		expRemaining := 2

		args := createMockRateLimiterArgs()
		args.MaxFailures = int64(maxFailures)
		args.LimitPeriodInSec = uint64(maxDuration)

		redisClient := &testscommon.RedisClientStub{
			IncrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 1, nil
			},
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				require.Fail(t, "should have not been called")
				return 0, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowedAndIncreaseTrials("key")
		require.Nil(t, err)

		require.Equal(t, true, res.Allowed)
		require.Equal(t, expRemaining, res.Remaining)
	})

	t.Run("should work on second try", func(t *testing.T) {
		t.Parallel()

		maxFailures := 3
		maxDuration := 10

		expRemaining := 1
		expRetryAfter := time.Second * time.Duration(6)

		args := createMockRateLimiterArgs()
		args.MaxFailures = int64(maxFailures)
		args.LimitPeriodInSec = uint64(maxDuration)

		wasSetExpireIfNotExistingCalled := false
		wasExpireTimeCalled := false
		redisClient := &testscommon.RedisClientStub{
			IncrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 2, nil
			},
			SetExpireIfNotExistingCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				wasSetExpireIfNotExistingCalled = true
				return true, nil
			},
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				wasExpireTimeCalled = true
				return expRetryAfter, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowedAndIncreaseTrials("key")
		require.Nil(t, err)

		require.Equal(t, true, res.Allowed)
		require.Equal(t, expRemaining, res.Remaining)
		require.Equal(t, expRetryAfter, res.ResetAfter)
		require.True(t, wasExpireTimeCalled)
		require.True(t, wasSetExpireIfNotExistingCalled)
	})

	t.Run("should block on exceeded trials, on forth try", func(t *testing.T) {
		t.Parallel()

		maxFailures := 3
		maxDuration := 9

		expRetryAfter := time.Second * time.Duration(9)

		args := createMockRateLimiterArgs()
		args.MaxFailures = int64(maxFailures)
		args.LimitPeriodInSec = uint64(maxDuration)

		redisClient := &testscommon.RedisClientStub{
			IncrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 4, nil
			},
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return expRetryAfter, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		res, err := rl.CheckAllowedAndIncreaseTrials("key")
		require.Nil(t, err)

		require.Equal(t, false, res.Allowed)
		require.Equal(t, 0, res.Remaining)
		require.Equal(t, expRetryAfter, res.ResetAfter)
	})
}

func TestReset(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()

		redisClient := &testscommon.RedisClientStub{
			DeleteCalled: func(ctx context.Context, key string) error {
				return expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.Reset("key")
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
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
	})
}

func TestConcurrentOperationsShouldWork(t *testing.T) {
	t.Parallel()

	type cacheEntry struct {
		value int64
		ttl   time.Duration
	}
	localCache := make(map[string]*cacheEntry, 0)
	mutLocalCache := sync.RWMutex{}

	args := createMockRateLimiterArgs()
	args.Storer = &testscommon.RedisClientStub{
		IncrementCalled: func(ctx context.Context, key string) (int64, error) {
			mutLocalCache.Lock()
			defer mutLocalCache.Unlock()
			entry, has := localCache[key]
			if !has {
				entry = &cacheEntry{
					value: 0,
					ttl:   0,
				}
				localCache[key] = entry
			}

			entry.value++

			return 1, nil
		},
		SetExpireIfNotExistingCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			mutLocalCache.Lock()
			defer mutLocalCache.Unlock()
			entry, has := localCache[key]
			if has {
				if entry.ttl != 0 {
					return false, nil
				}

				entry.ttl = ttl
				return true, nil
			}

			return false, errors.New("missing key")
		},
		ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
			mutLocalCache.RLock()
			entry, has := localCache[key]
			mutLocalCache.RUnlock()
			if !has {
				assert.Fail(t, "this should never happen")
				return 0, errors.New("missing key")
			}

			return entry.ttl, nil
		},
		DeleteCalled: func(ctx context.Context, key string) error {
			mutLocalCache.Lock()
			delete(localCache, key)
			mutLocalCache.Unlock()

			return nil
		},
	}
	rl, err := redis.NewRateLimiter(args)
	assert.NoError(t, err)

	testKey := "test:key"

	numCalls := 100
	wg := sync.WaitGroup{}
	wg.Add(numCalls)

	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			switch idx {
			case 0:
				_, checkAllowedAndIncreaseTrialsErr := rl.CheckAllowedAndIncreaseTrials(testKey)
				assert.NoError(t, checkAllowedAndIncreaseTrialsErr)
			case 1:
				resetErr := rl.Reset(testKey)
				assert.NoError(t, resetErr)
			default:
				assert.Fail(t, "should have not been called")
			}

			wg.Done()
		}(i % 2)
	}

	wg.Wait()
}
