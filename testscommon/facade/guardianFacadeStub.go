package facade

import (
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-sdk-go/core"
)

// GuardianFacadeStub -
type GuardianFacadeStub struct {
	VerifyCodeCalled               func(userAddress core.AddressHandler, request requests.VerificationPayload) error
	RegisterUserCalled             func(userAddress core.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error)
	SignTransactionCalled          func(userAddress core.AddressHandler, request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactionsCalled func(userAddress core.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error)
	RegisteredUsersCalled          func() (uint32, error)
}

// VerifyCode -
func (stub *GuardianFacadeStub) VerifyCode(userAddress core.AddressHandler, request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(userAddress, request)
	}
	return nil
}

// RegisterUser -
func (stub *GuardianFacadeStub) RegisterUser(userAddress core.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(userAddress, request)
	}
	return make([]byte, 0), "", nil
}

// SignTransaction -
func (stub *GuardianFacadeStub) SignTransaction(userAddress core.AddressHandler, request requests.SignTransaction) ([]byte, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(userAddress, request)
	}
	return make([]byte, 0), nil
}

// SignMultipleTransactions -
func (stub *GuardianFacadeStub) SignMultipleTransactions(userAddress core.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(userAddress, request)
	}
	return make([][]byte, 0), nil
}

// RegisteredUsers -
func (stub *GuardianFacadeStub) RegisteredUsers() (uint32, error) {
	if stub.RegisteredUsersCalled != nil {
		return stub.RegisteredUsersCalled()
	}
	return 0, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *GuardianFacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
