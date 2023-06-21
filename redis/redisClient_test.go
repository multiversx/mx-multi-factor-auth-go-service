package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestNewRedisClientWrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil redis client", func(t *testing.T) {
		t.Parallel()

		rcw, err := redis.NewRedisClientWrapper(nil, "pr")
		require.Equal(t, redis.ErrNilRedisClient, err)
		require.Nil(t, rcw)
	})

	t.Run("invalid redis prefix", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		rc := redisClient.NewClient(&redisClient.Options{
			Addr: server.Addr(),
		})

		rcw, err := redis.NewRedisClientWrapper(rc, "")
		require.Equal(t, redis.ErrInvalidKeyPrefix, err)
		require.Nil(t, rcw)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		rc := redisClient.NewClient(&redisClient.Options{
			Addr: server.Addr(),
		})

		rcw, err := redis.NewRedisClientWrapper(rc, "pr")
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

	rcw, err := redis.NewRedisClientWrapper(rc, "pr")
	require.Nil(t, err)
	require.False(t, rcw.IsInterfaceNil())

	ttl := time.Second * time.Duration(1)

	_, err = rcw.SetEntryIfNotExisting(context.TODO(), "key1", 3, ttl)
	require.Nil(t, err)

	v, err := rcw.Decrement(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, int64(2), v)

	exp, err := rcw.ExpireTime(context.TODO(), "key1")
	require.Nil(t, err)
	require.Equal(t, ttl, exp)

	err = rcw.Delete(context.TODO(), "key1")
	require.Nil(t, err)

	exp, err = rcw.ExpireTime(context.TODO(), "key1")
	require.Equal(t, redis.ErrKeyNotExists, err)
	require.Equal(t, time.Duration(0), exp)

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

	exp, err = rcw.ExpireTime(context.TODO(), "key1")
	require.Equal(t, redis.ErrNoExpirationTimeForKey, err)
	require.Equal(t, time.Duration(0), exp)
}
