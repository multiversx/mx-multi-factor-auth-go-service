package core

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/mock"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgGuardianKeyGenerator {
	return ArgGuardianKeyGenerator{
		Mnemonic: "mnemonic",
		KeyGen:   &mock.KeyGenMock{},
	}
}

func TestNewGuardianKeyGenerator(t *testing.T) {
	t.Parallel()

	t.Run("invalid mnemonic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Mnemonic = ""
		kg, err := NewGuardianKeyGenerator(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "mnemonic"))
		assert.Nil(t, kg)
	})
	t.Run("nil key gen should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeyGen = nil
		kg, err := NewGuardianKeyGenerator(args)
		assert.Equal(t, err, ErrNilKeyGenerator)
		assert.Nil(t, kg)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		kg, err := NewGuardianKeyGenerator(createMockArgs())
		assert.Nil(t, err)
		assert.NotNil(t, kg)
	})
}

func TestGuardianKeyGenerator_GenerateKeys(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("index 0 should not work", func(t *testing.T) {
		t.Parallel()

		kg, _ := NewGuardianKeyGenerator(createMockArgs())
		assert.NotNil(t, kg)

		keys, err := kg.GenerateKeys(0)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.Nil(t, keys)
	})
	t.Run("KeyGen fails first time", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeyGen = &mock.KeyGenMock{
			PrivateKeyFromByteArrayMock: func(b []byte) (crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}

		kg, _ := NewGuardianKeyGenerator(args)
		assert.NotNil(t, kg)

		keys, err := kg.GenerateKeys(1)
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
		assert.NotNil(t, kg)

		keys, err := kg.GenerateKeys(1)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, keys)
	})
	t.Run("should work with real components", func(t *testing.T) {
		t.Parallel()

		kg, _ := NewGuardianKeyGenerator(ArgGuardianKeyGenerator{
			Mnemonic: "moral volcano peasant pass circle pen over picture flat shop clap goat never lyrics gather prepare woman film husband gravity behind test tiger improve",
			KeyGen:   signing.NewKeyGenerator(ed25519.NewEd25519()),
		})
		assert.NotNil(t, kg)

		numSteps := 100
		for i := 0; i < numSteps; i++ {
			keysFirstTry, err := kg.GenerateKeys(1)
			assert.Nil(t, err)
			assert.NotNil(t, keysFirstTry)
			firstKeyBytes, _ := keysFirstTry[0].ToByteArray()
			assert.Equal(t, "b8ca6f8203fb4b545a8e83c5384da033c415db155b53fb5b8eba7ff5a039d6398049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f8", hex.EncodeToString(firstKeyBytes))
			secondKeyBytes, _ := keysFirstTry[1].ToByteArray()
			assert.Equal(t, "e253a571ca153dc2aee845819f74bcc9773b0586edead15a94cb7235a5027436b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba", hex.EncodeToString(secondKeyBytes))
		}
	})
	t.Run("should work with many calls", func(t *testing.T) {
		t.Parallel()

		kg, _ := NewGuardianKeyGenerator(ArgGuardianKeyGenerator{
			Mnemonic: "moral volcano peasant pass circle pen over picture flat shop clap goat never lyrics gather prepare woman film husband gravity behind test tiger improve",
			KeyGen:   signing.NewKeyGenerator(ed25519.NewEd25519()),
		})
		assert.NotNil(t, kg)

		keysMap := make(map[string]struct{})
		numSteps := 100
		for i := 1; i <= numSteps; i++ {
			keys, err := kg.GenerateKeys(uint32(2 * i))
			assert.Nil(t, err)
			assert.NotNil(t, keys)

			checkKeyAndAddToMap(t, keys[0], keysMap)
			checkKeyAndAddToMap(t, keys[1], keysMap)
		}

		assert.Equal(t, 2*numSteps, len(keysMap))
	})
}

func TestGuardianKeyGenerator_GenerateManagedKey(t *testing.T) {
	t.Parallel()

	kg, err := NewGuardianKeyGenerator(ArgGuardianKeyGenerator{
		Mnemonic: "moral volcano peasant pass circle pen over picture flat shop clap goat never lyrics gather prepare woman film husband gravity behind test tiger improve",
		KeyGen:   signing.NewKeyGenerator(ed25519.NewEd25519()),
	})
	assert.Nil(t, err)
	assert.NotNil(t, kg)

	key, err := kg.GenerateManagedKey()
	assert.Nil(t, err)
	keyBytes, err := key.ToByteArray()
	assert.Nil(t, err)
	assert.Equal(t, "413f42575f7f26fad3317a778771212fdb80245850981e48b58a4f25e344e8f90139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1", hex.EncodeToString(keyBytes))
}

func checkKeyAndAddToMap(t *testing.T, key crypto.PrivateKey, keysMap map[string]struct{}) {
	keyBytes, err := key.ToByteArray()
	assert.Nil(t, err)
	_, exists := keysMap[string(keyBytes)]
	assert.False(t, exists)

	keysMap[string(keyBytes)] = struct{}{}
}
