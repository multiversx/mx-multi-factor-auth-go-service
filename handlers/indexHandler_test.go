package handlers_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgsIndexHandler() handlers.ArgIndexHandler {
	return handlers.ArgIndexHandler{
		BucketIDProvider:  &testscommon.BucketIDProviderStub{},
		BucketIndexHolder: &testscommon.BucketIndexHolderStub{},
	}
}

func TestNewIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil BucketIndexHolder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler()
		args.BucketIndexHolder = nil
		handler, err := handlers.NewIndexHandler(args)
		assert.Equal(t, handlers.ErrNilBucketIndexHolder, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("nil BucketIDProvider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler()
		args.BucketIDProvider = nil
		handler, err := handlers.NewIndexHandler(args)
		assert.Equal(t, handlers.ErrNilBucketIDProvider, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work, empty", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(createMockArgsIndexHandler())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
}

func TestIndexHandler_AllocateIndex(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	providedAddr := []byte("addr")
	providedIndex := uint32(10)
	providedBucketId := uint32(5)
	providedNumberOfBuckets := uint32(15)
	t.Run("update index returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler()
		args.BucketIndexHolder = &testscommon.BucketIndexHolderStub{
			UpdateIndexReturningNextCalled: func(address []byte) (uint32, error) {
				assert.Equal(t, providedAddr, address)
				return 0, expectedErr
			},
		}
		handler, _ := handlers.NewIndexHandler(args)
		assert.False(t, check.IfNil(handler))

		nextIndex, err := handler.AllocateIndex(providedAddr)
		assert.Equal(t, expectedErr, err)
		assert.Zero(t, nextIndex)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsIndexHandler()
		args.BucketIndexHolder = &testscommon.BucketIndexHolderStub{
			UpdateIndexReturningNextCalled: func(address []byte) (uint32, error) {
				assert.Equal(t, providedAddr, address)
				return providedIndex, nil
			},
			NumberOfBucketsCalled: func() uint32 {
				return providedNumberOfBuckets
			},
		}
		args.BucketIDProvider = &testscommon.BucketIDProviderStub{
			GetBucketForAddressCalled: func(address []byte) uint32 {
				return providedBucketId
			},
		}
		handler, _ := handlers.NewIndexHandler(args)
		assert.False(t, check.IfNil(handler))

		expectedIndex := 2 * (providedIndex*providedNumberOfBuckets + providedBucketId)
		nextIndex, err := handler.AllocateIndex(providedAddr)
		assert.Nil(t, err)
		assert.Equal(t, expectedIndex, nextIndex)
	})
}
