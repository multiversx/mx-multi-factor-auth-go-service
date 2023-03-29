package bucket

import (
	"encoding/binary"
	"sync"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewMongoDBIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil storer should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewMongoDBIndexHandler(nil, &testscommon.MongoDBClientStub{})
		assert.Equal(t, core.ErrNilStorer, err)
		assert.True(t, check.IfNil(handler))
	})

	t.Run("nil mongo clinet should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewMongoDBIndexHandler(&testscommon.StorerStub{}, nil)
		assert.Equal(t, core.ErrNilMongoDBClient, err)
		assert.True(t, check.IfNil(handler))
	})

	t.Run("should work, bucket has lastIndexKey", func(t *testing.T) {
		t.Parallel()

		handler, err := NewMongoDBIndexHandler(&testscommon.StorerStub{
			HasCalled: func(key []byte) error {
				assert.Equal(t, []byte(lastIndexKey), key)
				return nil
			},
		}, &testscommon.MongoDBClientStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})

	t.Run("should work, empty bucket", func(t *testing.T) {
		t.Parallel()

		handler, err := NewMongoDBIndexHandler(&testscommon.StorerStub{
			HasCalled: func(key []byte) error {
				assert.Equal(t, []byte(lastIndexKey), key)
				return expectedErr
			},
			PutCalled: func(key, data []byte) error {
				assert.Equal(t, []byte(lastIndexKey), key)
				return nil
			},
		}, &testscommon.MongoDBClientStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})

	t.Run("empty bucket and put lastIndexKey fails", func(t *testing.T) {
		t.Parallel()

		handler, err := NewMongoDBIndexHandler(&testscommon.StorerStub{}, &testscommon.MongoDBClientStub{
			PutIndexIfNotExistsCalled: func(collID mongodb.CollectionID, key []byte, index uint32) error {
				return expectedErr
			},
		})
		assert.Equal(t, expectedErr, err)
		assert.True(t, check.IfNil(handler))
	})
}

func TestMongoDBIndexHandler_Operations(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should have not panicked")
		}
	}()

	handler, _ := NewMongoDBIndexHandler(&testscommon.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			index := make([]byte, uint32Bytes)
			binary.BigEndian.PutUint32(index, 0)
			return index, nil
		},
	}, &testscommon.MongoDBClientStub{
		IncrementIndexCalled: func(collID mongodb.CollectionID, key []byte) (uint32, error) {
			return 1, nil
		},
	})

	numCalls := 10000
	var wg sync.WaitGroup
	wg.Add(numCalls)
	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			switch idx % 6 {
			case 0:
				_, err := handler.AllocateBucketIndex()
				assert.Nil(t, err)
			case 1:
				assert.Nil(t, handler.Put([]byte("key"), []byte("data")))
			case 2:
				_, err := handler.Get([]byte("key"))
				assert.Nil(t, err)
			case 3:
				assert.Nil(t, handler.Has([]byte("key")))
			case 4:
				assert.Nil(t, handler.Close())
			case 5:
				_, err := handler.GetLastIndex()
				assert.Nil(t, err)
			default:
				assert.Fail(t, "should not hit default")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
