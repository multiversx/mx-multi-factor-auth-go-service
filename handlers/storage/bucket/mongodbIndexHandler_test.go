package bucket

import (
	"sync"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewMongoDBIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil mongo clinet should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewMongoDBIndexHandler(nil)
		assert.Equal(t, core.ErrNilMongoDBClient, err)
		assert.Nil(t, handler)
	})

	t.Run("empty bucket and put lastIndexKey fails", func(t *testing.T) {
		t.Parallel()

		handler, err := NewMongoDBIndexHandler(&testscommon.MongoDBClientStub{
			PutIndexIfNotExistsCalled: func(collID mongodb.CollectionID, key []byte, index uint32) error {
				return expectedErr
			},
		})
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, handler)
	})

	t.Run("should work, empty bucket", func(t *testing.T) {
		t.Parallel()

		handler, err := NewMongoDBIndexHandler(&testscommon.MongoDBClientStub{})
		assert.Nil(t, err)
		assert.NotNil(t, handler)
	})
}

func TestMongodbIndexHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var mid *mongodbIndexHandler
	assert.True(t, mid.IsInterfaceNil())

	mid, _ = NewMongoDBIndexHandler(&testscommon.MongoDBClientStub{})
	assert.False(t, mid.IsInterfaceNil())
}

func TestMongoDBIndexHandler_Operations(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should have not panicked")
		}
	}()

	handler, _ := NewMongoDBIndexHandler(&testscommon.MongoDBClientStub{
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
