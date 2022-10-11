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
		MainKey:      "main key",
		SecondaryKey: "secondary key",
		KeyGen:       &mock.KeyGenMock{},
	}
}

func TestNewGuardianKeyGenerator(t *testing.T) {
	t.Parallel()

	t.Run("invalid main key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.MainKey = ""
		kg, err := NewGuardianKeyGenerator(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "main key"))
		assert.True(t, check.IfNil(kg))
	})
	t.Run("invalid secondary key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SecondaryKey = ""
		kg, err := NewGuardianKeyGenerator(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "secondary key"))
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
	t.Run("invalid main key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeyGen = &mock.KeyGenMock{
			PrivateKeyFromByteArrayMock: func(b []byte) (crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}

		kg, err := NewGuardianKeyGenerator(args)
		assert.False(t, check.IfNil(kg))

		keys, err := kg.GenerateKeys(0)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, keys)
	})
	t.Run("invalid secondary key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		numCalls := 0
		args.KeyGen = &mock.KeyGenMock{
			PrivateKeyFromByteArrayMock: func(b []byte) (crypto.PrivateKey, error) {
				if numCalls > 0 {
					return nil, expectedErr
				}
				numCalls++
				return nil, nil
			},
		}

		kg, err := NewGuardianKeyGenerator(args)
		assert.False(t, check.IfNil(kg))

		keys, err := kg.GenerateKeys(0)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, keys)
	})
	t.Run("should work with real components", func(t *testing.T) {
		t.Parallel()

		kg, err := NewGuardianKeyGenerator(ArgGuardianKeyGenerator{
			MainKey:      "acid twice post genre topic observe valid viable gesture fortune funny dawn around blood enemy page update reduce decline van bundle zebra rookie real",
			SecondaryKey: "bid involve twenty cave offer life hello three walnut travel rare bike edit canyon ice brave theme furnace cotton swing wear bread fine latin",
			KeyGen:       signing.NewKeyGenerator(ed25519.NewEd25519()),
		})
		assert.False(t, check.IfNil(kg))

		keysFirstTry, err := kg.GenerateKeys(0)
		assert.Nil(t, err)
		assert.NotNil(t, keysFirstTry)
		mainKeyBytes, _ := keysFirstTry[0].ToByteArray()
		assert.Equal(t, "0b7966138e80b8f3bb64046f56aea4250fd7bacad6ed214165cea6767fd0bc2cdfefe0453840e5903f2bd519de9b0ed6e9621e57e28ba0b4c1b15115091dd72f", hex.EncodeToString(mainKeyBytes))
		secondaryKeyBytes, _ := keysFirstTry[1].ToByteArray()
		assert.Equal(t, "15cfe2140ee9821f706423036ba58d1e6ec13dbc4ebf206732ad40b5236af403be8aa862028f37acd00e12e152487971806761c61759fa4ca03c023e42063a41", hex.EncodeToString(secondaryKeyBytes))

		keySecondTry, err := kg.GenerateKeys(0)
		assert.Nil(t, err)
		assert.NotNil(t, keySecondTry)
		assert.Equal(t, keysFirstTry, keySecondTry)

		keyThirdTry, err := kg.GenerateKeys(1)
		assert.Nil(t, err)
		assert.NotNil(t, keyThirdTry)
		assert.NotEqual(t, keysFirstTry, keyThirdTry)
	})
}
