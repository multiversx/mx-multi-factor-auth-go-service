package testscommon

import (
	tcsCore "github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-sdk-go/core"
)

// ServiceResolverStub -
type ServiceResolverStub struct {
	RegisterUserCalled             func(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error)
	VerifyCodeCalled               func(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) error
	SignTransactionCalled          func(userIp string, request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactionsCalled func(userIp string, request requests.SignMultipleTransactions) ([][]byte, error)
	RegisteredUsersCalled          func() (uint32, error)
	TcsConfigCalled                func() *tcsCore.TcsConfig
}

// RegisterUser -
func (stub *ServiceResolverStub) RegisterUser(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(userAddress, request)
	}
	return &requests.OTP{}, "", nil
}

// VerifyCode -
func (stub *ServiceResolverStub) VerifyCode(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(userAddress, userIp, request)
	}
	return nil
}

// SignTransaction -
func (stub *ServiceResolverStub) SignTransaction(userIp string, request requests.SignTransaction) ([]byte, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(userIp, request)
	}
	return make([]byte, 0), nil
}

// SignMultipleTransactions -
func (stub *ServiceResolverStub) SignMultipleTransactions(userIp string, request requests.SignMultipleTransactions) ([][]byte, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(userIp, request)
	}
	return make([][]byte, 0), nil
}

// RegisteredUsers -
func (stub *ServiceResolverStub) RegisteredUsers() (uint32, error) {
	if stub.RegisteredUsersCalled != nil {
		return stub.RegisteredUsersCalled()
	}
	return 0, nil
}

// TcsConfig returns the current configuration of the TCS
func (stub *ServiceResolverStub) TcsConfig() *tcsCore.TcsConfig {
	if stub.TcsConfigCalled != nil {
		return stub.TcsConfigCalled()
	}

	return &tcsCore.TcsConfig{
		OTPDelay:         0,
		BackoffWrongCode: 0,
	}
}

// IsInterfaceNil -
func (stub *ServiceResolverStub) IsInterfaceNil() bool {
	return stub == nil
}
