package bucket

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
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

func TestShardedStorageWithIndex_getBucketForKey(t *testing.T) {
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

		bucket, bucketID, err := sswi.getBucketForKey(providedAddr)
		assert.True(t, errors.Is(err, core.ErrInvalidBucketID))
		assert.Nil(t, bucket)
		assert.Zero(t, bucketID)
	})
	t.Run("should work with key address only", func(t *testing.T) {
		t.Parallel()

		providedAddr, _ := data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
		providedIdx := uint32(1)
		provider := &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				assert.Equal(t, providedAddr.AddressBytes(), address)
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

		bucket, bucketID, err := sswi.getBucketForKey(providedAddr.AddressBytes())
		assert.Nil(t, err)
		assert.Equal(t, bucketHandlers[providedIdx], bucket) // pointer testing
		assert.Equal(t, providedIdx, bucketID)
	})
	t.Run("should work with key address and guardian", func(t *testing.T) {
		t.Parallel()

		providedAddr, _ := data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
		providedGuardian := []byte("guardian")
		providedIdx := uint32(1)
		provider := &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				assert.True(t, bytes.Contains(address, providedAddr.AddressBytes()))
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

		providedKey := []byte(fmt.Sprintf("%s_%s", providedGuardian, providedAddr.AddressBytes()))
		bucket, bucketID, err := sswi.getBucketForKey(providedKey)
		assert.Nil(t, err)
		assert.Equal(t, bucketHandlers[providedIdx], bucket) // pointer testing
		assert.Equal(t, providedIdx, bucketID)
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
				AllocateBucketIndexCalled: func() (uint32, error) {
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

func TestShardedStorageWithIndex_geBucketIDAndBaseIndex(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testGetBucketIDAndBaseIndex(false))
	t.Run("should work", testGetBucketIDAndBaseIndex(true))
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

func TestShardedStorageWithIndex_Count(t *testing.T) {
	t.Parallel()

	t.Run("one bucked returns error should error", func(t *testing.T) {
		t.Parallel()

		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{
				GetLastIndexCalled: func() (uint32, error) {
					return uint32(100), nil
				},
			},
			1: &testscommon.BucketIndexHandlerStub{
				GetLastIndexCalled: func() (uint32, error) {
					return 0, expectedErr
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))
		count, err := sswi.Count()
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, uint32(0), count)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedLastIndex0, providedLastIndex1, providedLastIndex2 := uint32(100), uint32(200), uint32(300)
		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{
				GetLastIndexCalled: func() (uint32, error) {
					return providedLastIndex0, nil
				},
			},
			1: &testscommon.BucketIndexHandlerStub{
				GetLastIndexCalled: func() (uint32, error) {
					return providedLastIndex1, nil
				},
			},
			2: &testscommon.BucketIndexHandlerStub{
				GetLastIndexCalled: func() (uint32, error) {
					return providedLastIndex2, nil
				},
			},
		}
		args := ArgShardedStorageWithIndex{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		sswi, _ := NewShardedStorageWithIndex(args)
		assert.False(t, check.IfNil(sswi))
		count, err := sswi.Count()
		assert.Nil(t, err)
		expectedCount := providedLastIndex0 + providedLastIndex1 + providedLastIndex2
		assert.Equal(t, expectedCount, count)
	})
}

func testGetBucketIDAndBaseIndex(shouldWork bool) func(t *testing.T) {
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
				AllocateBucketIndexCalled: func() (uint32, error) {
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

		bucketID, index, err := sswi.getBucketIDAndBaseIndex(providedAddr)
		if shouldWork {
			assert.Nil(t, err)
			assert.True(t, wasCalled)
			assert.Equal(t, uint32(10), index)
			assert.Equal(t, providedIdx, bucketID)
		} else {
			assert.True(t, errors.Is(err, core.ErrInvalidBucketID))
			assert.False(t, wasCalled)
			assert.Zero(t, index)
			assert.Zero(t, bucketID)
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
