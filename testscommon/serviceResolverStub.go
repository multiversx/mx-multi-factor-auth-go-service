package testscommon

import "github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"

// ServiceResolverStub -
type ServiceResolverStub struct {
	GetGuardianAddressCalled       func(request requests.GetGuardianAddress) (string, error)
	RegisterUserCalled             func(request requests.RegistrationPayload) ([]byte, error)
	VerifyCodeCalled               func(request requests.VerificationPayload) error
	SendTransactionCalled          func(request requests.SendTransaction) ([]byte, error)
	SendMultipleTransactionsCalled func(request requests.SendMultipleTransaction) ([][]byte, error)
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

// SendTransaction -
func (stub *ServiceResolverStub) SendTransaction(request requests.SendTransaction) ([]byte, error) {
	if stub.SendTransactionCalled != nil {
		return stub.SendTransactionCalled(request)
	}
	return make([]byte, 0), nil
}

// SendMultipleTransactions -
func (stub *ServiceResolverStub) SendMultipleTransactions(request requests.SendMultipleTransaction) ([][]byte, error) {
	if stub.SendMultipleTransactionsCalled != nil {
		return stub.SendMultipleTransactionsCalled(request)
	}
	return make([][]byte, 0), nil
}

// IsInterfaceNil -
func (stub *ServiceResolverStub) IsInterfaceNil() bool {
	return stub == nil
}
