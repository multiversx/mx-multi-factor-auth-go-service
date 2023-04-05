package encryption

import (
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/stretchr/testify/require"
)

func TestNewEncryptor(t *testing.T) {
	t.Parallel()

	testKeygen := signing.NewKeyGenerator(ed25519.NewEd25519())
	testSk, _ := testKeygen.GeneratePair()
	testMarshaller := &testscommon.MarshallerStub{}

	t.Run("nil marshaller should return error", func(t *testing.T) {
		t.Parallel()

		enc, err := NewEncryptor(nil, testKeygen, testSk)
		require.Nil(t, enc)
		require.Equal(t, ErrNilMarshaller, err)
	})
	t.Run("nil keygen should return error", func(t *testing.T) {
		t.Parallel()

		enc, err := NewEncryptor(testMarshaller, nil, testSk)
		require.Nil(t, enc)
		require.Equal(t, ErrNilKeyGenerator, err)
	})
	t.Run("nil sk should return error", func(t *testing.T) {
		t.Parallel()

		enc, err := NewEncryptor(testMarshaller, testKeygen, nil)
		require.Nil(t, enc)
		require.Equal(t, ErrNilPrivateKey, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		enc, err := NewEncryptor(testMarshaller, testKeygen, testSk)
		require.NotNil(t, enc)
		require.Nil(t, err)
	})
}

func TestEncryptor_Encrypt(t *testing.T) {
	t.Parallel()

	testKeygen := signing.NewKeyGenerator(ed25519.NewEd25519())
	testSk, _ := testKeygen.GeneratePair()
	testMarshaller, _ := factoryMarshaller.NewMarshalizer(factoryMarshaller.JsonMarshalizer)

	t.Run("nil data should not error", func(t *testing.T) {
		t.Parallel()

		enc, _ := NewEncryptor(testMarshaller, signing.NewKeyGenerator(ed25519.NewEd25519()), testSk)
		encData, err := enc.EncryptData(nil)
		require.Nil(t, err)
		require.Nil(t, encData)
	})
	t.Run("empty data should not error", func(t *testing.T) {
		t.Parallel()

		enc, _ := NewEncryptor(testMarshaller, signing.NewKeyGenerator(ed25519.NewEd25519()), testSk)
		emptyData := []byte("")
		encData, err := enc.EncryptData(emptyData)
		require.Nil(t, err)
		require.Equal(t, emptyData, encData)
	})
	t.Run("marshal error should return error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		marshaller := &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}

		enc, _ := NewEncryptor(marshaller, signing.NewKeyGenerator(ed25519.NewEd25519()), testSk)
		encData, err := enc.EncryptData([]byte("data to encrypt"))
		require.Equal(t, expectedErr, err)
		require.Nil(t, encData)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dataToEncrypt := []byte("data to encrypt")
		enc, _ := NewEncryptor(testMarshaller, signing.NewKeyGenerator(ed25519.NewEd25519()), testSk)
		encData, err := enc.EncryptData(dataToEncrypt)
		require.Nil(t, err)
		require.NotNil(t, encData)
		require.NotEqual(t, dataToEncrypt, encData)
	})
}

func TestEncryptor_Decrypt(t *testing.T) {
	t.Parallel()

	testKeygen := signing.NewKeyGenerator(ed25519.NewEd25519())
	testSk, _ := testKeygen.GeneratePair()
	testMarshaller, _ := factoryMarshaller.NewMarshalizer(factoryMarshaller.JsonMarshalizer)

	t.Run("nil data should not error", func(t *testing.T) {
		t.Parallel()

		enc, _ := NewEncryptor(testMarshaller, signing.NewKeyGenerator(ed25519.NewEd25519()), testSk)
		decData, err := enc.DecryptData(nil)
		require.Nil(t, err)
		require.Nil(t, decData)
	})
	t.Run("empty data should not error", func(t *testing.T) {
		t.Parallel()

		enc, _ := NewEncryptor(testMarshaller, signing.NewKeyGenerator(ed25519.NewEd25519()), testSk)
		emptyData := []byte("")
		decData, err := enc.DecryptData(emptyData)
		require.Nil(t, err)
		require.Equal(t, emptyData, decData)
	})
	t.Run("unmarshal error should return error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		marshaller := &testscommon.MarshallerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}

		enc, _ := NewEncryptor(marshaller, signing.NewKeyGenerator(ed25519.NewEd25519()), testSk)
		decData, err := enc.DecryptData([]byte("data to encrypt"))
		require.Equal(t, expectedErr, err)
		require.Nil(t, decData)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dataToEncrypt := []byte("data to encrypt")
		enc, _ := NewEncryptor(testMarshaller, signing.NewKeyGenerator(ed25519.NewEd25519()), testSk)
		encData, _ := enc.EncryptData(dataToEncrypt)
		decData, err := enc.DecryptData(encData)
		require.Nil(t, err)
		require.NotNil(t, decData)
		require.Equal(t, dataToEncrypt, decData)
	})
}
