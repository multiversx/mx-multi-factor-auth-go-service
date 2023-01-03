package facade

import (
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// AuthFacadeStub -
type AuthFacadeStub struct {
	VerifyCodeCalled               func(userAddress erdCore.AddressHandler, request requests.VerificationPayload) error
	RegisterUserCalled             func(userAddress erdCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error)
	RestApiInterfaceCalled         func() string
	PprofEnabledCalled             func() bool
	SignTransactionCalled          func(userAddress erdCore.AddressHandler, request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactionsCalled func(userAddress erdCore.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error)
}

// RestApiInterface -
func (stub *AuthFacadeStub) RestApiInterface() string {
	if stub.RestApiInterfaceCalled != nil {
		return stub.RestApiInterfaceCalled()
	}
	return "localhost:8080"
}

// PprofEnabled -
func (stub *AuthFacadeStub) PprofEnabled() bool {
	if stub.PprofEnabledCalled != nil {
		return stub.PprofEnabledCalled()
	}
	return false
}

// VerifyCode -
func (stub *AuthFacadeStub) VerifyCode(userAddress erdCore.AddressHandler, request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(userAddress, request)
	}
	return nil
}

// RegisterUser -
func (stub *AuthFacadeStub) RegisterUser(userAddress erdCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(userAddress, request)
	}
	return make([]byte, 0), "", nil
}

// SignTransaction -
func (stub *AuthFacadeStub) SignTransaction(userAddress erdCore.AddressHandler, request requests.SignTransaction) ([]byte, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(userAddress, request)
	}
	return make([]byte, 0), nil
}

// SignMultipleTransactions -
func (stub *AuthFacadeStub) SignMultipleTransactions(userAddress erdCore.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(userAddress, request)
	}
	return make([][]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *AuthFacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
