package handlers_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	storageTests "github.com/ElrondNetwork/elrond-go-storage/testscommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil db should error", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(nil, &testscommon.MarshallerStub{})
		assert.Equal(t, handlers.ErrNilDB, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(&testscommon.StorerStub{}, nil)
		assert.Equal(t, handlers.ErrNilMarshaller, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work with empty db", func(t *testing.T) {
		t.Parallel()

		handler, err := handlers.NewIndexHandler(&testscommon.StorerStub{}, &testscommon.MarshallerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		marshaller := &testscommon.MarshallerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		db := &testscommon.StorerStub{
			RangeKeysCalled: func(handler func(key []byte, val []byte) bool) {
				handler([]byte("key"), []byte("val"))
			},
		}
		handler, err := handlers.NewIndexHandler(db, marshaller)
		assert.Equal(t, expectedErr, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work with db", func(t *testing.T) {
		t.Parallel()

		marshaller := &storageTests.MarshalizerMock{}
		db := testscommon.NewStorerMock()
		buff, err := marshaller.Marshal(&core.UserInfo{
			Index: 0,
		})
		assert.Nil(t, err)
		err = db.Put([]byte("u1"), buff)
		assert.Nil(t, err)

		buff, err = marshaller.Marshal(&core.UserInfo{
			Index: 15,
		})
		assert.Nil(t, err)
		err = db.Put([]byte("u2"), buff)
		assert.Nil(t, err)

		buff, err = marshaller.Marshal(&core.UserInfo{
			Index: 5,
		})
		assert.Nil(t, err)
		err = db.Put([]byte("u3"), buff)
		assert.Nil(t, err)

		handler, err := handlers.NewIndexHandler(db, marshaller)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		index := handler.AllocateIndex()
		assert.Equal(t, uint32(16), index)

		handler.RevertIndex()
		index = handler.AllocateIndex()
		assert.Equal(t, uint32(16), index)
	})
	t.Run("should work with db and concurrent calls", func(t *testing.T) {
		t.Parallel()

		providedLastIndex := uint32(15)
		marshaller := &storageTests.MarshalizerMock{}
		db := testscommon.NewStorerMock()
		buff, err := marshaller.Marshal(&core.UserInfo{
			Index: 0,
		})
		assert.Nil(t, err)
		err = db.Put([]byte("u1"), buff)
		assert.Nil(t, err)

		buff, err = marshaller.Marshal(&core.UserInfo{
			Index: providedLastIndex,
		})
		assert.Nil(t, err)
		err = db.Put([]byte("u2"), buff)
		assert.Nil(t, err)

		buff, err = marshaller.Marshal(&core.UserInfo{
			Index: 5,
		})
		assert.Nil(t, err)
		err = db.Put([]byte("u3"), buff)
		assert.Nil(t, err)

		handler, err := handlers.NewIndexHandler(db, marshaller)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		numCalls := 100
		var wg sync.WaitGroup
		wg.Add(numCalls)
		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				switch idx % 3 {
				case 0, 1:
					handler.AllocateIndex()
				case 2:
					handler.RevertIndex()
				}
				wg.Done()
			}(i)
		}
		wg.Wait()

		index := handler.AllocateIndex()
		assert.Equal(t, providedLastIndex+uint32(numCalls)/3+2, index)
	})
}
