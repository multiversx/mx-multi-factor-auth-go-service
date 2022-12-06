package bucket

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewShardedStorageWithIndex(t *testing.T) {
	t.Parallel()

	t.Run("nil BucketIDProvider should error", func(t *testing.T) {
		t.Parallel()

		args := ArgShardedStorageWithIndex{
			BucketIDProvider: nil,
			BucketHandlers:   nil,
		}
		sswi, err := NewShardedStorageWithIndex(args)
		assert.Equal(t, core.ErrNilBucketIDProvider, err)
		assert.True(t, check.IfNil(sswi))
	})
	t.Run("nil BucketHandlers should error", func(t *testing.T) {
		t.Parallel()

		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   nil,
		}
		sswi, err := NewShardedStorageWithIndex(args)
		assert.Equal(t, core.ErrInvalidBucketHandlers, err)
		assert.True(t, check.IfNil(sswi))
	})
	t.Run("empty BucketHandlers should error", func(t *testing.T) {
		t.Parallel()

		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   make(map[uint32]core.BucketIndexHandler, 0),
		}
		sswi, err := NewShardedStorageWithIndex(args)
		assert.Equal(t, core.ErrInvalidBucketHandlers, err)
		assert.True(t, check.IfNil(sswi))
	})
	t.Run("nil BucketHandler should error", func(t *testing.T) {
		t.Parallel()

		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{},
			1: &testscommon.BucketIndexHandlerStub{},
			2: nil,
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		sswi, err := NewShardedStorageWithIndex(args)
		assert.True(t, errors.Is(err, core.ErrNilBucketHandler))
		assert.True(t, strings.Contains(err.Error(), "id 2"))
		assert.True(t, check.IfNil(sswi))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{},
			1: &testscommon.BucketIndexHandlerStub{},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		sswi, err := NewShardedStorageWithIndex(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(sswi))
	})
}

func TestShardedStorageWithIndex_getBucketForAddress(t *testing.T) {
	t.Parallel()

	t.Run("provider returns invalid id", func(t *testing.T) {
		t.Parallel()

		providedAddr := []byte("addr")
		providedIdx := uint32(1)
		provider := &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				assert.Equal(t, providedAddr, address)
				return providedIdx
			},
		}
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			providedIdx - 1: &testscommon.BucketIndexHandlerStub{},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))

		bucket, err := sswi.getBucketForAddress(providedAddr)
		assert.True(t, errors.Is(err, core.ErrInvalidBucketID))
		assert.Nil(t, bucket)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedAddr := []byte("addr")
		providedIdx := uint32(1)
		provider := &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				assert.Equal(t, providedAddr, address)
				return providedIdx
			},
		}
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			providedIdx: &testscommon.BucketIndexHandlerStub{},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))

		bucket, err := sswi.getBucketForAddress(providedAddr)
		assert.Nil(t, err)
		assert.Equal(t, bucketHandlers[providedIdx], bucket) // pointer testing
	})
}

func TestIndexHandler_AllocateIndex(t *testing.T) {
	t.Parallel()

	providedAddr := []byte("addr")
	t.Run("get base index returns error", func(t *testing.T) {
		t.Parallel()

		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{},
			1: &testscommon.BucketIndexHandlerStub{},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{
				GetBucketForAddressCalled: func(address []byte) uint32 {
					assert.Equal(t, providedAddr, address)
					return uint32(len(bucketHandlers))
				},
			},
			BucketHandlers: bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))

		nextIndex, err := sswi.AllocateIndex(providedAddr)
		assert.True(t, errors.Is(err, core.ErrInvalidBucketID))
		assert.Zero(t, nextIndex)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedBucketID := uint32(5)
		providedIndex := uint32(100)
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{},
			1: &testscommon.BucketIndexHandlerStub{},
			2: &testscommon.BucketIndexHandlerStub{},
			3: &testscommon.BucketIndexHandlerStub{},
			4: &testscommon.BucketIndexHandlerStub{},
			5: &testscommon.BucketIndexHandlerStub{
				UpdateIndexReturningNextCalled: func() (uint32, error) {
					return providedIndex, nil
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{
				GetBucketForAddressCalled: func(address []byte) uint32 {
					assert.Equal(t, providedAddr, address)
					return providedBucketID
				},
			},
			BucketHandlers: bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))

		expectedIndex := 2 * (providedIndex*uint32(len(bucketHandlers)) + providedBucketID)
		nextIndex, err := sswi.AllocateIndex(providedAddr)
		assert.Nil(t, err)
		assert.Equal(t, expectedIndex, nextIndex)
	})
}

func TestShardedStorageWithIndex_getBaseIndex(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testGetBaseIndex(false))
	t.Run("should work", testGetBaseIndex(true))
}

func TestShardedStorageWithIndex_Has(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testHas(false))
	t.Run("should work", testHas(true))
}

