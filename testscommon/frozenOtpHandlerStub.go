package testscommon

// FrozenOtpHandlerStub is a stub implementation of the FrozenOtpHandler interface
type FrozenOtpHandlerStub struct {
	IncrementFailuresCalled     func(account string, ip string)
	IsVerificationAllowedCalled func(account string, ip string) bool
	ResetCalled                 func(account string, ip string)
	BackoffTimeCalled           func() uint64
}

// IncrementFailures increments the number of verification failures for the given account and ip
func (stub *FrozenOtpHandlerStub) IncrementFailures(account string, ip string) {
	if stub.IncrementFailuresCalled != nil {
		stub.IncrementFailuresCalled(account, ip)
	}
}

// IsVerificationAllowed returns true if the verification is allowed for the given account and ip
func (stub *FrozenOtpHandlerStub) IsVerificationAllowed(account string, ip string) bool {
	if stub.IsVerificationAllowedCalled != nil {
		return stub.IsVerificationAllowedCalled(account, ip)
	}

	return true
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

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FrozenOtpHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
