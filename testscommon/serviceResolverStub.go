package testscommon

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// ServiceResolverStub -
type ServiceResolverStub struct {
	GetGuardianAddressCalled       func(userAddress core.AddressHandler) (string, error)
	RegisterUserCalled             func(userAddress core.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error)
	VerifyCodeCalled               func(userAddress core.AddressHandler, request requests.VerificationPayload) error
	SignTransactionCalled          func(userAddress core.AddressHandler, request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactionsCalled func(userAddress core.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error)
}

// GetGuardianAddress -
func (stub *ServiceResolverStub) GetGuardianAddress(userAddress core.AddressHandler) (string, error) {
	if stub.GetGuardianAddressCalled != nil {
		return stub.GetGuardianAddressCalled(userAddress)
	}

	return "", nil
}

// RegisterUser -
func (stub *ServiceResolverStub) RegisterUser(userAddress core.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(userAddress, request)
	}
	return make([]byte, 0), "", nil
}

// VerifyCode -
func (stub *ServiceResolverStub) VerifyCode(userAddress core.AddressHandler, request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(userAddress, request)
	}
	return nil
}

// SignTransaction -
func (stub *ServiceResolverStub) SignTransaction(userAddress core.AddressHandler, request requests.SignTransaction) ([]byte, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(userAddress, request)
	}
	return make([]byte, 0), nil
}

// SignMultipleTransactions -
func (stub *ServiceResolverStub) SignMultipleTransactions(userAddress core.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(userAddress, request)
	}
	return make([][]byte, 0), nil
}

// IsInterfaceNil -
func (stub *ServiceResolverStub) IsInterfaceNil() bool {
	return stub == nil
}
