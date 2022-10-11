package core

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
)

func TestNewIndexHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil storer should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewIndexHandler(nil)
		assert.Equal(t, ErrNilDB, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, err := NewIndexHandler(&testsCommon.StorerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
}

func TestIndexHandler_GetIndex(t *testing.T) {
	t.Parallel()

	t.Run("empty db", func(t *testing.T) {
		t.Parallel()

		dbLen := 0
		handler, err := NewIndexHandler(&testsCommon.StorerStub{
			LenCalled: func() int {
				return dbLen
			},
		})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		assert.Equal(t, uint32(1), handler.GetIndex())
	})
	t.Run("db not empty", func(t *testing.T) {
		t.Parallel()

		dbLen := 100
		handler, err := NewIndexHandler(&testsCommon.StorerStub{
			LenCalled: func() int {
				return dbLen
			},
		})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		assert.Equal(t, uint32(dbLen)+1, handler.GetIndex())
	})
}
