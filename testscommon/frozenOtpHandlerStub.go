package testscommon

import "github.com/multiversx/multi-factor-auth-go-service/core/requests"

// FrozenOtpHandlerStub is a stub implementation of the FrozenOtpHandler interface
type FrozenOtpHandlerStub struct {
	IsVerificationAllowedAndDecreaseTrialsCalled func(account string, ip string) (*requests.OTPCodeVerifyData, bool, error)
	ResetCalled                                  func(account string, ip string)
	BackoffTimeCalled                            func() uint64
	MaxFailuresCalled                            func() uint64
}

// IsVerificationAllowedAndDecreaseTrials returns true if the verification is allowed for the given account and ip
func (stub *FrozenOtpHandlerStub) IsVerificationAllowedAndDecreaseTrials(account string, ip string) (*requests.OTPCodeVerifyData, bool, error) {
	if stub.IsVerificationAllowedAndDecreaseTrialsCalled != nil {
		return stub.IsVerificationAllowedAndDecreaseTrialsCalled(account, ip)
	}

	return nil, true, nil
}

// Reset removes the account and ip from local cache
func (stub *FrozenOtpHandlerStub) Reset(account string, ip string) {
	if stub.ResetCalled != nil {
		stub.ResetCalled(account, ip)
	}
}

// BackOffTime returns the configured back off time
func (stub *FrozenOtpHandlerStub) BackOffTime() uint64 {
	if stub.BackoffTimeCalled != nil {
		return stub.BackoffTimeCalled()
	}
	return 0
}

// MaxFailures -
func (stub *FrozenOtpHandlerStub) MaxFailures() uint64 {
	if stub.MaxFailuresCalled != nil {
		return stub.MaxFailuresCalled()
	}

	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FrozenOtpHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
