package testscommon

import (
	"github.com/multiversx/mx-sdk-go/core"

	tcsCore "github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

// ServiceResolverStub -
type ServiceResolverStub struct {
	GetGuardianAddressCalled        func(userAddress core.AddressHandler) (string, error)
	RegisterUserCalled              func(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error)
	VerifyCodeCalled                func(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error)
	SetSecurityModeNoExpireCalled   func(userIp string, request requests.SetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error)
	UnsetSecurityModeNoExpireCalled func(userIp string, request requests.UnsetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error)
	SignMessageCalled               func(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error)
	SignTransactionCalled           func(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error)
	SignMultipleTransactionsCalled  func(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error)
	RegisteredUsersCalled           func() (uint32, error)
	TcsConfigCalled                 func() *tcsCore.TcsConfig
}

// RegisterUser -
func (stub *ServiceResolverStub) RegisterUser(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(userAddress, request)
	}
	return &requests.OTP{}, "", nil
}

// VerifyCode -
func (stub *ServiceResolverStub) VerifyCode(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error) {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(userAddress, userIp, request)
	}
	return nil, nil
}

// SignMessage -
func (stub *ServiceResolverStub) SignMessage(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error) {
	if stub.SignMessageCalled != nil {
		return stub.SignMessageCalled(userIp, request)
	}
	return make([]byte, 0), nil, nil
}

// SetSecurityModeNoExpire -
func (stub *ServiceResolverStub) SetSecurityModeNoExpire(userIp string, request requests.SetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error) {
	if stub.SetSecurityModeNoExpireCalled != nil {
		return stub.SetSecurityModeNoExpireCalled(userIp, request)
	}
	return nil, nil
}

// UnsetSecurityModeNoExpire -
func (stub *ServiceResolverStub) UnsetSecurityModeNoExpire(userIp string, request requests.UnsetSecurityModeNoExpireMessage) (*requests.OTPCodeVerifyData, error) {
	if stub.UnsetSecurityModeNoExpireCalled != nil {
		return stub.UnsetSecurityModeNoExpireCalled(userIp, request)
	}
	return nil, nil
}

// SignTransaction -
func (stub *ServiceResolverStub) SignTransaction(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(userIp, request)
	}
	return make([]byte, 0), nil, nil
}

// SignMultipleTransactions -
func (stub *ServiceResolverStub) SignMultipleTransactions(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(userIp, request)
	}
	return make([][]byte, 0), nil, nil
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