func TestShardedStorageWithIndex_Get(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testGet(false))
	t.Run("should work", testGet(true))
}

func TestShardedStorageWithIndex_Put(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testPut(false))
	t.Run("should work", testPut(true))
}

func TestShardedStorageWithIndex_Close(t *testing.T) {
	t.Parallel()

	t.Run("one bucket returns error", func(t *testing.T) {
		t.Parallel()

		calledCounter := 0
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{
				CloseCalled: func() error {
					calledCounter++
					return nil
				},
			},
			1: &testscommon.BucketIndexHandlerStub{
				CloseCalled: func() error {
					calledCounter++
					return expectedErr
				},
			},
			2: &testscommon.BucketIndexHandlerStub{
				CloseCalled: func() error {
					calledCounter++
					return nil
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))
		assert.Equal(t, expectedErr, sswi.Close())
		assert.Equal(t, 3, calledCounter)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		calledCounter := 0
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{
				CloseCalled: func() error {
					calledCounter++
					return nil
				},
			},
			1: &testscommon.BucketIndexHandlerStub{
				CloseCalled: func() error {
					calledCounter++
					return nil
				},
			},
			2: &testscommon.BucketIndexHandlerStub{
				CloseCalled: func() error {
					calledCounter++
					return nil
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))
		assert.Nil(t, sswi.Close())
		assert.Equal(t, 3, calledCounter)
	})
}

func testGetBaseIndex(shouldWork bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		providedAddr := []byte("addr")
		providedIdx := uint32(1)
		provider := &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				assert.Equal(t, providedAddr, address)
				return providedIdx
			},
		}
		wasCalled := false
		key := providedIdx
		if !shouldWork {
			key = providedIdx + 1
		}
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			key: &testscommon.BucketIndexHandlerStub{
				UpdateIndexReturningNextCalled: func() (uint32, error) {
					wasCalled = true
					return 10, nil
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))

		index, err := sswi.getBaseIndex(providedAddr)
		if shouldWork {
			assert.Nil(t, err)
			assert.True(t, wasCalled)
			assert.Equal(t, uint32(10), index)
		} else {
			assert.True(t, errors.Is(err, core.ErrInvalidBucketID))
			assert.False(t, wasCalled)
			assert.Zero(t, index)
		}
	}
}

func testHas(shouldWork bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		providedAddr := []byte("addr")
		providedIdx := uint32(1)
		provider := &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				assert.Equal(t, providedAddr, address)
				return providedIdx
			},
		}
		wasCalled := false
		key := providedIdx
		if !shouldWork {
			key = providedIdx + 1
		}
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			key: &testscommon.BucketIndexHandlerStub{
				HasCalled: func(key []byte) error {
					assert.Equal(t, providedAddr, key)
					wasCalled = true
					return nil
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))

		if shouldWork {
			assert.Nil(t, sswi.Has(providedAddr))
			assert.True(t, wasCalled)
		} else {
			assert.True(t, errors.Is(sswi.Has(providedAddr), core.ErrInvalidBucketID))
			assert.False(t, wasCalled)
		}
	}
}

func testGet(shouldWork bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		providedAddr := []byte("addr")
		providedIdx := uint32(1)
		providedData := []byte("data")
		provider := &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				assert.Equal(t, providedAddr, address)
				return providedIdx
			},
		}
		wasCalled := false
		key := providedIdx
		if !shouldWork {
			key = providedIdx + 1
		}
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			key: &testscommon.BucketIndexHandlerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					assert.Equal(t, providedAddr, key)
					wasCalled = true
					return providedData, nil
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))

		data, err := sswi.Get(providedAddr)
		if shouldWork {
			assert.Nil(t, err)
			assert.Equal(t, providedData, data)
			assert.True(t, wasCalled)
		} else {
			assert.True(t, errors.Is(sswi.Put(providedAddr, providedData), core.ErrInvalidBucketID))
			assert.Equal(t, make([]byte, 0), data)
			assert.False(t, wasCalled)
		}
	}
}

func testPut(shouldWork bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		providedAddr := []byte("addr")
		providedIdx := uint32(1)
		providedData := []byte("data")
		provider := &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				assert.Equal(t, providedAddr, address)
				return providedIdx
			},
		}
		wasCalled := false
		key := providedIdx
		if !shouldWork {
			key = providedIdx + 1
		}
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			key: &testscommon.BucketIndexHandlerStub{
				PutCalled: func(key, data []byte) error {
					assert.Equal(t, providedAddr, key)
					assert.Equal(t, providedData, data)
					wasCalled = true
					return nil
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))

		if shouldWork {
			assert.Nil(t, sswi.Put(providedAddr, providedData))
			assert.True(t, wasCalled)
		} else {
			assert.True(t, errors.Is(sswi.Put(providedAddr, providedData), core.ErrInvalidBucketID))
			assert.False(t, wasCalled)
		}
	}
}
