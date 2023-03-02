package redis_test

import (
	"testing"

	"github.com/go-redis/redismock/v8"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/stretchr/testify/require"
)

func TestRedisClientWrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil redis client, should fail", func(t *testing.T) {
		t.Parallel()

		storer, err := redis.NewRedisClientWrapper(nil)
		require.Nil(t, storer)
		require.Equal(t, redis.ErrNilRedisClient, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		client, _ := redismock.NewClientMock()

		storer, err := redis.NewRedisClientWrapper(client)
		require.Nil(t, err)
		require.NotNil(t, storer)
	})
}

func TestRedisClientWrapper_Operations(t *testing.T) {
	t.Parallel()

	client, mock := redismock.NewClientMock()

	clientWrapper, err := redis.NewRedisClientWrapper(client)
	require.Nil(t, err)

	key := "key1"
	value := "data1"

	mock.ExpectSet(key, value, 0).SetVal(key)
	mock.ExpectGet(key).SetVal(value)
	mock.ExpectExists(key).SetVal(1)
	mock.ExpectDel(string(key)).SetVal(1)

	_, err = clientWrapper.Set(key, value, 0)
	require.NoError(t, err)

	val, err := clientWrapper.Get(key)
	require.NoError(t, err)
	require.Equal(t, value, val)

	num, err := clientWrapper.Exists(key)
	require.Nil(t, err)
	require.Equal(t, int64(1), num)

	num, err = clientWrapper.Del(key)
	require.Nil(t, err)
	require.Equal(t, int64(1), num)

	err = clientWrapper.Close()
	require.Nil(t, err)
}
