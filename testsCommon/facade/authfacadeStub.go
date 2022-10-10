package facade

import (
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// FacadeStub -
type FacadeStub struct {
	VerifyCodesCalled        func(request requests.VerifyCodes) error
	RegisterUserCalled       func(request requests.Register) ([]byte, error)
	RestApiInterfaceCalled   func() string
	PprofEnabledCalled       func() bool
	GetGuardianAddressCalled func(request requests.GetGuardianAddress) (string, error)
	SendTransactionCalled    func(request requests.SendTransaction) ([]byte, error)
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

// VerifyCodes -
func (stub *FacadeStub) VerifyCodes(request requests.VerifyCodes) error {
	if stub.VerifyCodesCalled != nil {
		return stub.VerifyCodesCalled(request)
	}
	return nil
}

// RegisterUser -
func (stub *FacadeStub) RegisterUser(request requests.Register) ([]byte, error) {
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

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
