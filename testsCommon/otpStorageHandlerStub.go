package testsCommon

import "github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"

// OTPStorageHandlerStub -
type OTPStorageHandlerStub struct {
	SaveCalled func(account, guardian string, otp handlers.OTP) error
	GetCalled  func(account, guardian string) (handlers.OTP, error)
}

// Save -
func (stub *OTPStorageHandlerStub) Save(account, guardian string, otp handlers.OTP) error {
	if stub.SaveCalled != nil {
		return stub.SaveCalled(account, guardian, otp)
	}
	return nil
}

// Get -
func (stub *OTPStorageHandlerStub) Get(account, guardian string) (handlers.OTP, error) {
	if stub.GetCalled != nil {
		return stub.GetCalled(account, guardian)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *OTPStorageHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
