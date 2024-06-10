package facade

import (
	"github.com/multiversx/mx-sdk-go/core"

	tcsCore "github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

// GuardianFacadeStub -
type GuardianFacadeStub struct {
	VerifyCodeCalled               func(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error)
	RegisterUserCalled             func(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error)
	SignMessageCalled              func(userAddress core.AddressHandler, request requests.SignMessage) ([]byte, error)
	SignTransactionCalled          func(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error)
	SignMultipleTransactionsCalled func(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error)
	RegisteredUsersCalled          func() (uint32, error)
	GetMetricsCalled               func() map[string]*requests.EndpointMetricsResponse
	GetMetricsForPrometheusCalled  func() string
	TcsConfigCalled                func() *tcsCore.TcsConfig
}

// VerifyCode -
func (stub *GuardianFacadeStub) VerifyCode(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error) {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(userAddress, userIp, request)
	}
	return nil, nil
}

// RegisterUser -
func (stub *GuardianFacadeStub) RegisterUser(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(userAddress, request)
	}
	return &requests.OTP{}, "", nil
}

// SignMessage -
func (stub *GuardianFacadeStub) SignMessage(userAddress core.AddressHandler, request requests.SignMessage) ([]byte, error) {
	if stub.SignMessageCalled != nil {
		return stub.SignMessageCalled(userAddress, request)
	}
	return make([]byte, 0), nil
}

// SignTransaction -
func (stub *GuardianFacadeStub) SignTransaction(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(userIp, request)
	}
	return make([]byte, 0), nil, nil
}

// SignMultipleTransactions -
func (stub *GuardianFacadeStub) SignMultipleTransactions(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(userIp, request)
	}
	return make([][]byte, 0), nil, nil
}

// RegisteredUsers -
func (stub *GuardianFacadeStub) RegisteredUsers() (uint32, error) {
	if stub.RegisteredUsersCalled != nil {
		return stub.RegisteredUsersCalled()
	}
	return 0, nil
}

// TcsConfig returns the current configuration of the TCS
func (stub *GuardianFacadeStub) TcsConfig() *tcsCore.TcsConfig {
	if stub.TcsConfigCalled != nil {
		return stub.TcsConfigCalled()
	}

	return &tcsCore.TcsConfig{
		OTPDelay:         0,
		BackoffWrongCode: 0,
	}
}

// GetMetrics -
func (stub *GuardianFacadeStub) GetMetrics() map[string]*requests.EndpointMetricsResponse {
	if stub.GetMetricsCalled != nil {
		return stub.GetMetricsCalled()
	}

	return nil
}

// GetMetricsForPrometheus -
func (stub *GuardianFacadeStub) GetMetricsForPrometheus() string {
	if stub.GetMetricsForPrometheusCalled != nil {
		return stub.GetMetricsForPrometheusCalled()
	}

	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *GuardianFacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
