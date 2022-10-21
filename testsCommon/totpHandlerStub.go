package testsCommon

import (
	"crypto"

	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
)

// TOTPHandlerStub -
type TOTPHandlerStub struct {
	CreateTOTPCalled    func(account string, hash crypto.Hash) (handlers.OTP, error)
	TOTPFromBytesCalled func(encryptedMessage []byte) (handlers.OTP, error)
}

// CreateTOTP -
func (stub *TOTPHandlerStub) CreateTOTP(account string, hash crypto.Hash) (handlers.OTP, error) {
	if stub.CreateTOTPCalled != nil {
		return stub.CreateTOTPCalled(account, hash)
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
