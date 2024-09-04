package redis_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
)

func TestNewRedisClientWrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil redis client", func(t *testing.T) {
		t.Parallel()

		rcw, err := redis.NewRedisClientWrapper(nil)
		require.Equal(t, redis.ErrNilRedisClient, err)
		require.Nil(t, rcw)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		rc := redisClient.NewClient(&redisClient.Options{
			Addr: server.Addr(),
		})

		rcw, err := redis.NewRedisClientWrapper(rc)
		require.Nil(t, err)
		require.False(t, rcw.IsInterfaceNil())

		require.True(t, rcw.IsConnected(context.TODO()))

		_ = rc.Close()
		require.False(t, rcw.IsConnected(context.TODO()))
	})
}

func TestOperations(t *testing.T) {
	t.Parallel()

	server := miniredis.RunT(t)
	rc := redisClient.NewClient(&redisClient.Options{
		Addr: server.Addr(),
	})

	rcw, err := redis.NewRedisClientWrapper(rc)
	require.Nil(t, err)
	require.False(t, rcw.IsInterfaceNil())

	ttl := time.Second * time.Duration(1)
	greaterTtl := time.Minute

	retries, err := rcw.Increment(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, int64(1), retries)

	retries, err = rcw.Decrement(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, int64(0), retries)

	wasSet, err := rcw.SetExpire(context.TODO(), "key1", ttl)
	require.Nil(t, err)
	require.True(t, wasSet)

	wasSet, err = rcw.SetGreaterExpireTTL(context.TODO(), "key1", greaterTtl)
	require.Nil(t, err)
	require.True(t, wasSet)

	err = rcw.ResetCounterAndKeepTTL(context.TODO(), "key1")
	require.Nil(t, err)

	retries, err = rcw.Increment(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, int64(1), retries)

	retries, err = rcw.Increment(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, int64(2), retries)

	err = rcw.ResetCounterAndKeepTTL(context.TODO(), "key1")
	require.Nil(t, err)
}

func TestConcurrentOperations(t *testing.T) {
	t.Parallel()

	server := miniredis.RunT(t)
	rc := redisClient.NewClient(&redisClient.Options{
		Addr: server.Addr(),
	})

	rcw, err := redis.NewRedisClientWrapper(rc)
	require.Nil(t, err)

	wg := sync.WaitGroup{}

	numConcurrentCalls := 500
	wg.Add(numConcurrentCalls)

	ttl := time.Millisecond * time.Duration(1)
	for i := 1; i <= numConcurrentCalls; i++ {
		go func(idx int) {
			switch idx % 5 {
			case 0:
				_, errIncrement := rcw.Increment(context.Background(), "key1")
				assert.Nil(t, errIncrement)
			case 1:
				_, errExpireTime := rcw.ExpireTime(context.TODO(), "key1")
				if errExpireTime != redis.ErrKeyNotExists && errExpireTime != redis.ErrNoExpirationTimeForKey {
					assert.Nil(t, errExpireTime)
				}
			case 2:
				_, errSetExpire := rcw.SetExpire(context.TODO(), "key1", ttl)
				assert.Nil(t, errSetExpire)
			case 3:
				errResetCounterAndKeepTTL := rcw.ResetCounterAndKeepTTL(context.TODO(), "key1")
				assert.Nil(t, errResetCounterAndKeepTTL)
			case 4:
				_, errSetGreaterExpireTTL := rcw.SetGreaterExpireTTL(context.TODO(), "key1", ttl)
				assert.Nil(t, errSetGreaterExpireTTL)
			default:
				assert.Fail(t, "should not hit default")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
