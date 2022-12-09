package facade

import (
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// FacadeStub -
type FacadeStub struct {
	VerifyCodeCalled               func(userAddress erdCore.AddressHandler, request requests.VerificationPayload) error
	RegisterUserCalled             func(userAddress erdCore.AddressHandler) ([]byte, string, error)
	RestApiInterfaceCalled         func() string
	PprofEnabledCalled             func() bool
	SignTransactionCalled          func(userAddress erdCore.AddressHandler, request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactionsCalled func(userAddress erdCore.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error)
}

// RestApiInterface -
func (stub *FacadeStub) RestApiInterface() string {
	if stub.RestApiInterfaceCalled != nil {
		return stub.RestApiInterfaceCalled()
	}
	return "localhost:8080"
}

// PprofEnabled -
func (stub *FacadeStub) PprofEnabled() bool {
	if stub.PprofEnabledCalled != nil {
		return stub.PprofEnabledCalled()
	}
	return false
}

// VerifyCode -
func (stub *FacadeStub) VerifyCode(userAddress erdCore.AddressHandler, request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(userAddress, request)
	}
	return nil
}

// RegisterUser -
func (stub *FacadeStub) RegisterUser(userAddress erdCore.AddressHandler) ([]byte, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(userAddress)
	}
	return make([]byte, 0), "", nil
}

// SignTransaction -
func (stub *FacadeStub) SignTransaction(userAddress erdCore.AddressHandler, request requests.SignTransaction) ([]byte, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(userAddress, request)
	}
	return make([]byte, 0), nil
}

// SignMultipleTransactions -
func (stub *FacadeStub) SignMultipleTransactions(userAddress erdCore.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(userAddress, request)
	}
	return make([][]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
