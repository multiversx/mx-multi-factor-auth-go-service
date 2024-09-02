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

	retries, err := rcw.Increment(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, int64(1), retries)

	retries, err = rcw.Decrement(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, int64(0), retries)

	wasSet, err := rcw.SetExpire(context.TODO(), "key1", ttl)
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

	wasSet, err = rcw.SetPersist(context.TODO(), "key1")
	require.Nil(t, err)
	require.True(t, wasSet)

	wasSet, err = rcw.SetPersist(context.TODO(), "key1")
	require.Nil(t, err)
	require.False(t, wasSet)

	wasSet, err = rcw.SetPersist(context.TODO(), "invalidKey")
	require.NotNil(t, err)
	require.False(t, wasSet)

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
			switch idx % 4 {
			case 0:
				_, err := rcw.Increment(context.Background(), "key1")
				assert.Nil(t, err)
			case 1:
				_, err := rcw.ExpireTime(context.TODO(), "key1")
				if err != redis.ErrKeyNotExists && err != redis.ErrNoExpirationTimeForKey {
					assert.Nil(t, err)
				}
			case 2:
				_, err := rcw.SetExpire(context.TODO(), "key1", ttl)
				assert.Nil(t, err)
			case 3:
				err := rcw.ResetCounterAndKeepTTL(context.TODO(), "key1")
				assert.Nil(t, err)
			default:
				assert.Fail(t, "should not hit default")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
