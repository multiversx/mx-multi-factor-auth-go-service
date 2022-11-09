package facade

import (
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// FacadeStub -
type FacadeStub struct {
	VerifyCodeCalled               func(request requests.VerificationPayload) error
	RegisterUserCalled             func(request requests.RegistrationPayload) ([]byte, error)
	RestApiInterfaceCalled         func() string
	PprofEnabledCalled             func() bool
	GetGuardianAddressCalled       func(request requests.GetGuardianAddress) (string, error)
	SendTransactionCalled          func(request requests.SendTransaction) ([]byte, error)
	SendMultipleTransactionsCalled func(request requests.SendMultipleTransaction)
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
func (stub *FacadeStub) VerifyCode(request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(request)
	}
	return nil
}

// RegisterUser -
func (stub *FacadeStub) RegisterUser(request requests.RegistrationPayload) ([]byte, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(request)
	}
	return make([]byte, 0), nil
}

// GetGuardianAddress -
func (stub *FacadeStub) GetGuardianAddress(request requests.GetGuardianAddress) (string, error) {
	if stub.GetGuardianAddressCalled != nil {
		return stub.GetGuardianAddressCalled(request)
	}
	return "", nil
}

// SendTransaction -
func (stub *FacadeStub) SendTransaction(request requests.SendTransaction) ([]byte, error) {
	if stub.SendTransactionCalled != nil {
		return stub.SendTransactionCalled(request)
	}
	return make([]byte, 0), nil
}

// SendMultipleTransactions -
func (stub *FacadeStub) SendMultipleTransactions(request requests.SendMultipleTransaction) ([][]byte, error) {
	if stub.SendMultipleTransactionsCalled != nil {
		return stub.SendMultipleTransactionsCalled(request)
	}
	return make([][]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
