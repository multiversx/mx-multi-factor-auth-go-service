package testsCommon

import "github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"

// ServiceResolverStub -
type ServiceResolverStub struct {
	GetGuardianAddressCalled func(request requests.GetGuardianAddress) (string, error)
	RegisterUserCalled       func(request requests.Register) ([]byte, error)
	VerifyCodesCalled        func(request requests.VerifyCodes) error
}

// GetGuardianAddress -
func (stub *ServiceResolverStub) GetGuardianAddress(request requests.GetGuardianAddress) (string, error) {
	if stub.GetGuardianAddressCalled != nil {
		return stub.GetGuardianAddressCalled(request)
	}
	return "", nil
}

// RegisterUser -
func (stub *ServiceResolverStub) RegisterUser(request requests.Register) ([]byte, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(request)
	}
	return make([]byte, 0), nil
}

// VerifyCodes -
func (stub *ServiceResolverStub) VerifyCodes(request requests.VerifyCodes) error {
	if stub.VerifyCodesCalled != nil {
		return stub.VerifyCodesCalled(request)
	}
	return nil
}

// IsInterfaceNil -
func (stub *ServiceResolverStub) IsInterfaceNil() bool {
	return stub == nil
}
