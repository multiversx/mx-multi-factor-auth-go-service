package testscommon

import "github.com/ElrondNetwork/elrond-go-crypto"

// KeysGeneratorStub -
type KeysGeneratorStub struct {
	GenerateManagedKeyCalled func() (crypto.PrivateKey, error)
	GenerateKeysCalled       func(index uint32) ([]crypto.PrivateKey, error)
}

// GenerateManagedKey -
func (stub *KeysGeneratorStub) GenerateManagedKey() (crypto.PrivateKey, error) {
	if stub.GenerateManagedKeyCalled != nil {
		return stub.GenerateManagedKeyCalled()
	}
	return nil, nil
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
