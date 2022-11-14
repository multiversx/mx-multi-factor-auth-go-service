package handlers_test

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil db should error", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(nil)
		assert.Equal(t, handlers.ErrNilDB, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work, empty", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(testscommon.NewStorerMock())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
}

func TestIndexHandler_AllocateIndex(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("get from db returns error", func(t *testing.T) {
		t.Parallel()

		db := &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		handler, err := handlers.NewIndexHandler(db)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		index, err := handler.AllocateIndex()
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, uint32(0), index)
	})
	t.Run("put to db returns error", func(t *testing.T) {
		t.Parallel()

		db := &testscommon.StorerStub{
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
		handler, err := handlers.NewIndexHandler(db)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		index, err := handler.AllocateIndex()
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, uint32(0), index)
	})
	t.Run("should work with empty db", func(t *testing.T) {
		t.Parallel()

		db := testscommon.NewStorerMock()
		handler, err := handlers.NewIndexHandler(db)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		index, err := handler.AllocateIndex()
		assert.Nil(t, err)
		assert.Equal(t, uint32(1), index)
	})
	t.Run("should work with filled db", func(t *testing.T) {
		t.Parallel()

		db := testscommon.NewStorerMock()
		providedIndex := uint32(150)
		latestIndexBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(latestIndexBytes, providedIndex)
		err := db.Put([]byte("lastAllocatedIndex"), latestIndexBytes)
		assert.Nil(t, err)

		handler, err := handlers.NewIndexHandler(db)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		index, err := handler.AllocateIndex()
		assert.Nil(t, err)
		assert.Equal(t, providedIndex+1, index)
	})
}
