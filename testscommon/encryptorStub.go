package testscommon

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
	return data, nil
}

// DecryptData decrypts the provided data
func (es *EncryptorStub) DecryptData(data []byte) ([]byte, error) {
	if es.DecryptDataCalled != nil {
		return es.DecryptDataCalled(data)
	}
	return data, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (es *EncryptorStub) IsInterfaceNil() bool {
	return es == nil
}
