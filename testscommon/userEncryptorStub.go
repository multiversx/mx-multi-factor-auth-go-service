package testscommon

import "github.com/multiversx/mx-multi-factor-auth-go-service/core"

// UserEncryptorStub is a stub implementation of UserEncryptor
type UserEncryptorStub struct {
	EncryptUserInfoCalled func(userInfo *core.UserInfo) (*core.UserInfo, error)
	DecryptUserInfoCalled func(userInfo *core.UserInfo) (*core.UserInfo, error)
}

// EncryptUserInfo encrypts the provided user info
func (ues *UserEncryptorStub) EncryptUserInfo(userInfo *core.UserInfo) (*core.UserInfo, error) {
	if ues.EncryptUserInfoCalled != nil {
		return ues.EncryptUserInfoCalled(userInfo)
	}
	return userInfo, nil
}

// DecryptUserInfo decrypts the provided user info
func (ues *UserEncryptorStub) DecryptUserInfo(userInfo *core.UserInfo) (*core.UserInfo, error) {
	if ues.DecryptUserInfoCalled != nil {
		return ues.DecryptUserInfoCalled(userInfo)
	}
	return userInfo, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ues *UserEncryptorStub) IsInterfaceNil() bool {
	return ues == nil
}
