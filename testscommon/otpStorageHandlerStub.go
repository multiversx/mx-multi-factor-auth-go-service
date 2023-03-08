package testscommon

import "github.com/multiversx/multi-factor-auth-go-service/handlers"

// OTPStorageHandlerStub -
type OTPStorageHandlerStub struct {
	SaveCalled func(account, guardian []byte, otp handlers.OTP) error
	GetCalled  func(account, guardian []byte) (handlers.OTP, error)
}

// Save -
func (stub *OTPStorageHandlerStub) Save(account, guardian []byte, otp handlers.OTP) error {
	if stub.SaveCalled != nil {
		return stub.SaveCalled(account, guardian, otp)
	}
	return nil
}

// Get -
func (stub *OTPStorageHandlerStub) Get(account, guardian []byte) (handlers.OTP, error) {
	if stub.GetCalled != nil {
		return stub.GetCalled(account, guardian)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *OTPStorageHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
