package testscommon

import (
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/data/mock"
)

type simpleStorer interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
}

// UserStorerWrapperMock -
type UserStorerWrapperMock struct {
	marshaller core.Marshaller
	storer     simpleStorer

	LoadCalled func(key []byte) (*core.OTPInfo, error)
	SaveCalled func(key []byte, otpInfo *core.OTPInfo) error
}

// NewUserStorerWrapperMock -
func NewUserStorerWrapperMock(storer simpleStorer) *UserStorerWrapperMock {
	return &UserStorerWrapperMock{
		marshaller: &mock.MarshalizerMock{},
		storer:     storer,
	}
}

// Load -
func (u *UserStorerWrapperMock) Load(key []byte) (*core.OTPInfo, error) {
	if u.LoadCalled != nil {
		return u.LoadCalled(key)
	}

	data, err := u.storer.Get(key)
	if err != nil {
		return nil, err
	}

	var otpInfo *core.OTPInfo
	err = u.marshaller.Unmarshal(&otpInfo, data)
	if err != nil {
		return nil, err
	}

	return otpInfo, nil
}

// Save -
func (u *UserStorerWrapperMock) Save(key []byte, otpInfo *core.OTPInfo) error {
	if u.SaveCalled != nil {
		return u.SaveCalled(key, otpInfo)
	}

	otpInfoBytes, err := u.marshaller.Marshal(otpInfo)
	if err != nil {
		return err
	}

	return u.storer.Put(key, otpInfoBytes)
}

// IsInterfaceNil -
func (u *UserStorerWrapperMock) IsInterfaceNil() bool {
	return u == nil
}
