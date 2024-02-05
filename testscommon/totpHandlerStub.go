package testscommon

import (
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
)

// TOTPHandlerStub -
type TOTPHandlerStub struct {
	CreateTOTPCalled    func(account string) (handlers.OTP, error)
	TOTPFromBytesCalled func(encryptedMessage []byte) (handlers.OTP, error)
}

// CreateTOTP -
func (stub *TOTPHandlerStub) CreateTOTP(account string) (handlers.OTP, error) {
	if stub.CreateTOTPCalled != nil {
		return stub.CreateTOTPCalled(account)
	}
	return nil, nil
}

// TOTPFromBytes -
func (stub *TOTPHandlerStub) TOTPFromBytes(encryptedMessage []byte) (handlers.OTP, error) {
	if stub.TOTPFromBytesCalled != nil {
		return stub.TOTPFromBytesCalled(encryptedMessage)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *TOTPHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
