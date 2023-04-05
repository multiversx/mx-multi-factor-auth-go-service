package testscommon

// FrozenOtpHandlerStub is a stub implementation of the FrozenOtpHandler interface
type FrozenOtpHandlerStub struct {
	IncrementFailuresCalled     func(account []byte, ip string)
	IsVerificationAllowedCalled func(account []byte, ip string) bool
}

// IncrementFailures increments the number of verification failures for the given account and ip
func (f *FrozenOtpHandlerStub) IncrementFailures(account []byte, ip string) {
	if f.IncrementFailuresCalled != nil {
		f.IncrementFailuresCalled(account, ip)
	}
}

// IsVerificationAllowed returns true if the verification is allowed for the given account and ip
func (f *FrozenOtpHandlerStub) IsVerificationAllowed(account []byte, ip string) bool {
	if f.IsVerificationAllowedCalled != nil {
		return f.IsVerificationAllowedCalled(account, ip)
	}

	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *FrozenOtpHandlerStub) IsInterfaceNil() bool {
	return f == nil
}
