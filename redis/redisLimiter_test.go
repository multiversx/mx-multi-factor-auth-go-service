package redis_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon"
)

var expectedErr = errors.New("expected err")

func createMockRateLimiterArgs() redis.ArgsRateLimiter {
	return redis.ArgsRateLimiter{
		OperationTimeoutInSec:   10,
		MaxFailures:             3,
		LimitPeriodInSec:        60,
		SecurityModeMaxFailures: 100,
		SecurityModeLimitPeriod: 86400,
		Storer:                  &testscommon.RedisClientStub{},
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

	t.Run("invalid security mode max failures", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.SecurityModeMaxFailures = 0

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
	})

	t.Run("invalid security mode limit period in seconds", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.SecurityModeLimitPeriod = 0

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

		require.Equal(t, time.Duration(args.LimitPeriodInSec)*time.Second, rl.Period(redis.NormalMode))
		require.Equal(t, time.Duration(args.SecurityModeLimitPeriod)*time.Second, rl.Period(redis.SecurityMode))
		require.Equal(t, int(args.MaxFailures), rl.Rate(redis.NormalMode))
		require.Equal(t, int(args.SecurityModeMaxFailures), rl.Rate(redis.SecurityMode))
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

		modes := []redis.Mode{redis.NormalMode, redis.SecurityMode}
		for _, mode := range modes {
			res, err := rl.CheckAllowedAndIncreaseTrials("key", mode)
			require.Equal(t, expectedErr, err)
			require.Nil(t, res)
		}
	})

	t.Run("returns err on storer set expire fail", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		redisClient := &testscommon.RedisClientStub{
			IncrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 1, nil // first try
			},
			SetExpireCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				return false, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		modes := []redis.Mode{redis.NormalMode, redis.SecurityMode}
		for _, mode := range modes {
			res, err := rl.CheckAllowedAndIncreaseTrials("key", mode)
			require.Equal(t, expectedErr, err)
			require.Nil(t, res)
		}
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

		modes := []redis.Mode{redis.NormalMode, redis.SecurityMode}
		for _, mode := range modes {
			res, err := rl.CheckAllowedAndIncreaseTrials("key", mode)
			require.Equal(t, expectedErr, err)
			require.Nil(t, res)
		}
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

		testData := []struct {
			mode              redis.Mode
			expectedRemaining int
		}{
			{redis.NormalMode, expRemaining},
			{redis.SecurityMode, int(args.SecurityModeMaxFailures) - 1},
		}
		for _, data := range testData {
			res, err := rl.CheckAllowedAndIncreaseTrials("key", data.mode)
			require.Nil(t, err)

			require.Equal(t, true, res.Allowed)
			require.Equal(t, data.expectedRemaining, res.Remaining)
		}
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

		wasSetExpireCalled := false
		wasExpireTimeCalled := false
		wasSetExpireIfNotExistsCalled := false
		redisClient := &testscommon.RedisClientStub{
			IncrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 2, nil
			},
			SetExpireCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				wasSetExpireCalled = true
				return true, nil
			},
			SetExpireIfNotExistsCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				wasSetExpireIfNotExistsCalled = true
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

		res, err := rl.CheckAllowedAndIncreaseTrials("key", redis.NormalMode)
		require.Nil(t, err)

		require.Equal(t, true, res.Allowed)
		require.Equal(t, expRemaining, res.Remaining)
		require.Equal(t, expRetryAfter, res.ResetAfter)
		require.True(t, wasExpireTimeCalled)
		require.False(t, wasSetExpireCalled)
		require.True(t, wasSetExpireIfNotExistsCalled)
	})

	t.Run("should block on exceeded trials", func(t *testing.T) {
		t.Parallel()

		maxFailures := 3
		maxDuration := 9
		securityModeMaxFailures := 100
		securityModeMaxDuration := 86400

		args := createMockRateLimiterArgs()
		args.MaxFailures = int64(maxFailures)
		args.LimitPeriodInSec = uint64(maxDuration)
		args.SecurityModeMaxFailures = int64(securityModeMaxFailures)
		args.SecurityModeLimitPeriod = uint64(securityModeMaxDuration)
		normalModeKey := "key"
		secureModeKey := "secureModeKey"

		testData := []struct {
			mode               redis.Mode
			key                string
			expectedRemaining  int
			expectedResetAfter time.Duration
		}{
			{redis.NormalMode, normalModeKey, 0, time.Second * time.Duration(args.LimitPeriodInSec)},
			{redis.SecurityMode, secureModeKey, 0, time.Second * time.Duration(args.SecurityModeLimitPeriod)},
		}

		unexpectedErr := errors.New("unexpected error")
		redisClient := &testscommon.RedisClientStub{
			IncrementCalled: func(ctx context.Context, key string) (int64, error) {
				switch key {
				case normalModeKey:
					return int64(maxFailures) + 1, nil
				case secureModeKey:
					return int64(securityModeMaxFailures) + 1, nil
				}

				return 0, unexpectedErr
			},
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				switch key {
				case normalModeKey:
					return time.Second * time.Duration(args.LimitPeriodInSec), nil
				case secureModeKey:
					return time.Second * time.Duration(args.SecurityModeLimitPeriod), nil
				}
				return 0, unexpectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		for _, data := range testData {
			res, err := rl.CheckAllowedAndIncreaseTrials(data.key, data.mode)
			require.Nil(t, err)

			require.Equal(t, false, res.Allowed)
			require.Equal(t, data.expectedRemaining, res.Remaining)
			require.Equal(t, data.expectedResetAfter, res.ResetAfter)
		}
	})
}

