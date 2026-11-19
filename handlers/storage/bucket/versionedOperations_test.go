package bucket

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/mx-multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createShardedStorageForVersionTests(t *testing.T, handler core.IndexHandler, bucketID uint32) *shardedStorageWithIndex {
	sswi, err := NewShardedStorageWithIndex(ArgShardedStorageWithIndex{
		BucketIDProvider: &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				return 1
			},
		},
		BucketHandlers: map[uint32]core.IndexHandler{bucketID: handler},
	})
	require.Nil(t, err)

	return sswi
}

func TestShardedStorageWithIndex_VersionedOperations(t *testing.T) {
	t.Parallel()

	providedKey := []byte("addr")
	providedData := []byte("data")

	t.Run("invalid bucket should error", func(t *testing.T) {
		t.Parallel()

		sswi := createShardedStorageForVersionTests(t, &testscommon.BucketIndexHandlerStub{}, 2)

		data, version, err := sswi.GetWithVersion(providedKey)
		assert.True(t, errors.Is(err, core.ErrInvalidBucketID))
		assert.Equal(t, make([]byte, 0), data)
		assert.Zero(t, version)

		assert.True(t, errors.Is(sswi.Insert(providedKey, providedData), core.ErrInvalidBucketID))
		assert.True(t, errors.Is(sswi.PutIfVersion(providedKey, providedData, 1), core.ErrInvalidBucketID))
	})
	t.Run("should forward to the right bucket", func(t *testing.T) {
		t.Parallel()

		calls := make(map[string]int)
		sswi := createShardedStorageForVersionTests(t, &testscommon.BucketIndexHandlerStub{
			GetWithVersionCalled: func(key []byte) ([]byte, int64, error) {
				assert.Equal(t, providedKey, key)
				calls["get"]++
				return providedData, 5, nil
			},
			InsertCalled: func(key, data []byte) error {
				assert.Equal(t, providedKey, key)
				assert.Equal(t, providedData, data)
				calls["insert"]++
				return storage.ErrKeyAlreadyExists
			},
			PutIfVersionCalled: func(key, data []byte, expectedVersion int64) error {
				assert.Equal(t, providedKey, key)
				assert.Equal(t, providedData, data)
				assert.Equal(t, int64(5), expectedVersion)
				calls["put"]++
				return storage.ErrVersionMismatch
			},
		}, 1)

		data, version, err := sswi.GetWithVersion(providedKey)
		assert.Nil(t, err)
		assert.Equal(t, providedData, data)
		assert.Equal(t, int64(5), version)

		assert.Equal(t, storage.ErrKeyAlreadyExists, sswi.Insert(providedKey, providedData))
		assert.Equal(t, storage.ErrVersionMismatch, sswi.PutIfVersion(providedKey, providedData, 5))
		assert.Equal(t, map[string]int{"get": 1, "insert": 1, "put": 1}, calls)
	})
}

func TestMongoDBIndexHandler_VersionedOperations(t *testing.T) {
	t.Parallel()

	providedColl := mongodb.CollectionID("collName")
	providedKey := []byte("key")
	providedData := []byte("data")
	calls := make(map[string]int)
	handler, err := NewMongoDBIndexHandler(&testscommon.MongoDBClientStub{
		GetWithVersionCalled: func(coll mongodb.CollectionID, key []byte) ([]byte, int64, error) {
			assert.Equal(t, providedColl, coll)
			assert.Equal(t, providedKey, key)
			calls["get"]++
			return providedData, 3, nil
		},
		InsertCalled: func(coll mongodb.CollectionID, key []byte, data []byte) error {
			assert.Equal(t, providedColl, coll)
			calls["insert"]++
			return nil
		},
		PutIfVersionCalled: func(coll mongodb.CollectionID, key []byte, data []byte, expectedVersion int64) error {
			assert.Equal(t, providedColl, coll)
			assert.Equal(t, int64(3), expectedVersion)
			calls["put"]++
			return nil
		},
	}, providedColl)
	require.Nil(t, err)

	data, version, err := handler.GetWithVersion(providedKey)
	assert.Nil(t, err)
	assert.Equal(t, providedData, data)
	assert.Equal(t, int64(3), version)
	assert.Nil(t, handler.Insert(providedKey, providedData))
	assert.Nil(t, handler.PutIfVersion(providedKey, providedData, 3))
	assert.Equal(t, map[string]int{"get": 1, "insert": 1, "put": 1}, calls)
}

func TestBucketIndexHandler_VersionedOperations(t *testing.T) {
	t.Parallel()

	t.Run("get with version returns version 0", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewBucketIndexHandler(testscommon.NewStorerMock())
		require.Nil(t, handler.Put([]byte("key"), []byte("data")))

		data, version, err := handler.GetWithVersion([]byte("key"))
		assert.Nil(t, err)
		assert.Equal(t, []byte("data"), data)
		assert.Zero(t, version)
	})
	t.Run("get with version on missing key returns error", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewBucketIndexHandler(testscommon.NewStorerMock())

		_, _, err := handler.GetWithVersion([]byte("missing"))
		assert.NotNil(t, err)
	})
	t.Run("insert should fail if key exists", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewBucketIndexHandler(testscommon.NewStorerMock())
		assert.Nil(t, handler.Insert([]byte("key"), []byte("data")))
		assert.Equal(t, storage.ErrKeyAlreadyExists, handler.Insert([]byte("key"), []byte("other")))

		data, err := handler.Get([]byte("key"))
		assert.Nil(t, err)
		assert.Equal(t, []byte("data"), data)
	})
	t.Run("put if version ignores the version", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewBucketIndexHandler(testscommon.NewStorerMock())
		assert.Nil(t, handler.Insert([]byte("key"), []byte("data")))
		assert.Nil(t, handler.PutIfVersion([]byte("key"), []byte("new data"), 42))

		data, err := handler.Get([]byte("key"))
		assert.Nil(t, err)
		assert.Equal(t, []byte("new data"), data)
	})
}
