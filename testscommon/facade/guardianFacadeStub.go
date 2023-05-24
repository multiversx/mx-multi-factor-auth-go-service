package facade

import (
	tcsCore "github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-sdk-go/core"
)

// GuardianFacadeStub -
type GuardianFacadeStub struct {
	VerifyCodeCalled               func(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) error
	RegisterUserCalled             func(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error)
	SignTransactionCalled          func(userIp string, request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactionsCalled func(userIp string, request requests.SignMultipleTransactions) ([][]byte, error)
	RegisteredUsersCalled          func() (uint32, error)
	GetMetricsCalled               func() map[string]*requests.EndpointMetricsResponse
	GetMetricsForPrometheusCalled  func() string
}

// VerifyCode -
func (stub *GuardianFacadeStub) VerifyCode(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) error {
	if stub.VerifyCodeCalled != nil {
		return stub.VerifyCodeCalled(userAddress, userIp, request)
	}
	return nil
}

// RegisterUser -
func (stub *GuardianFacadeStub) RegisterUser(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
	if stub.RegisterUserCalled != nil {
		return stub.RegisterUserCalled(userAddress, request)
	}
	return &requests.OTP{}, "", nil
}

// SignTransaction -
func (stub *GuardianFacadeStub) SignTransaction(userIp string, request requests.SignTransaction) ([]byte, error) {
	if stub.SignTransactionCalled != nil {
		return stub.SignTransactionCalled(userIp, request)
	}
	return make([]byte, 0), nil
}

// SignMultipleTransactions -
func (stub *GuardianFacadeStub) SignMultipleTransactions(userIp string, request requests.SignMultipleTransactions) ([][]byte, error) {
	if stub.SignMultipleTransactionsCalled != nil {
		return stub.SignMultipleTransactionsCalled(userIp, request)
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

// TcsConfig returns the current configuration of the TCS
func (stub *GuardianFacadeStub) TcsConfig() *tcsCore.TcsConfig {
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
