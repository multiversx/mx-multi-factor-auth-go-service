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

	wasSet, err = rcw.SetPersist(context.TODO(), "key1")
	require.Nil(t, err)
	require.True(t, wasSet)

	wasSet, err = rcw.SetPersist(context.TODO(), "key1")
	require.Nil(t, err)
	require.False(t, wasSet)

	wasSet, err = rcw.SetPersist(context.TODO(), "invalidKey")
	require.Nil(t, err)
	require.False(t, wasSet)

	err = rcw.ResetCounterAndKeepTTL(context.TODO(), "key1")
	require.Nil(t, err)

	wasSet, err = rcw.SetAllIfNoneExist(context.TODO(), []string{"key2"}, ttl)
	require.Nil(t, err)
	require.True(t, wasSet)

	wasSet, err = rcw.SetAllIfNoneExist(context.TODO(), []string{"key2"}, ttl)
	require.Nil(t, err)
	require.False(t, wasSet)

	server.FastForward(ttl)

	wasSet, err = rcw.SetAllIfNoneExist(context.TODO(), []string{"key2"}, ttl)
	require.Nil(t, err)
	require.True(t, wasSet)
}

func TestRedisClientWrapper_SetAllIfNoneExist(t *testing.T) {
	t.Parallel()

	ttl := time.Minute

	t.Run("invalid args should error", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		rcw, _ := redis.NewRedisClientWrapper(redisClient.NewClient(&redisClient.Options{Addr: server.Addr()}))

		wasSet, err := rcw.SetAllIfNoneExist(context.TODO(), nil, ttl)
		require.Equal(t, redis.ErrNoKeysProvided, err)
		require.False(t, wasSet)

		wasSet, err = rcw.SetAllIfNoneExist(context.TODO(), []string{"key"}, 0)
		require.Equal(t, redis.ErrInvalidTTL, err)
		require.False(t, wasSet)

		wasSet, err = rcw.SetAllIfNoneExist(context.TODO(), []string{"key", "key"}, ttl)
		require.Equal(t, redis.ErrDuplicatedKeys, err)
		require.False(t, wasSet)
		require.False(t, server.Exists("key"))
	})
	t.Run("should set all keys with ttl", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		rcw, _ := redis.NewRedisClientWrapper(redisClient.NewClient(&redisClient.Options{Addr: server.Addr()}))

		wasSet, err := rcw.SetAllIfNoneExist(context.TODO(), []string{"key1", "key2"}, ttl)
		require.Nil(t, err)
		require.True(t, wasSet)
		require.Equal(t, ttl, server.TTL("key1"))
		require.Equal(t, ttl, server.TTL("key2"))
	})
	t.Run("existing key should not set any key", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		rcw, _ := redis.NewRedisClientWrapper(redisClient.NewClient(&redisClient.Options{Addr: server.Addr()}))

		wasSet, err := rcw.SetAllIfNoneExist(context.TODO(), []string{"key2"}, ttl)
		require.Nil(t, err)
		require.True(t, wasSet)

		wasSet, err = rcw.SetAllIfNoneExist(context.TODO(), []string{"key1", "key2"}, ttl)
		require.Nil(t, err)
		require.False(t, wasSet)
		require.False(t, server.Exists("key1"))

		wasSet, err = rcw.SetAllIfNoneExist(context.TODO(), []string{"key1", "key3"}, ttl)
		require.Nil(t, err)
		require.True(t, wasSet)
	})
}

func TestRedisClientWrapper_Decrement(t *testing.T) {
	t.Parallel()

	server := miniredis.RunT(t)
	rcw, _ := redis.NewRedisClientWrapper(redisClient.NewClient(&redisClient.Options{Addr: server.Addr()}))

	value, err := rcw.Decrement(context.TODO(), "missing")
	require.Nil(t, err)
	require.Equal(t, int64(0), value)
	require.False(t, server.Exists("missing"))

	require.Nil(t, server.Set("zero", "0"))
	value, err = rcw.Decrement(context.TODO(), "zero")
	require.Nil(t, err)
	require.Equal(t, int64(0), value)

	require.Nil(t, server.Set("negative", "-4"))
	server.SetTTL("negative", time.Minute)
	value, err = rcw.Decrement(context.TODO(), "negative")
	require.Nil(t, err)
	require.Equal(t, int64(0), value)
	stored, _ := server.Get("negative")
	require.Equal(t, "0", stored)
	require.Equal(t, time.Minute, server.TTL("negative"))

	require.Nil(t, server.Set("persistent", "-2"))
	value, err = rcw.Decrement(context.TODO(), "persistent")
	require.Nil(t, err)
	require.Equal(t, int64(0), value)
	require.Equal(t, time.Duration(0), server.TTL("persistent"))

	require.Nil(t, server.Set("positive", "2"))
	value, err = rcw.Decrement(context.TODO(), "positive")
	require.Nil(t, err)
	require.Equal(t, int64(1), value)
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
