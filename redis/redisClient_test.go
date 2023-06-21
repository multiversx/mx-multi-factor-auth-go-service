package redis_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	_, err = rcw.SetEntryIfNotExisting(context.TODO(), "key1", 3, ttl)
	require.Nil(t, err)

	v, exp, err := rcw.DecrementWithExpireTime(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, int64(2), v)
	require.Equal(t, ttl, exp)

	err = rcw.Delete(context.TODO(), "key1")
	require.Nil(t, err)

	wasSet, err := rcw.SetEntryIfNotExisting(context.TODO(), "key1", 3, ttl)
	require.Nil(t, err)
	require.True(t, wasSet)

	wasSet, err = rcw.SetEntryIfNotExisting(context.TODO(), "key1", 3, ttl)
	require.Nil(t, err)
	require.False(t, wasSet)

	err = rcw.Delete(context.TODO(), "key1")
	require.Nil(t, err)

	wasSet, err = rcw.SetEntryIfNotExisting(context.TODO(), "key1", 3, 0)
	require.Nil(t, err)
	require.True(t, wasSet)
}

func TestConcurrentOperations(t *testing.T) {
	t.Parallel()

	server := miniredis.RunT(t)
	rc := redisClient.NewClient(&redisClient.Options{
		Addr: server.Addr(),
	})

	rcw, err := redis.NewRedisClientWrapper(rc)
	require.Nil(t, err)

	ttl := time.Second * time.Duration(1)

	wg := sync.WaitGroup{}

	numConcurrentCalls := 500
	wg.Add(numConcurrentCalls)

	for i := 1; i <= numConcurrentCalls; i++ {
		go func(idx int) {
			switch idx % 3 {
			case 0:
				_, err = rcw.SetEntryIfNotExisting(context.TODO(), "key1", 3, ttl)
				assert.Nil(t, err)
			case 1:
				_, _, err := rcw.DecrementWithExpireTime(context.TODO(), "key1")
				if err != redis.ErrKeyNotExists && err != redis.ErrNoExpirationTimeForKey {
					assert.Nil(t, err)
				}
			case 2:
				err = rcw.Delete(context.TODO(), "key1")
				assert.Nil(t, err)
			default:
				assert.Fail(t, "should not hit default")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
