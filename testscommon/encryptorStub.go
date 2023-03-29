package testscommon

// Encryptor is the interface that defines the methods that can be used to encrypt and decrypt data
type Encryptor interface {
	EncryptData(data []byte) ([]byte, error)
	DecryptData(data []byte) ([]byte, error)
	IsInterfaceNil() bool
}

// EncryptorStub is a stub implementation of Encryptor
type EncryptorStub struct {
	EncryptDataCalled func(data []byte) ([]byte, error)
	DecryptDataCalled func(data []byte) ([]byte, error)
}

// EncryptData encrypts the provided data
func (es *EncryptorStub) EncryptData(data []byte) ([]byte, error) {
	if es.EncryptDataCalled != nil {
		return es.EncryptDataCalled(data)
	}
	return nil, nil
}

// DecryptData decrypts the provided data
func (es *EncryptorStub) DecryptData(data []byte) ([]byte, error) {
	if es.DecryptDataCalled != nil {
		return es.DecryptDataCalled(data)
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (es *EncryptorStub) IsInterfaceNil() bool {
	return es == nil
}