func TestReset(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()

		redisClient := &testscommon.RedisClientStub{
			ResetCounterAndKeepTTLCalled: func(ctx context.Context, key string) error {
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
			ResetCounterAndKeepTTLCalled: func(ctx context.Context, key string) error {
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

func TestDecrementSecurityFailedTrials(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()

		redisClient := &testscommon.RedisClientStub{
			DecrementCalled: func(ctx context.Context, key string) (int64, error) {
				return 0, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.DecrementSecurityFailedTrials("key")
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()

		wasCalled := false
		redisClient := &testscommon.RedisClientStub{
			DecrementCalled: func(ctx context.Context, key string) (int64, error) {
				wasCalled = true
				return 0, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.DecrementSecurityFailedTrials("key")
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
		SetExpireCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
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
		ResetCounterAndKeepTTLCalled: func(ctx context.Context, key string) error {
			mutLocalCache.Lock()
			_, has := localCache[key]
			if has {
				localCache[key].value = 0
			}
			mutLocalCache.Unlock()

			return nil
		},
	}
	rl, err := redis.NewRateLimiter(args)
	assert.NoError(t, err)

	testKey := "test:key"
	testSecureModeKey := "test:secureModeKey"

	numCalls := 100
	wg := sync.WaitGroup{}
	wg.Add(numCalls)

	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			switch idx {
			case 0:
				_, checkAllowedAndIncreaseTrialsErr := rl.CheckAllowedAndIncreaseTrials(testKey, redis.NormalMode)
				assert.NoError(t, checkAllowedAndIncreaseTrialsErr)
			case 1:
				_, checkAllowedAndIncreaseTrialsErr := rl.CheckAllowedAndIncreaseTrials(testSecureModeKey, redis.SecurityMode)
				assert.NoError(t, checkAllowedAndIncreaseTrialsErr)
			case 2:
				decrementError := rl.DecrementSecurityFailedTrials(testSecureModeKey)
				assert.NoError(t, decrementError)
			case 3:
				resetErr := rl.Reset(testKey)
				assert.NoError(t, resetErr)
			default:
				assert.Fail(t, "should have not been called")
			}

			wg.Done()
		}(i % 4)
	}

	wg.Wait()
}
