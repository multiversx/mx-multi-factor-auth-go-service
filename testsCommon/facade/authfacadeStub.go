package facade

import (
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// FacadeStub -
type FacadeStub struct {
	ValidateAndSendCalled    func(request requests.SendTransaction) (string, error)
	RegisterUserCalled       func(request requests.Register) ([]byte, error)
	RestApiInterfaceCalled   func() string
	PprofEnabledCalled       func() bool
	GetGuardianAddressCalled func() string
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

// ValidateAndSend -
func (stub *FacadeStub) ValidateAndSend(request requests.SendTransaction) (string, error) {
	if stub.ValidateAndSendCalled != nil {
		return stub.ValidateAndSendCalled(request)
	}
	return "", nil
}

// RegisterUser -
func (stub *FacadeStub) RegisterUser(request requests.Register) ([]byte, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(request)
	}
	return make([]byte, 0), nil
}

// GetGuardianAddress -
func (stub *FacadeStub) GetGuardianAddress() string {
	if stub.GetGuardianAddressCalled != nil {
		return stub.GetGuardianAddressCalled()
	}
	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
