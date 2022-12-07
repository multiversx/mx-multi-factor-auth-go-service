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

func TestNewBucketIndexHolder(t *testing.T) {
	t.Parallel()

	t.Run("nil BucketIDProvider should error", func(t *testing.T) {
		t.Parallel()

		args := ArgBucketIndexHolder{
			BucketIDProvider: nil,
			BucketHandlers:   nil,
		}
		holder, err := NewBucketIndexHolder(args)
		assert.Equal(t, core.ErrNilBucketIDProvider, err)
		assert.True(t, check.IfNil(holder))
	})
	t.Run("nil BucketHandlers should error", func(t *testing.T) {
		t.Parallel()

		args := ArgBucketIndexHolder{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   nil,
		}
		holder, err := NewBucketIndexHolder(args)
		assert.Equal(t, core.ErrInvalidBucketHandlers, err)
		assert.True(t, check.IfNil(holder))
	})
	t.Run("empty BucketHandlers should error", func(t *testing.T) {
		t.Parallel()

		args := ArgBucketIndexHolder{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   make(map[uint32]core.BucketIndexHandler, 0),
		}
		holder, err := NewBucketIndexHolder(args)
		assert.Equal(t, core.ErrInvalidBucketHandlers, err)
		assert.True(t, check.IfNil(holder))
	})
	t.Run("nil BucketHandler should error", func(t *testing.T) {
		t.Parallel()

		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{},
			1: &testscommon.BucketIndexHandlerStub{},
			2: nil,
		}
		args := ArgBucketIndexHolder{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		holder, err := NewBucketIndexHolder(args)
		assert.True(t, errors.Is(err, core.ErrNilBucketHandler))
		assert.True(t, strings.Contains(err.Error(), "id 2"))
		assert.True(t, check.IfNil(holder))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bucketHandlers := map[uint32]core.BucketIndexHandler{
			0: &testscommon.BucketIndexHandlerStub{},
			1: &testscommon.BucketIndexHandlerStub{},
		}
		args := ArgBucketIndexHolder{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		holder, err := NewBucketIndexHolder(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(holder))
		assert.Equal(t, uint32(len(bucketHandlers)), holder.NumberOfBuckets())
	})
}

func TestBucketIndexHolder_getBucketForAddress(t *testing.T) {
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
		args := ArgBucketIndexHolder{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		holder, _ := NewBucketIndexHolder(args)
		assert.False(t, check.IfNil(holder))

		bucket, err := holder.getBucketForAddress(providedAddr)
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
		args := ArgBucketIndexHolder{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		holder, _ := NewBucketIndexHolder(args)
		assert.False(t, check.IfNil(holder))

		bucket, err := holder.getBucketForAddress(providedAddr)
		assert.Nil(t, err)
		assert.Equal(t, bucketHandlers[providedIdx], bucket) // pointer testing
	})
}

func TestBucketIndexHolder_UpdateIndexReturningNext(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testUpdateIndexReturningNext(false))
	t.Run("should work", testUpdateIndexReturningNext(true))
}

func TestBucketIndexHolder_Has(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testHas(false))
	t.Run("should work", testHas(true))
}

func TestBucketIndexHolder_Get(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testGet(false))
	t.Run("should work", testGet(true))
}

func TestBucketIndexHolder_Put(t *testing.T) {
	t.Parallel()

	t.Run("invalid address", testPut(false))
	t.Run("should work", testPut(true))
}

func TestBucketIndexHolder_Close(t *testing.T) {
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
		args := ArgBucketIndexHolder{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		holder, _ := NewBucketIndexHolder(args)
		assert.False(t, check.IfNil(holder))
		assert.Equal(t, expectedErr, holder.Close())
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
		args := ArgBucketIndexHolder{
			BucketIDProvider: &testscommon.BucketIDProviderStub{},
			BucketHandlers:   bucketHandlers,
		}
		holder, _ := NewBucketIndexHolder(args)
		assert.False(t, check.IfNil(holder))
		assert.Nil(t, holder.Close())
		assert.Equal(t, 3, calledCounter)
	})
}

func testUpdateIndexReturningNext(shouldWork bool) func(t *testing.T) {
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
		args := ArgBucketIndexHolder{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		holder, _ := NewBucketIndexHolder(args)
		assert.False(t, check.IfNil(holder))

		index, err := holder.UpdateIndexReturningNext(providedAddr)
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
		args := ArgBucketIndexHolder{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		holder, _ := NewBucketIndexHolder(args)
		assert.False(t, check.IfNil(holder))

		if shouldWork {
			assert.Nil(t, holder.Has(providedAddr))
			assert.True(t, wasCalled)
		} else {
			assert.True(t, errors.Is(holder.Has(providedAddr), core.ErrInvalidBucketID))
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
		args := ArgBucketIndexHolder{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		holder, _ := NewBucketIndexHolder(args)
		assert.False(t, check.IfNil(holder))

		data, err := holder.Get(providedAddr)
		if shouldWork {
			assert.Nil(t, err)
			assert.Equal(t, providedData, data)
			assert.True(t, wasCalled)
		} else {
			assert.True(t, errors.Is(holder.Put(providedAddr, providedData), core.ErrInvalidBucketID))
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
		args := ArgBucketIndexHolder{
			BucketIDProvider: provider,
			BucketHandlers:   bucketHandlers,
		}
		holder, _ := NewBucketIndexHolder(args)
		assert.False(t, check.IfNil(holder))

		if shouldWork {
			assert.Nil(t, holder.Put(providedAddr, providedData))
			assert.True(t, wasCalled)
		} else {
			assert.True(t, errors.Is(holder.Put(providedAddr, providedData), core.ErrInvalidBucketID))
			assert.False(t, wasCalled)
		}
	}
}