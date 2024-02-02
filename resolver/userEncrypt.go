package resolver

import (
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type userEncryptor struct {
	encryptor Encryptor
}

// NewUserEncryptor creates a new instance of userEncryptor
func NewUserEncryptor(encryptor Encryptor) (*userEncryptor, error) {
	if check.IfNil(encryptor) {
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

	otpFirstGuardian, err := ue.encryptor.EncryptData(userInfo.FirstGuardian.OTPData.OTP)
	if err != nil {
		return nil, err
	}

	secondGuardianSk, err := ue.encryptor.EncryptData(userInfo.SecondGuardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	otpSecondGuardian, err := ue.encryptor.EncryptData(userInfo.SecondGuardian.OTPData.OTP)
	if err != nil {
		return nil, err
	}

	encryptedUserInfo := *userInfo
	encryptedUserInfo.FirstGuardian.PrivateKey = firstGuardianSk
	encryptedUserInfo.FirstGuardian.OTPData.OTP = otpFirstGuardian
	encryptedUserInfo.SecondGuardian.PrivateKey = secondGuardianSk
	encryptedUserInfo.SecondGuardian.OTPData.OTP = otpSecondGuardian

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

	decryptedFirstGuardianOTP, err := ue.encryptor.DecryptData(userInfo.FirstGuardian.OTPData.OTP)
	if err != nil {
		return nil, err
	}

	decryptedSecondGuardianSk, err := ue.encryptor.DecryptData(userInfo.SecondGuardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	decryptedSecondGuardianSkOTP, err := ue.encryptor.DecryptData(userInfo.SecondGuardian.OTPData.OTP)
	if err != nil {
		return nil, err
	}

	decryptedUserInfo := *userInfo
	decryptedUserInfo.FirstGuardian.PrivateKey = decryptedFirstGuardianSk
	decryptedUserInfo.FirstGuardian.OTPData.OTP = decryptedFirstGuardianOTP
	decryptedUserInfo.SecondGuardian.PrivateKey = decryptedSecondGuardianSk
	decryptedUserInfo.SecondGuardian.OTPData.OTP = decryptedSecondGuardianSkOTP

	return &decryptedUserInfo, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ue *userEncryptor) IsInterfaceNil() bool {
	return ue == nil
}
