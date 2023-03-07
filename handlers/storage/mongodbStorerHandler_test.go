package storage_test

import (
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
)

func TestMongoDBStorerHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil mongodb client should fail", func(t *testing.T) {
		t.Parallel()

		storer, err := storage.NewMongoDBStorerHandler(nil, mongodb.Collection(""))
		require.Nil(t, storer)
		require.Equal(t, storage.ErrNilMongoDBClient, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		storer, err := storage.NewMongoDBStorerHandler(&testscommon.MongoDBClientStub{}, mongodb.UsersCollection)
		require.Nil(t, err)
		require.False(t, storer.IsInterfaceNil())
	})
}

func TestMongoDBStorerHandler_Operations(t *testing.T) {
	t.Parallel()

	key1 := []byte("key1")
	value := []byte("data1")

	putWasCalled := false
	getWasCalled := false
	hasWasCalled := false
	removeWasCalled := false
	client := &testscommon.MongoDBClientStub{
		PutCalled: func(coll mongodb.Collection, key, data []byte) error {
			require.Equal(t, mongodb.UsersCollection, coll)
			putWasCalled = true

			return nil
		},
		GetCalled: func(coll mongodb.Collection, key []byte) ([]byte, error) {
			require.Equal(t, key1, key)
			getWasCalled = true
			return value, nil
		},
		HasCalled: func(coll mongodb.Collection, key []byte) error {
			hasWasCalled = true
			return nil
		},
		RemoveCalled: func(coll mongodb.Collection, key []byte) error {
			removeWasCalled = true
			return nil
		},
	}

	storer, err := storage.NewMongoDBStorerHandler(client, mongodb.UsersCollection)
	require.Nil(t, err)

	err = storer.Put(key1, value)
	require.Nil(t, err)

	err = storer.Has(key1)
	require.Nil(t, err)

	v, err := storer.Get(key1)
	require.Nil(t, err)
	require.Equal(t, value, v)

	err = storer.Remove(key1)
	require.Nil(t, err)

	require.True(t, putWasCalled)
	require.True(t, getWasCalled)
	require.True(t, hasWasCalled)
	require.True(t, removeWasCalled)
}
