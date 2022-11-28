package handlers_test

import (
	"encoding/binary"
	"errors"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgsIndexHandler(numOfBuckets uint32) handlers.ArgIndexHandler {
	buckets := make(map[uint32]core.Storer, numOfBuckets)
	for i := uint32(0); i < numOfBuckets; i++ {
		buckets[i] = testscommon.NewStorerMock()
	}
	return handlers.ArgIndexHandler{
		BucketIDProvider: &testscommon.BucketIDProviderStub{},
		IndexBuckets:     buckets,
	}
}

func TestNewIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil bucket id provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler(4)
		args.BucketIDProvider = nil
		handler, err := handlers.NewIndexHandler(args)
		assert.Equal(t, handlers.ErrNilBucketIDProvider, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("no db should error", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(createMockArgsIndexHandler(0))
		assert.Equal(t, handlers.InvalidNumberOfBuckets, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("nil dbs should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler(5)
		args.IndexBuckets[0] = nil
		args.IndexBuckets[2] = nil
		handler, err := handlers.NewIndexHandler(args)
		assert.True(t, errors.Is(err, handlers.ErrNilDB))
		assert.True(t, strings.Contains(err.Error(), "2"))
		assert.True(t, check.IfNil(handler))
	})
	t.Run("init of one db fails", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsIndexHandler(4)
		args.IndexBuckets[2] = &testscommon.StorerStub{
			HasCalled: func(key []byte) error {
				return expectedErr
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		handler, err := handlers.NewIndexHandler(args)
		assert.Equal(t, expectedErr, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work, empty", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(createMockArgsIndexHandler(4))
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
}

func TestIndexHandler_AllocateIndex(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	providedAddrBucket0 := []byte("addr bucket 0")
	providedAddrBucket1 := []byte("addr bucket 1")
	providedAddrBucket2 := []byte("addr bucket 2")
	providedAddrBucket3 := []byte("addr bucket 3")
	t.Run("invalid resulted bucket id should error for missing mutex", func(t *testing.T) {
		t.Parallel()

		numberOfBuckets := uint32(4)
		args := createMockArgsIndexHandler(numberOfBuckets)
		args.BucketIDProvider = &testscommon.BucketIDProviderStub{
			GetIDFromAddressCalled: func(address []byte) uint32 {
				return uint32(numberOfBuckets) + 1
			},
		}

		handler, err := handlers.NewIndexHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		index, err := handler.AllocateIndex(providedAddrBucket0)
		assert.True(t, errors.Is(err, handlers.ErrInvalidBucketID))
		assert.Equal(t, uint32(0), index)
	})
	t.Run("invalid resulted bucket id should error for missing bucket", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(createMockArgsIndexHandler(4))
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		handler.RemoveBucket(0)
		index, err := handler.AllocateIndex(providedAddrBucket0)
		assert.Equal(t, handlers.ErrInvalidBucketID, err)
		assert.Equal(t, uint32(0), index)
	})
	t.Run("get from db returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler(1)
		args.IndexBuckets[0] = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		handler, err := handlers.NewIndexHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		index, err := handler.AllocateIndex(providedAddrBucket0)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, uint32(0), index)
	})
	t.Run("put to db returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler(1)
		args.IndexBuckets[0] = &testscommon.StorerStub{
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
			GetCalled: func(key []byte) ([]byte, error) {
				lastIndexBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(lastIndexBytes, 10)
				return lastIndexBytes, nil
			},
			HasCalled: func(key []byte) error {
				return nil
			},
		}
		handler, err := handlers.NewIndexHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		index, err := handler.AllocateIndex(providedAddrBucket0)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, uint32(0), index)
	})
	t.Run("should work with empty dbs", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler(4)
		args.BucketIDProvider = &testscommon.BucketIDProviderStub{
			GetIDFromAddressCalled: func(address []byte) uint32 {
				switch string(address) {
				case string(providedAddrBucket0):
					return 0
				case string(providedAddrBucket1):
					return 1
				case string(providedAddrBucket2):
					return 2
				case string(providedAddrBucket3):
					return 3
				default:
					assert.Fail(t, "should not happen")
					return 4
				}
			},
		}
		handler, err := handlers.NewIndexHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		testReturnedIndex(t, handler, providedAddrBucket0, uint32(4))
		testReturnedIndex(t, handler, providedAddrBucket1, uint32(5))
		testReturnedIndex(t, handler, providedAddrBucket2, uint32(6))
		testReturnedIndex(t, handler, providedAddrBucket3, uint32(7))
	})
	t.Run("should work with filled dbs", func(t *testing.T) {
		t.Parallel()

		numBuckets := uint32(2)
		args := createMockArgsIndexHandler(numBuckets)
		args.BucketIDProvider = &testscommon.BucketIDProviderStub{
			GetIDFromAddressCalled: func(address []byte) uint32 {
				switch string(address) {
				case string(providedAddrBucket0):
					return 0
				case string(providedAddrBucket1):
					return 1
				default:
					assert.Fail(t, "should not happen")
					return 2
				}
			},
		}
		providedIndexBucket0 := uint32(150)
		latestIndexBytes0 := make([]byte, 4)
		binary.BigEndian.PutUint32(latestIndexBytes0, providedIndexBucket0)
		assert.Nil(t, args.IndexBuckets[0].Put([]byte("lastAllocatedIndex"), latestIndexBytes0))

		providedIndexBucket1 := uint32(300)
		latestIndexBytes1 := make([]byte, 4)
		binary.BigEndian.PutUint32(latestIndexBytes1, providedIndexBucket1)
		assert.Nil(t, args.IndexBuckets[1].Put([]byte("lastAllocatedIndex"), latestIndexBytes1))

		handler, err := handlers.NewIndexHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		expectedIndex0 := (providedIndexBucket0+1)*numBuckets + 0
		expectedIndex1 := (providedIndexBucket1+1)*numBuckets + 1
		testReturnedIndex(t, handler, providedAddrBucket0, expectedIndex0)
		testReturnedIndex(t, handler, providedAddrBucket1, expectedIndex1)
	})
	t.Run("should work with concurrent calls and real bucket id provider", func(t *testing.T) {
		t.Parallel()

		numberOfBuckets := uint32(4)
		args := createMockArgsIndexHandler(numberOfBuckets)
		args.BucketIDProvider, _ = core.NewBucketIDProvider(numberOfBuckets)
		handler, err := handlers.NewIndexHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		var mutMap sync.Mutex
		indexesMap := make(map[int]uint32)
		numCalls := 1000
		var wg sync.WaitGroup
		wg.Add(numCalls)
		for i := 0; i < numCalls; i++ {
			go func(i int) {
				defer wg.Done()

				index, err := handler.AllocateIndex([]byte{byte(i)})
				assert.Nil(t, err)
				mutMap.Lock()
				indexesMap[i] = index
				mutMap.Unlock()
			}(i)
		}
		wg.Wait()
		assert.Equal(t, numCalls, len(indexesMap))

		indexesSlice := make([]uint32, 0)
		for _, value := range indexesMap {
			indexesSlice = append(indexesSlice, value)
		}
		sort.Slice(indexesSlice, func(i, j int) bool {
			return indexesSlice[i] < indexesSlice[j]
		})
		for i := 0; i < len(indexesSlice)-1; i++ {
			if indexesSlice[i] >= indexesSlice[i+1] {
				assert.Fail(t, "should not have gaps or duplicates")
				return
			}
		}
		assert.Equal(t, uint32(4), indexesSlice[0])
		providedByte := uint32(3)                                                          // 4th bucket
		expectedIndex := uint32(numCalls) + numberOfBuckets + providedByte%numberOfBuckets // 1007
		index, err := handler.AllocateIndex([]byte{byte(providedByte)})
		assert.Nil(t, err)
		assert.Equal(t, expectedIndex, index)
	})
}

func testReturnedIndex(t *testing.T, handler core.IndexHandler, providedAddr []byte, expectedIndex uint32) {
	index, err := handler.AllocateIndex(providedAddr)
	assert.Nil(t, err)
	assert.Equal(t, expectedIndex, index)
}
