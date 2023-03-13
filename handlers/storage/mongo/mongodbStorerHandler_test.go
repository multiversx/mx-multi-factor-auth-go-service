package mongo_test

import (
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage/mongo"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
)

func TestMongoDBStorerHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil mongodb client should fail", func(t *testing.T) {
		t.Parallel()

		storer, err := mongo.NewMongoDBStorerHandler(nil, "")
		require.Nil(t, storer)
		require.Equal(t, core.ErrNilMongoDBClient, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		storer, err := mongo.NewMongoDBStorerHandler(&testscommon.MongoDBClientStub{}, mongodb.UsersCollectionID)
		require.Nil(t, err)
		require.False(t, storer.IsInterfaceNil())
	})
}

func TestMongoDBStorerHandler_Operations(t *testing.T) {
	t.Parallel()

	key1 := []byte("key1")
	value := []byte("data1")

	putWasCalled := false
	hasWasCalled := false
	removeWasCalled := false
	closeWasCalled := false
	getCalledCount := 0
	client := &testscommon.MongoDBClientStub{
		PutCalled: func(coll mongodb.CollectionID, key, data []byte) error {
			require.Equal(t, mongodb.UsersCollectionID, coll)
			putWasCalled = true

			return nil
		},
		GetCalled: func(coll mongodb.CollectionID, key []byte) ([]byte, error) {
			require.Equal(t, key1, key)
			getCalledCount++
			return value, nil
		},
		HasCalled: func(coll mongodb.CollectionID, key []byte) error {
			hasWasCalled = true
			return nil
		},
		RemoveCalled: func(coll mongodb.CollectionID, key []byte) error {
			removeWasCalled = true
			return nil
		},
		CloseCalled: func() error {
			closeWasCalled = true
			return nil
		},
	}

	storer, err := mongo.NewMongoDBStorerHandler(client, mongodb.UsersCollectionID)
	require.Nil(t, err)

	err = storer.Put(key1, value)
	require.Nil(t, err)

	err = storer.Has(key1)
	require.Nil(t, err)

	v, err := storer.Get(key1)
	require.Nil(t, err)
	require.Equal(t, value, v)

	_, err = storer.SearchFirst(key1)
	require.Nil(t, err)

	err = storer.Remove(key1)
	require.Nil(t, err)

	err = storer.Close()
	require.Nil(t, err)

	require.NotPanics(t, func() { storer.ClearCache() }, "should not panic")

	require.True(t, putWasCalled)
	require.True(t, hasWasCalled)
	require.True(t, removeWasCalled)
	require.True(t, closeWasCalled)
	require.Equal(t, 2, getCalledCount)
}
