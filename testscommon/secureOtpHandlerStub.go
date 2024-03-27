package testscommon

import "github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"

// SecureOtpHandlerStub is a stub implementation of the SecureOtpHandler interface
type SecureOtpHandlerStub struct {
	IsVerificationAllowedAndIncreaseTrialsCalled func(account string, ip string) (*requests.OTPCodeVerifyData, error)
	ResetCalled                                  func(account string, ip string)
	DecrementSecurityModeFailedTrialsCalled      func(account string) error
	BackoffTimeCalled                            func() uint64
	MaxFailuresCalled                            func() uint64
}

// IsVerificationAllowedAndIncreaseTrials returns true if the verification is allowed for the given account and ip
func (stub *SecureOtpHandlerStub) IsVerificationAllowedAndIncreaseTrials(account string, ip string) (*requests.OTPCodeVerifyData, error) {
	if stub.IsVerificationAllowedAndIncreaseTrialsCalled != nil {
		return stub.IsVerificationAllowedAndIncreaseTrialsCalled(account, ip)
	}

	return &requests.OTPCodeVerifyData{}, nil
}

// Reset removes the account and ip from local cache
func (stub *SecureOtpHandlerStub) Reset(account string, ip string) {
	if stub.ResetCalled != nil {
		stub.ResetCalled(account, ip)
	}
}

// DecrementSecurityModeFailedTrials decrements the security mode failed trials
func (stub *SecureOtpHandlerStub) DecrementSecurityModeFailedTrials(account string) error {
	if stub.DecrementSecurityModeFailedTrialsCalled != nil {
		return stub.DecrementSecurityModeFailedTrialsCalled(account)
	}

	return nil
}

// BackOffTime returns the configured back off time
func (stub *SecureOtpHandlerStub) BackOffTime() uint64 {
	if stub.BackoffTimeCalled != nil {
		return stub.BackoffTimeCalled()
	}
	return 0
}

// MaxFailures -
func (stub *SecureOtpHandlerStub) MaxFailures() uint64 {
	if stub.MaxFailuresCalled != nil {
		return stub.MaxFailuresCalled()
	}

	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *SecureOtpHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
