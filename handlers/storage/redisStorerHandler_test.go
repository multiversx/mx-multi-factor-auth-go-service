package storage_test

import (
	"testing"

	"github.com/go-redis/redismock/v8"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/stretchr/testify/require"
)

func TestNewRedisStorerHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil redis client, should fail", func(t *testing.T) {
		t.Parallel()

		storer, err := storage.NewRedisStorerHandler(nil)
		require.Nil(t, storer)
		require.Equal(t, handlers.ErrNilRedisClient, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		client, _ := redismock.NewClientMock()

		storer, err := storage.NewRedisStorerHandler(client)
		require.Nil(t, err)
		require.NotNil(t, storer)
	})
}

func TestRedisStorerHandler_Operations(t *testing.T) {
	t.Parallel()

	client, mock := redismock.NewClientMock()

	storer, err := storage.NewRedisStorerHandler(client)
	require.Nil(t, err)

	key := "key1"
	value := []byte("data1")

	mock.ExpectSet(key, value, 0).SetVal(key)
	mock.ExpectGet(key).SetVal(string(value))
	mock.ExpectExists(key).SetVal(1)
	mock.ExpectDel(string(key)).SetVal(1)

	err = storer.Put([]byte(key), value)
	require.NoError(t, err)

	val, err := storer.Get([]byte(key))
	require.NoError(t, err)
	require.Equal(t, value, val)

	has := storer.Has([]byte(key))
	require.Nil(t, has)

	ok := storer.Remove([]byte(key))
	require.Nil(t, ok)
}
