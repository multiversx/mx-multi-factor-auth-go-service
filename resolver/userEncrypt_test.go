package resolver

import (
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

	t.Run("should return error when userInfo is nil", func(t *testing.T) {
		t.Parallel()

		ue, _ := NewUserEncryptor(&testscommon.EncryptorStub{})
		encryptedUserInfo, err := ue.EncryptUserInfo(nil)
		require.Nil(t, encryptedUserInfo)
		require.Equal(t, ErrNilUserInfo, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		encryptor := &testscommon.EncryptorStub{}
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
		require.NotNil(t, encryptedUserInfo)
		require.Nil(t, err)
		checkEncryptedUserInfo(t, userInfo, encryptedUserInfo)
	})
}

func TestUserEncryptor_DecryptUserInfo(t *testing.T) {
	t.Parallel()

	testMarshaller, _ := factoryMarshaller.NewMarshalizer(factoryMarshaller.JsonMarshalizer)
	encryptor, err := encryption.NewEncryptor(testMarshaller, testKeygen, testSk)
	require.Nil(t, err)

	t.Run("should return error when userInfo is nil", func(t *testing.T) {
		t.Parallel()

		ue, _ := NewUserEncryptor(&testscommon.EncryptorStub{})
		decryptedUserInfo, err := ue.DecryptUserInfo(nil)
		require.Nil(t, decryptedUserInfo)
		require.Equal(t, ErrNilUserInfo, err)
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

func checkEncryptedUserInfo(t *testing.T, userInfo *core.UserInfo, encryptedUserInfo *core.UserInfo) {
	require.NotEqual(t, userInfo.FirstGuardian.PrivateKey, encryptedUserInfo.FirstGuardian.PrivateKey, "firstGuardianSk should be encrypted")
	require.NotEqual(t, userInfo.SecondGuardian.PrivateKey, encryptedUserInfo.SecondGuardian.PrivateKey, "secondGuardianSk should be encrypted")
	require.Equal(t, userInfo.FirstGuardian.PublicKey, encryptedUserInfo.FirstGuardian.PublicKey, "firstGuardianPk should not be encrypted")
	require.Equal(t, userInfo.SecondGuardian.PublicKey, encryptedUserInfo.SecondGuardian.PublicKey, "secondGuardianPk should not be encrypted")
	require.Equal(t, userInfo.Index, encryptedUserInfo.Index, "index should not be encrypted")
	require.Equal(t, userInfo.FirstGuardian.State, encryptedUserInfo.FirstGuardian.State, "firstGuardian state should not be encrypted")
	require.Equal(t, userInfo.SecondGuardian.State, encryptedUserInfo.SecondGuardian.State, "secondGuardian state should not be encrypted")
	require.Equal(t, userInfo.FirstGuardian.OTPData, encryptedUserInfo.FirstGuardian.OTPData, "firstGuardian OTPData should not be encrypted")
}
