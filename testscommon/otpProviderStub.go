package testscommon

import (
	"crypto"

	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
)

// OTPProviderStub -
type OTPProviderStub struct {
	GenerateTOTPCalled  func(account string, hash crypto.Hash) (handlers.OTP, error)
	TOTPFromBytesCalled func(encryptedMessage []byte) (handlers.OTP, error)
}

// GenerateTOTP -
func (stub *OTPProviderStub) GenerateTOTP(account string, hash crypto.Hash) (handlers.OTP, error) {
	if stub.GenerateTOTPCalled != nil {
		return stub.GenerateTOTPCalled(account, hash)
	}
	return nil, nil
}

// TOTPFromBytes -
func (stub *OTPProviderStub) TOTPFromBytes(encryptedMessage []byte) (handlers.OTP, error) {
	if stub.TOTPFromBytesCalled != nil {
		return stub.TOTPFromBytesCalled(encryptedMessage)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *OTPProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
