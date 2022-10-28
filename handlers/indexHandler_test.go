package handlers

import (
	"errors"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/stretchr/testify/assert"
)

func TestNewIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil db should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewIndexHandler(nil, &testscommon.MarshalizerStub{})
		assert.Equal(t, ErrNilDB, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewIndexHandler(&mock.PersisterStub{}, nil)
		assert.Equal(t, ErrNilMarshaller, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work with empty db", func(t *testing.T) {
		t.Parallel()

		handler, err := NewIndexHandler(&mock.PersisterStub{}, &testscommon.MarshalizerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		marshaller := &testscommon.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		db := &mock.PersisterStub{
			RangeKeysCalled: func(handler func(key []byte, val []byte) bool) {
				handler([]byte("key"), []byte("val"))
			},
		}
		handler, err := NewIndexHandler(db, marshaller)
		assert.Equal(t, expectedErr, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work with db", func(t *testing.T) {
		t.Parallel()

		marshaller := testscommon.MarshalizerMock{}
		db := testscommon.NewMemDbMock()
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

		handler, err := NewIndexHandler(db, marshaller)
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
		marshaller := testscommon.MarshalizerMock{}
		db := testscommon.NewMemDbMock()
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

		handler, err := NewIndexHandler(db, marshaller)
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
