package storage_test

import (
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewRedisStorerHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil redis client, should fail", func(t *testing.T) {
		t.Parallel()

		storer, err := storage.NewRedisStorerHandler(nil)
		require.Nil(t, storer)
		require.Equal(t, storage.ErrNilRedisClientWrapper, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		storer, err := storage.NewRedisStorerHandler(&testscommon.RedisClientWrapperStub{})
		require.Nil(t, err)
		require.NotNil(t, storer)
	})
}

func TestRedisStorerHandler_Operations(t *testing.T) {
	t.Parallel()

	getWasCalled := false
	setWasCalled := false
	existsWasCalled := false
	client := &testscommon.RedisClientWrapperStub{
		GetCalled: func(key string) (string, error) {
			getWasCalled = true
			return "", nil
		},
		SetCalled: func(key string, value interface{}, expiration time.Duration) (string, error) {
			setWasCalled = true
			return "", nil
		},
		ExistsCalled: func(keys string) (int64, error) {
			existsWasCalled = true
			return 1, nil
		},
	}

	storer, err := storage.NewRedisStorerHandler(client)
	require.Nil(t, err)

	_, err = storer.Get([]byte("key1"))
	require.Nil(t, err)

	err = storer.Put([]byte("key1"), []byte("value1"))
	require.Nil(t, err)

	err = storer.Has([]byte("key1"))
	require.Nil(t, err)

	require.True(t, getWasCalled)
	require.True(t, setWasCalled)
	require.True(t, existsWasCalled)
}
