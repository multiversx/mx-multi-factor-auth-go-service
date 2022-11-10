package testscommon

import "github.com/ElrondNetwork/elrond-sdk-erdgo/core"

// CredentialsHandlerStub -
type CredentialsHandlerStub struct {
	VerifyCalled            func(credentials string) error
	GetAccountAddressCalled func(credentials string) (core.AddressHandler, error)
}

// Verify -
func (stub *CredentialsHandlerStub) Verify(credentials string) error {
	if stub.VerifyCalled != nil {
		return stub.VerifyCalled(credentials)
	}
	return nil
}

// GetAccountAddress -
func (stub *CredentialsHandlerStub) GetAccountAddress(credentials string) (core.AddressHandler, error) {
	if stub.GetAccountAddressCalled != nil {
		return stub.GetAccountAddressCalled(credentials)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *CredentialsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
