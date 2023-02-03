package core

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestNewCryptoComponentsHolderFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil keyGen should error", func(t *testing.T) {
		t.Parallel()

		components, err := NewCryptoComponentsHolderFactory(nil)
		assert.Nil(t, components)
		assert.Equal(t, ErrNilKeyGenerator, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
		components, err := NewCryptoComponentsHolderFactory(keyGen)
		assert.NotNil(t, components)
		assert.Nil(t, err)
	})
}

func TestCryptoComponentsHolderFactory_Create(t *testing.T) {
	t.Parallel()

	keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
	assert.False(t, check.IfNil(keyGen))

	factory, err := NewCryptoComponentsHolderFactory(keyGen)
	assert.Nil(t, err)
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
