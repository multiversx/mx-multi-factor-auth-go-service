package resolver

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/encryption"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/stretchr/testify/require"
)

func TestNewUserEncryptor(t *testing.T) {
	t.Parallel()

	t.Run("should return error when encryptor is nil", func(t *testing.T) {
		t.Parallel()

		ue, err := NewUserEncryptor(nil)
		require.Nil(t, ue)
		require.Equal(t, ErrNilEncryptor, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		encryptor := &testscommon.EncryptorStub{}
		ue, err := NewUserEncryptor(encryptor)
		require.NotNil(t, ue)
		require.Nil(t, err)
	})
}

func TestUserEncryptor_EncryptUserInfo(t *testing.T) {
	t.Parallel()

	testMarshaller, _ := factoryMarshaller.NewMarshalizer(factoryMarshaller.JsonMarshalizer)
	firstGuardianSk := []byte("firstGuardianSk")
	secondGuardianSk := []byte("secondGuardianSk")
	firstGuardianOTP := []byte("firstGuardianOtp")
	secondGuardianOTP := []byte("secondGuardianOtp")
	userInfo := &core.UserInfo{
		FirstGuardian: core.GuardianInfo{
			PublicKey:  []byte("firstGuardianPk"),
			PrivateKey: firstGuardianSk,
			State:      0,
			OTPData: core.OTPInfo{
				OTP:                     firstGuardianOTP,
				LastTOTPChangeTimestamp: 100,
			},
		},
		SecondGuardian: core.GuardianInfo{
			PublicKey:  []byte("secondGuardianPk"),
			PrivateKey: secondGuardianSk,
			State:      0,
			OTPData: core.OTPInfo{
				OTP:                     secondGuardianOTP,
				LastTOTPChangeTimestamp: 100,
			},
		},
	}

	t.Run("should return error when userInfo is nil", func(t *testing.T) {
		t.Parallel()

		encryptor, _ := encryption.NewEncryptor(testMarshaller, testKeygen, testSk)
		ue, _ := NewUserEncryptor(encryptor)
		encryptedUserInfo, err := ue.EncryptUserInfo(nil)
		require.Nil(t, encryptedUserInfo)
		require.Equal(t, ErrNilUserInfo, err)
	})
	t.Run("first guardian private key encryption error should return error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		encryptor := &testscommon.EncryptorStub{
			EncryptDataCalled: func(data []byte) ([]byte, error) {
				if bytes.Equal(data, firstGuardianSk) {
					return nil, expectedError
				}
				return data, nil
			},
		}
		ue, _ := NewUserEncryptor(encryptor)

		encryptedUserInfo, err := ue.EncryptUserInfo(userInfo)
		require.Nil(t, encryptedUserInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("second guardian private key encryption error should return error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		encryptor := &testscommon.EncryptorStub{
			EncryptDataCalled: func(data []byte) ([]byte, error) {
				if bytes.Equal(data, secondGuardianSk) {
					return nil, expectedError
				}
				return data, nil
			},
		}
		ue, _ := NewUserEncryptor(encryptor)

		encryptedUserInfo, err := ue.EncryptUserInfo(userInfo)
		require.Nil(t, encryptedUserInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("first guardian otp encryption error should return error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		encryptor := &testscommon.EncryptorStub{
			EncryptDataCalled: func(data []byte) ([]byte, error) {
				if bytes.Equal(data, firstGuardianOTP) {
					return nil, expectedError
				}
				return data, nil
			},
		}
		ue, _ := NewUserEncryptor(encryptor)

		encryptedUserInfo, err := ue.EncryptUserInfo(userInfo)
		require.Nil(t, encryptedUserInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("second guardian otp encryption error should return error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		encryptor := &testscommon.EncryptorStub{
			EncryptDataCalled: func(data []byte) ([]byte, error) {
				if bytes.Equal(data, secondGuardianOTP) {
					return nil, expectedError
				}
				return data, nil
			},
		}
		ue, _ := NewUserEncryptor(encryptor)

		encryptedUserInfo, err := ue.EncryptUserInfo(userInfo)
		require.Nil(t, encryptedUserInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		encryptor, _ := encryption.NewEncryptor(testMarshaller, testKeygen, testSk)
		ue, _ := NewUserEncryptor(encryptor)
		userInfo := &core.UserInfo{
			FirstGuardian: core.GuardianInfo{
				PublicKey:  []byte("firstGuardianPk"),
				PrivateKey: []byte("firstGuardianSk"),
				State:      0,
				OTPData: core.OTPInfo{
					OTP:                     []byte("firstGuardianOtp"),
					LastTOTPChangeTimestamp: 100,
				},
			},
			SecondGuardian: core.GuardianInfo{
				PublicKey:  []byte("secondGuardianPk"),
				PrivateKey: []byte("secondGuardianSk"),
				State:      0,
				OTPData: core.OTPInfo{
					OTP:                     []byte("secondGuardianOtp"),
					LastTOTPChangeTimestamp: 200,
				},
			},
			Index: 1,
		}
		encryptedUserInfo, err := ue.EncryptUserInfo(userInfo)
		require.NotNil(t, encryptedUserInfo)
		require.Nil(t, err)
		checkEncryptedUserInfo(t, userInfo, encryptedUserInfo)
	})
}

func TestUserEncryptor_DecryptUserInfo(t *testing.T) {
	t.Parallel()

	testMarshaller, _ := factoryMarshaller.NewMarshalizer(factoryMarshaller.JsonMarshalizer)
	firstGuardianSk := []byte("firstGuardianSk")
	secondGuardianSk := []byte("secondGuardianSk")
	firstGuardianOTP := []byte("firstGuardianOtp")
	secondGuardianOTP := []byte("secondGuardianOtp")
	userInfo := &core.UserInfo{
		FirstGuardian: core.GuardianInfo{
			PublicKey:  []byte("firstGuardianPk"),
			PrivateKey: firstGuardianSk,
			State:      0,
			OTPData: core.OTPInfo{
				OTP:                     firstGuardianOTP,
				LastTOTPChangeTimestamp: 100,
			},
		},
		SecondGuardian: core.GuardianInfo{
			PublicKey:  []byte("secondGuardianPk"),
			PrivateKey: secondGuardianSk,
			State:      0,
			OTPData: core.OTPInfo{
				OTP:                     secondGuardianOTP,
				LastTOTPChangeTimestamp: 100,
			},
		},
	}

	encryptor, err := encryption.NewEncryptor(testMarshaller, testKeygen, testSk)
	require.Nil(t, err)

	t.Run("should return error when userInfo is nil", func(t *testing.T) {
		t.Parallel()

		ue, _ := NewUserEncryptor(&testscommon.EncryptorStub{})
		decryptedUserInfo, err := ue.DecryptUserInfo(nil)
		require.Nil(t, decryptedUserInfo)
		require.Equal(t, ErrNilUserInfo, err)
	})
	t.Run("should return error when first guardian private key decryption error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		encryptor := &testscommon.EncryptorStub{
			DecryptDataCalled: func(data []byte) ([]byte, error) {
				if bytes.Equal(data, firstGuardianSk) {
					return nil, expectedError
				}
				return data, nil
			},
		}
		ue, _ := NewUserEncryptor(encryptor)
		decryptedUserInfo, err := ue.DecryptUserInfo(userInfo)
		require.Nil(t, decryptedUserInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("should return error when second guardian private key decryption error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		encryptor := &testscommon.EncryptorStub{
			DecryptDataCalled: func(data []byte) ([]byte, error) {
				if bytes.Equal(data, secondGuardianSk) {
					return nil, expectedError
				}
				return data, nil
			},
		}
		ue, _ := NewUserEncryptor(encryptor)
		decryptedUserInfo, err := ue.DecryptUserInfo(userInfo)
		require.Nil(t, decryptedUserInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("should return error when first guardian otp decryption error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		encryptor := &testscommon.EncryptorStub{
			DecryptDataCalled: func(data []byte) ([]byte, error) {
				if bytes.Equal(data, firstGuardianOTP) {
					return nil, expectedError
				}
				return data, nil
			},
		}
		ue, _ := NewUserEncryptor(encryptor)
		decryptedUserInfo, err := ue.DecryptUserInfo(userInfo)
		require.Nil(t, decryptedUserInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("should return error when second guardian otp decryption error", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		encryptor := &testscommon.EncryptorStub{
			DecryptDataCalled: func(data []byte) ([]byte, error) {
				if bytes.Equal(data, secondGuardianOTP) {
					return nil, expectedError
				}
				return data, nil
			},
		}
		ue, _ := NewUserEncryptor(encryptor)
		decryptedUserInfo, err := ue.DecryptUserInfo(userInfo)
		require.Nil(t, decryptedUserInfo)
		require.Equal(t, expectedError, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ue, _ := NewUserEncryptor(encryptor)
		userInfo := &core.UserInfo{
			FirstGuardian: core.GuardianInfo{
				PublicKey:  []byte("firstGuardianPk"),
				PrivateKey: []byte("firstGuardianSk"),
				State:      0,
				OTPData:    core.OTPInfo{},
			},
			SecondGuardian: core.GuardianInfo{
				PublicKey:  []byte("secondGuardianPk"),
				PrivateKey: []byte("secondGuardianSk"),
				State:      0,
				OTPData:    core.OTPInfo{},
			},
			Index: 1,
		}
		encryptedUserInfo, err := ue.EncryptUserInfo(userInfo)
		require.Nil(t, err)
		require.NotNil(t, encryptedUserInfo)

		decryptedUserInfo, err := ue.DecryptUserInfo(encryptedUserInfo)
		require.NotNil(t, decryptedUserInfo)
		require.Nil(t, err)
		require.Equal(t, userInfo, decryptedUserInfo)
	})
}

func TestUserEncryptor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ue *userEncryptor
	require.True(t, ue.IsInterfaceNil())

	ue = &userEncryptor{}
	require.False(t, ue.IsInterfaceNil())
}

func checkEncryptedUserInfo(t *testing.T, userInfo *core.UserInfo, encryptedUserInfo *core.UserInfo) {
	require.NotEqual(t, userInfo.FirstGuardian.PrivateKey, encryptedUserInfo.FirstGuardian.PrivateKey, "firstGuardianSk should be encrypted")
	require.NotEqual(t, userInfo.SecondGuardian.PrivateKey, encryptedUserInfo.SecondGuardian.PrivateKey, "secondGuardianSk should be encrypted")
	require.NotEqual(t, userInfo.FirstGuardian.OTPData.OTP, encryptedUserInfo.FirstGuardian.OTPData.OTP, "firstGuardianOtp should be encrypted")
	require.NotEqual(t, userInfo.SecondGuardian.OTPData.OTP, encryptedUserInfo.SecondGuardian.OTPData.OTP, "secondGuardianOtp should be encrypted")
	require.Equal(t, userInfo.FirstGuardian.PublicKey, encryptedUserInfo.FirstGuardian.PublicKey, "firstGuardianPk should not be encrypted")
	require.Equal(t, userInfo.SecondGuardian.PublicKey, encryptedUserInfo.SecondGuardian.PublicKey, "secondGuardianPk should not be encrypted")
	require.Equal(t, userInfo.Index, encryptedUserInfo.Index, "index should not be encrypted")
	require.Equal(t, userInfo.FirstGuardian.State, encryptedUserInfo.FirstGuardian.State, "firstGuardian state should not be encrypted")
	require.Equal(t, userInfo.SecondGuardian.State, encryptedUserInfo.SecondGuardian.State, "secondGuardian state should not be encrypted")
	require.Equal(t, userInfo.FirstGuardian.OTPData.LastTOTPChangeTimestamp, encryptedUserInfo.FirstGuardian.OTPData.GetLastTOTPChangeTimestamp(), "firstGuardian last OTP change should not be encrypted")
	require.Equal(t, userInfo.SecondGuardian.OTPData.LastTOTPChangeTimestamp, encryptedUserInfo.SecondGuardian.OTPData.GetLastTOTPChangeTimestamp(), "secondGuardian last OTP change should not be encrypted")
}
