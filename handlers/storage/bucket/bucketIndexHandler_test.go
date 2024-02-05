package bucket

import (
	"encoding/binary"
	"errors"
	"sync"
	"testing"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func TestNewBucketIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil bucket should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewBucketIndexHandler(nil)
		assert.Equal(t, core.ErrNilBucket, err)
		assert.Nil(t, handler)
	})
	t.Run("should work, bucket has lastIndexKey", func(t *testing.T) {
		t.Parallel()

		handler, err := NewBucketIndexHandler(&testscommon.StorerStub{
			HasCalled: func(key []byte) error {
				assert.Equal(t, []byte(lastIndexKey), key)
				return nil
			},
		})
		assert.Nil(t, err)
		assert.NotNil(t, handler)
	})
	t.Run("should work, empty bucket", func(t *testing.T) {
		t.Parallel()

		handler, err := NewBucketIndexHandler(&testscommon.StorerStub{
			HasCalled: func(key []byte) error {
				assert.Equal(t, []byte(lastIndexKey), key)
				return expectedErr
			},
			PutCalled: func(key, data []byte) error {
				assert.Equal(t, []byte(lastIndexKey), key)
				return nil
			},
		})
		assert.Nil(t, err)
		assert.NotNil(t, handler)
	})
	t.Run("empty bucket and put lastIndexKey fails", func(t *testing.T) {
		t.Parallel()

		handler, err := NewBucketIndexHandler(&testscommon.StorerStub{
			HasCalled: func(key []byte) error {
				assert.Equal(t, []byte(lastIndexKey), key)
				return expectedErr
			},
			PutCalled: func(key, data []byte) error {
				assert.Equal(t, []byte(lastIndexKey), key)
				return expectedErr
			},
		})
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, handler)
	})
}

func TestBucketIndexHandler_AllocateBucketIndex(t *testing.T) {
	t.Parallel()

	t.Run("get returns error", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewBucketIndexHandler(&testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		})
		assert.NotNil(t, handler)

		index, err := handler.AllocateBucketIndex()
		assert.Equal(t, expectedErr, err)
		assert.Zero(t, index)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewBucketIndexHandler(testscommon.NewStorerMock())
		assert.NotNil(t, handler)

		index, err := handler.AllocateBucketIndex()
		assert.Nil(t, err)
		assert.Equal(t, uint32(1), index)
		lastIndex, err := handler.GetLastIndex()
		assert.Nil(t, err)
		assert.Equal(t, uint32(1), lastIndex)

		index, err = handler.AllocateBucketIndex()
		assert.Nil(t, err)
		assert.Equal(t, uint32(2), index)
		lastIndex, err = handler.GetLastIndex()
		assert.Nil(t, err)
		assert.Equal(t, uint32(2), lastIndex)
	})
}

func TestBucketIndexHandler_ConcurrentCallsShouldWork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should have not panicked")
		}
	}()

	handler, _ := NewBucketIndexHandler(&testscommon.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			index := make([]byte, uint32Bytes)
			binary.BigEndian.PutUint32(index, 0)
			return index, nil
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
