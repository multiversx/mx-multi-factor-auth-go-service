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
		OperationTimeoutInSec: 10,
		FreezeFailureConfig: redis.FailureConfig{
			MaxFailures:      3,
			LimitPeriodInSec: 60,
		},
		SecurityModeFailureConfig: redis.FailureConfig{
			MaxFailures:      100,
			LimitPeriodInSec: 86400,
		},
		Storer: &testscommon.RedisClientStub{},
	}
}

func TestNewRateLimiter(t *testing.T) {
	t.Parallel()

	t.Run("invalid max failures fir freeze config", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.FreezeFailureConfig.MaxFailures = 0

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
	})

	t.Run("invalid limit period in seconds for freeze config", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.FreezeFailureConfig.LimitPeriodInSec = 0

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
	})

	t.Run("invalid security mode max failures", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.SecurityModeFailureConfig.MaxFailures = 0

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, rl)
		require.True(t, errors.Is(err, core.ErrInvalidValue))
	})

	t.Run("invalid security mode limit period in seconds", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()
		args.SecurityModeFailureConfig.LimitPeriodInSec = 0

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

		require.Equal(t, time.Duration(args.FreezeFailureConfig.LimitPeriodInSec)*time.Second, rl.Period(redis.NormalMode))
		require.Equal(t, time.Duration(args.SecurityModeFailureConfig.LimitPeriodInSec)*time.Second, rl.Period(redis.SecurityMode))
		require.Equal(t, int(args.FreezeFailureConfig.MaxFailures), rl.Rate(redis.NormalMode))
		require.Equal(t, int(args.SecurityModeFailureConfig.MaxFailures), rl.Rate(redis.SecurityMode))
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
		args.FreezeFailureConfig.MaxFailures = int64(maxFailures)
		args.FreezeFailureConfig.LimitPeriodInSec = uint64(maxDuration)

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
			{redis.SecurityMode, int(args.SecurityModeFailureConfig.MaxFailures) - 1},
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
		args.FreezeFailureConfig.MaxFailures = int64(maxFailures)
		args.FreezeFailureConfig.LimitPeriodInSec = uint64(maxDuration)

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
		args.FreezeFailureConfig.MaxFailures = int64(maxFailures)
		args.FreezeFailureConfig.LimitPeriodInSec = uint64(maxDuration)
		args.SecurityModeFailureConfig.MaxFailures = int64(securityModeMaxFailures)
		args.SecurityModeFailureConfig.LimitPeriodInSec = uint64(securityModeMaxDuration)
		normalModeKey := "key"
		secureModeKey := "secureModeKey"

		testData := []struct {
			mode               redis.Mode
			key                string
			expectedRemaining  int
			expectedResetAfter time.Duration
		}{
			{redis.NormalMode, normalModeKey, 0, time.Second * time.Duration(args.FreezeFailureConfig.LimitPeriodInSec)},
			{redis.SecurityMode, secureModeKey, 0, time.Second * time.Duration(args.SecurityModeFailureConfig.LimitPeriodInSec)},
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
					return time.Second * time.Duration(args.FreezeFailureConfig.LimitPeriodInSec), nil
				case secureModeKey:
					return time.Second * time.Duration(args.SecurityModeFailureConfig.LimitPeriodInSec), nil
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

func TestSetSecurityModeNoExpire(t *testing.T) {
	t.Parallel()

	args := createMockRateLimiterArgs()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		redisClient := &testscommon.RedisClientStub{
			SetPersistCalled: func(ctx context.Context, key string) (bool, error) {
				return false, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.SetSecurityModeNoExpire("key1")
		require.Nil(t, err)
	})

	t.Run("should fail", func(t *testing.T) {
		t.Parallel()

		redisClient := &testscommon.RedisClientStub{
			SetPersistCalled: func(ctx context.Context, key string) (bool, error) {
				return false, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.SetSecurityModeNoExpire("key1")
		require.NotNil(t, err)
	})
}

func TestUnsetSecurityModeNoExpire(t *testing.T) {
	t.Parallel()

	args := createMockRateLimiterArgs()

	t.Run("should return nil", func(t *testing.T) {
		t.Parallel()

		redisClient := &testscommon.RedisClientStub{
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return 0, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.UnsetSecurityModeNoExpire("key1")
		require.Nil(t, err)
	})

	t.Run("should return nil because of ErrKeyNotExists", func(t *testing.T) {
		t.Parallel()

		redisClient := &testscommon.RedisClientStub{
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return 0, redis.ErrKeyNotExists
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.UnsetSecurityModeNoExpire("key1")
		require.Nil(t, err)
	})

	t.Run("should call SetExpire", func(t *testing.T) {
		t.Parallel()

		redisClient := &testscommon.RedisClientStub{
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return 0, redis.ErrNoExpirationTimeForKey
			},
			SetExpireCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				return false, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.UnsetSecurityModeNoExpire("key1")
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return another error", func(t *testing.T) {
		t.Parallel()

		redisClient := &testscommon.RedisClientStub{
			ExpireTimeCalled: func(ctx context.Context, key string) (time.Duration, error) {
				return 0, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.UnsetSecurityModeNoExpire("key1")
		require.Equal(t, expectedErr, err)
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

func TestExtendSecurityMode(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()

		redisClient := &testscommon.RedisClientStub{
			SetGreaterExpireTTLCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				return false, expectedErr
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.ExtendSecurityMode("key")
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockRateLimiterArgs()

		wasCalled := false
		redisClient := &testscommon.RedisClientStub{
			SetGreaterExpireTTLCalled: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
				wasCalled = true
				return true, nil
			},
		}
		args.Storer = redisClient

		rl, err := redis.NewRateLimiter(args)
		require.Nil(t, err)

		err = rl.ExtendSecurityMode("key")
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
			case 4:
				// do not check returned err, Reset might have been called
				_ = rl.ExtendSecurityMode(testKey)
			default:
				assert.Fail(t, "should have not been called")
			}

			wg.Done()
		}(i % 5)
	}

	wg.Wait()
}
