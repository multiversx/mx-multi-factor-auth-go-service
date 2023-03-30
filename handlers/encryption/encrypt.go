package encryption

import (
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/encryption/x25519"
)

type encryptor struct {
	encryptionMarshaller core.Marshaller
	keyGen               crypto.KeyGenerator
	managedPrivateKey    crypto.PrivateKey
	publicKey            crypto.PublicKey
}

// NewEncryptor creates a new encryptor instance
func NewEncryptor(encryptionMarshaller core.Marshaller, keyGen crypto.KeyGenerator, managedPrivateKey crypto.PrivateKey) (*encryptor, error) {
	if check.IfNil(encryptionMarshaller) {
		return nil, ErrNilMarshaller
	}
	if check.IfNil(keyGen) {
		return nil, ErrNilKeyGenerator
	}
	if check.IfNil(managedPrivateKey) {
		return nil, ErrNilPrivateKey
	}

	return &encryptor{
		encryptionMarshaller: encryptionMarshaller,
		keyGen:               keyGen,
		managedPrivateKey:    managedPrivateKey,
		publicKey:            managedPrivateKey.GeneratePublic(),
	}, nil
}

// EncryptData encrypts the provided data
func (enc *encryptor) EncryptData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	encryptionSk, _ := enc.keyGen.GeneratePair()
	encryptedData := &x25519.EncryptedData{}
	err := encryptedData.Encrypt(data, enc.publicKey, encryptionSk)
	if err != nil {
		return nil, err
	}

	return enc.encryptionMarshaller.Marshal(encryptedData)
}

// DecryptData decrypts the provided data
func (enc *encryptor) DecryptData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	encryptedData := &x25519.EncryptedData{}
	err := enc.encryptionMarshaller.Unmarshal(encryptedData, data)
	if err != nil {
		return nil, err
	}

	return encryptedData.Decrypt(enc.managedPrivateKey)
}

// IsInterfaceNil returns true if there is no value under the interface
func (enc *encryptor) IsInterfaceNil() bool {
	return enc == nil
}
