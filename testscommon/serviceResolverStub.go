package testscommon

import "github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"

// ServiceResolverStub -
type ServiceResolverStub struct {
	GetGuardianAddressCalled func(request requests.GetGuardianAddress) (string, error)
	RegisterUserCalled       func(request requests.RegistrationPayload) ([]byte, error)
	VerifyCodeCalled         func(request requests.VerificationPayload) error
}

// GetGuardianAddress -
func (stub *ServiceResolverStub) GetGuardianAddress(request requests.GetGuardianAddress) (string, error) {
	if stub.GetGuardianAddressCalled != nil {
		return stub.GetGuardianAddressCalled(request)
	}
	return "", nil
}

// RegisterUser -
func (stub *ServiceResolverStub) RegisterUser(request requests.RegistrationPayload) ([]byte, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(request)
	}
	return make([]byte, 0), nil
}

// VerifyCode -
func (stub *ServiceResolverStub) VerifyCode(request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(request)
	}
	return nil
}

// IsInterfaceNil -
func (stub *ServiceResolverStub) IsInterfaceNil() bool {
	return stub == nil
}
