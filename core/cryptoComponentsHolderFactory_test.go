package core

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/stretchr/testify/assert"
)

func TestCryptoComponentsHolderFactory_Create(t *testing.T) {
	t.Parallel()

	factory := &CryptoComponentsHolderFactory{}
	assert.False(t, check.IfNil(factory))

	t.Run("nil private", func(t *testing.T) {
		t.Parallel()

		holder, err := factory.Create(nil)
		assert.Nil(t, holder)
		assert.Equal(t, crypto.ErrInvalidParam, err)
	})
	t.Run("invalid private", func(t *testing.T) {
		t.Parallel()

		holder, err := factory.Create([]byte("invalid"))
		assert.Nil(t, holder)
		assert.Equal(t, crypto.ErrInvalidPrivateKey, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		sk, _ := hex.DecodeString("45f72e8b6e8d10086bacd2fc8fa1340f82a3f5d4ef31953b463ea03c606533a6")
		holder, err := factory.Create(sk)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(holder))
	})
}
