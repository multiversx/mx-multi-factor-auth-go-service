package testscommon

import "github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"

// SecureOtpHandlerStub is a stub implementation of the SecureOtpHandler interface
type SecureOtpHandlerStub struct {
	IsVerificationAllowedAndIncreaseTrialsCalled func(account string, ip string) (*requests.OTPCodeVerifyData, error)
	ResetCalled                                  func(account string, ip string)
	DecrementSecurityModeFailedTrialsCalled      func(account string) error
	FreezeBackoffTimeCalled                      func() uint64
	FreezeMaxFailuresCalled                      func() uint64
	SecurityModeBackOffTimeCalled                func() uint64
	SecurityModeMaxFailuresCalled                func() uint64
}

// IsVerificationAllowedAndIncreaseTrials returns true if the verification is allowed for the given account and ip
func (stub *SecureOtpHandlerStub) IsVerificationAllowedAndIncreaseTrials(account string, ip string) (*requests.OTPCodeVerifyData, error) {
	if stub.IsVerificationAllowedAndIncreaseTrialsCalled != nil {
		return stub.IsVerificationAllowedAndIncreaseTrialsCalled(account, ip)
	}

	return &requests.OTPCodeVerifyData{}, nil
}

// SetSecurityModeNoExpire -
func (stub *SecureOtpHandlerStub) SetSecurityModeNoExpire(key string) error {
	return nil
}

// UnsetSecurityModeNoExpire -
func (stub *SecureOtpHandlerStub) UnsetSecurityModeNoExpire(key string) error {
	return nil
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

// FreezeBackOffTime returns the configured back off time
func (stub *SecureOtpHandlerStub) FreezeBackOffTime() uint64 {
	if stub.FreezeBackoffTimeCalled != nil {
		return stub.FreezeBackoffTimeCalled()
	}
	return 0
}

// FreezeMaxFailures -
func (stub *SecureOtpHandlerStub) FreezeMaxFailures() uint64 {
	if stub.FreezeMaxFailuresCalled != nil {
		return stub.FreezeMaxFailuresCalled()
	}

	return 0
}

// SecurityModeBackOffTime -
func (stub *SecureOtpHandlerStub) SecurityModeBackOffTime() uint64 {
	if stub.SecurityModeBackOffTimeCalled != nil {
		return stub.SecurityModeBackOffTimeCalled()
	}

	return 0
}

// SecurityModeMaxFailures -
func (stub *SecureOtpHandlerStub) SecurityModeMaxFailures() uint64 {
	if stub.SecurityModeMaxFailuresCalled != nil {
		return stub.SecurityModeMaxFailuresCalled()
	}
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *SecureOtpHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
