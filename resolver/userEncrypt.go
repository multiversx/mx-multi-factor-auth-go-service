package resolver

import (
	"github.com/multiversx/multi-factor-auth-go-service/core"
)

type userEncryptor struct {
	encryptor Encryptor
}

// NewUserEncryptor creates a new instance of userEncryptor
func NewUserEncryptor(encryptor Encryptor) (*userEncryptor, error) {
	if encryptor == nil {
		return nil, ErrNilEncryptor
	}

	return &userEncryptor{
		encryptor: encryptor,
	}, nil
}

// EncryptUserInfo encrypts the provided user info
func (ue *userEncryptor) EncryptUserInfo(userInfo *core.UserInfo) (*core.UserInfo, error) {
	if userInfo == nil {
		return nil, ErrNilUserInfo
	}

	firstGuardianSk, err := ue.encryptor.EncryptData(userInfo.FirstGuardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	secondGuardianSk, err := ue.encryptor.EncryptData(userInfo.SecondGuardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	encryptedUserInfo := *userInfo
	encryptedUserInfo.FirstGuardian.PrivateKey = firstGuardianSk
	encryptedUserInfo.SecondGuardian.PrivateKey = secondGuardianSk

	return &encryptedUserInfo, nil
}

// DecryptUserInfo decrypts the provided user info
func (ue *userEncryptor) DecryptUserInfo(userInfo *core.UserInfo) (*core.UserInfo, error) {
	if userInfo == nil {
		return nil, ErrNilUserInfo
	}

	decryptedFirstGuardianSk, err := ue.encryptor.DecryptData(userInfo.FirstGuardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	decryptedSecondGuardianSk, err := ue.encryptor.DecryptData(userInfo.SecondGuardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	decryptedUserInfo := *userInfo
	decryptedUserInfo.FirstGuardian.PrivateKey = decryptedFirstGuardianSk
	decryptedUserInfo.SecondGuardian.PrivateKey = decryptedSecondGuardianSk

	return &decryptedUserInfo, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ue *userEncryptor) IsInterfaceNil() bool {
	return ue == nil
}
