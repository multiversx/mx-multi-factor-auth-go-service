package testsCommon

import crypto "github.com/ElrondNetwork/elrond-go-crypto"

// KeysGeneratorStub -
type KeysGeneratorStub struct {
	GenerateKeysCalled func(index uint32) ([]crypto.PrivateKey, error)
}

// GenerateKeys -
func (stub *KeysGeneratorStub) GenerateKeys(index uint32) ([]crypto.PrivateKey, error) {
	if stub.GenerateKeysCalled != nil {
		return stub.GenerateKeysCalled(index)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *KeysGeneratorStub) IsInterfaceNil() bool {
	return stub == nil
}
