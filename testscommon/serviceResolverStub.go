package testscommon

import "github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"

// ServiceResolverStub -
type ServiceResolverStub struct {
	GetGuardianAddressCalled       func(request requests.GetGuardianAddress) (string, error)
	RegisterUserCalled             func(request requests.RegistrationPayload) ([]byte, error)
	VerifyCodeCalled               func(request requests.VerificationPayload) error
	SignTransactionCalled          func(request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactionsCalled func(request requests.SignMultipleTransactions) ([][]byte, error)
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

// SignTransaction -
func (stub *ServiceResolverStub) SignTransaction(request requests.SignTransaction) ([]byte, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(request)
	}
	return make([]byte, 0), nil
}

// SignMultipleTransactions -
func (stub *ServiceResolverStub) SignMultipleTransactions(request requests.SignMultipleTransactions) ([][]byte, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(request)
	}
	return make([][]byte, 0), nil
}

// IsInterfaceNil -
func (stub *ServiceResolverStub) IsInterfaceNil() bool {
	return stub == nil
}
