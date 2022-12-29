package facade

import (
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// AuthFacadeStub -
type AuthFacadeStub struct {
	VerifyCodeCalled               func(request requests.VerificationPayload) error
	RegisterUserCalled             func(request requests.RegistrationPayload) ([]byte, string, error)
	RestApiInterfaceCalled         func() string
	PprofEnabledCalled             func() bool
	SignTransactionCalled          func(request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactionsCalled func(request requests.SignMultipleTransactions) ([][]byte, error)
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
func (stub *AuthFacadeStub) VerifyCode(request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(request)
	}
	return nil
}

// RegisterUser -
func (stub *AuthFacadeStub) RegisterUser(request requests.RegistrationPayload) ([]byte, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(request)
	}
	return make([]byte, 0), "", nil
}

// SignTransaction -
func (stub *AuthFacadeStub) SignTransaction(request requests.SignTransaction) ([]byte, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(request)
	}
	return make([]byte, 0), nil
}

// SignMultipleTransactions -
func (stub *AuthFacadeStub) SignMultipleTransactions(request requests.SignMultipleTransactions) ([][]byte, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(request)
	}
	return make([][]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *AuthFacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
