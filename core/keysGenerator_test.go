package core

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/mock"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgGuardianKeyGenerator {
	return ArgGuardianKeyGenerator{
		BaseKey: "base key",
		KeyGen:  &mock.KeyGenMock{},
	}
}

func TestNewGuardianKeyGenerator(t *testing.T) {
	t.Parallel()

	t.Run("invalid base key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.BaseKey = ""
		kg, err := NewGuardianKeyGenerator(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "base key"))
		assert.True(t, check.IfNil(kg))
	})
	t.Run("nil key gen should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeyGen = nil
		kg, err := NewGuardianKeyGenerator(args)
		assert.Equal(t, err, ErrNilKeyGenerator)
		assert.True(t, check.IfNil(kg))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		kg, err := NewGuardianKeyGenerator(createMockArgs())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(kg))
	})
}

func TestGuardianKeyGenerator_GenerateKeys(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("KeyGen fails first time", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeyGen = &mock.KeyGenMock{
			PrivateKeyFromByteArrayMock: func(b []byte) (crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}

		kg, _ := NewGuardianKeyGenerator(args)
		assert.False(t, check.IfNil(kg))

		keys, err := kg.GenerateKeys(0)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, keys)
	})
	t.Run("KeyGen fails second time", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		counter := 0
		args.KeyGen = &mock.KeyGenMock{
			PrivateKeyFromByteArrayMock: func(b []byte) (crypto.PrivateKey, error) {
				counter++
				if counter > 1 {
					return nil, expectedErr
				}
				return nil, nil
			},
		}

		kg, _ := NewGuardianKeyGenerator(args)
		assert.False(t, check.IfNil(kg))

		keys, err := kg.GenerateKeys(0)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, keys)
	})
	t.Run("should work with real components", func(t *testing.T) {
		t.Parallel()

		kg, _ := NewGuardianKeyGenerator(ArgGuardianKeyGenerator{
			BaseKey: "moral volcano peasant pass circle pen over picture flat shop clap goat never lyrics gather prepare woman film husband gravity behind test tiger improve",
			KeyGen:  signing.NewKeyGenerator(ed25519.NewEd25519()),
		})
		assert.False(t, check.IfNil(kg))

		numSteps := 100
		for i := 0; i < numSteps; i++ {
			keysFirstTry, err := kg.GenerateKeys(0)
			assert.Nil(t, err)
			assert.NotNil(t, keysFirstTry)
			firstKeyBytes, _ := keysFirstTry[0].ToByteArray()
			assert.Equal(t, "413f42575f7f26fad3317a778771212fdb80245850981e48b58a4f25e344e8f90139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1", hex.EncodeToString(firstKeyBytes))
			secondKeyBytes, _ := keysFirstTry[1].ToByteArray()
			assert.Equal(t, "b8ca6f8203fb4b545a8e83c5384da033c415db155b53fb5b8eba7ff5a039d6398049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f8", hex.EncodeToString(secondKeyBytes))
		}
	})
}
