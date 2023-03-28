package resolver

import (
	"github.com/multiversx/multi-factor-auth-go-service/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/encryption/x25519"
)

func (resolver *serviceResolver) encryptAndMarshalUserInfo(userInfo *core.UserInfo) ([]byte, error) {
	encryptedUserInfo, err := resolver.encryptUserInfo(userInfo)
	if err != nil {
		return nil, err
	}

	return resolver.userDataMarshaller.Marshal(encryptedUserInfo)
}

func (resolver *serviceResolver) unmarshalAndDecryptUserInfo(encryptedDataMarshalled []byte) (*core.UserInfo, error) {
	userInfo := &core.UserInfo{}
	err := resolver.userDataMarshaller.Unmarshal(userInfo, encryptedDataMarshalled)
	if err != nil {
		return nil, err
	}

	return resolver.decryptUserInfo(userInfo)
}

func (resolver *serviceResolver) encryptUserInfo(userInfo *core.UserInfo) (*core.UserInfo, error) {
	if userInfo == nil {
		return nil, ErrNilUserInfo
	}

	encryptionSk, _ := resolver.keyGen.GeneratePair()
	firstGuardianSk, err := resolver.encryptData(userInfo.FirstGuardian.PrivateKey, encryptionSk)
	if err != nil {
		return nil, err
	}

	secondGuardianSk, err := resolver.encryptData(userInfo.SecondGuardian.PrivateKey, encryptionSk)
	if err != nil {
		return nil, err
	}

	encryptedUserInfo := *userInfo
	encryptedUserInfo.FirstGuardian.PrivateKey = firstGuardianSk
	encryptedUserInfo.SecondGuardian.PrivateKey = secondGuardianSk

	return &encryptedUserInfo, nil
}

func (resolver *serviceResolver) encryptData(data []byte, encryptionSk crypto.PrivateKey) ([]byte, error) {
	encryptedData := &x25519.EncryptedData{}
	err := encryptedData.Encrypt(data, resolver.managedPrivateKey.GeneratePublic(), encryptionSk)
	if err != nil {
		return nil, err
	}

	encryptedDataBytes, err := resolver.encryptionMarshaller.Marshal(encryptedData)
	if err != nil {
		return nil, err
	}

	return encryptedDataBytes, nil
}

func (resolver *serviceResolver) decryptUserInfo(userInfo *core.UserInfo) (*core.UserInfo, error) {
	if userInfo == nil {
		return nil, ErrNilUserInfo
	}

	decryptedFirstGuardianSk, err := resolver.decryptData(userInfo.FirstGuardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	decryptedSecondGuardianSk, err := resolver.decryptData(userInfo.SecondGuardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	decryptedUserInfo := *userInfo
	decryptedUserInfo.FirstGuardian.PrivateKey = decryptedFirstGuardianSk
	decryptedUserInfo.SecondGuardian.PrivateKey = decryptedSecondGuardianSk

	return &decryptedUserInfo, nil
}

func (resolver *serviceResolver) decryptData(data []byte) ([]byte, error) {
	encryptedData := &x25519.EncryptedData{}
	err := resolver.encryptionMarshaller.Unmarshal(encryptedData, data)
	if err != nil {
		return nil, err
	}

	decryptedData, err := encryptedData.Decrypt(resolver.managedPrivateKey)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
