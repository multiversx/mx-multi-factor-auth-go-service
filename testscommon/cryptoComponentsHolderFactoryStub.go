package testscommon

import "github.com/ElrondNetwork/elrond-sdk-erdgo/core"

// CryptoComponentsHolderFactoryStub -
type CryptoComponentsHolderFactoryStub struct {
	CreateCalled func(privateKeyBytes []byte) (core.CryptoComponentsHolder, error)
}

// Create -
func (stub *CryptoComponentsHolderFactoryStub) Create(privateKeyBytes []byte) (core.CryptoComponentsHolder, error) {
	if stub.CreateCalled != nil {
		return stub.CreateCalled(privateKeyBytes)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *CryptoComponentsHolderFactoryStub) IsInterfaceNil() bool {
	return stub == nil
}
