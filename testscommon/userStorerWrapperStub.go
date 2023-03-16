package testscommon

import "github.com/multiversx/multi-factor-auth-go-service/core"

// UserStorerWrapperStub -
type UserStorerWrapperStub struct {
	LoadCalled func(key []byte) (*core.OTPInfo, error)
	SaveCalled func(key []byte, otpInfo *core.OTPInfo) error
}

// Load -
func (u *UserStorerWrapperStub) Load(key []byte) (*core.OTPInfo, error) {
	if u.LoadCalled != nil {
		return u.LoadCalled(key)
	}

	return nil, nil
}

// Save -
func (u *UserStorerWrapperStub) Save(key []byte, otpInfo *core.OTPInfo) error {
	if u.SaveCalled != nil {
		return u.SaveCalled(key, otpInfo)
	}

	return nil
}

// IsInterfaceNil -
func (u *UserStorerWrapperStub) IsInterfaceNil() bool {
	return u == nil
}
