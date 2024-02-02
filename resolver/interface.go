package resolver

import (
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

// CryptoComponentsHolderFactory is the interface that defines the methods that
// can be used to create a new instance of CryptoComponentsHolder
type CryptoComponentsHolderFactory interface {
	Create(privateKeyBytes []byte) (sdkCore.CryptoComponentsHolder, error)
	IsInterfaceNil() bool
}

// Encryptor is the interface that defines the methods that can be used to encrypt and decrypt data
type Encryptor interface {
	EncryptData(data []byte) ([]byte, error)
	DecryptData(data []byte) ([]byte, error)
	IsInterfaceNil() bool
}

// UserEncryptor is the interface that defines the methods that can be used to encrypt and decrypt user info
type UserEncryptor interface {
	EncryptUserInfo(userInfo *core.UserInfo) (*core.UserInfo, error)
	DecryptUserInfo(userInfo *core.UserInfo) (*core.UserInfo, error)
	IsInterfaceNil() bool
}
