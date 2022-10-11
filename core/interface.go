package core

import (
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// Guardian defines the methods available for a guardian component
type Guardian interface {
	UsersHandler
	ValidateAndSend(transaction data.Transaction) (string, error)
	GetAddress() string
	IsInterfaceNil() bool
}

// Provider defines the actions needed to be performed by a multi-auth provider
type Provider interface {
	LoadSavedAccounts() error
	Validate(account, userCode string) (bool, error)
	RegisterUser(account string) ([]byte, error)
	IsInterfaceNil() bool
}

// TxSigVerifier defines the methods available for a transaction signature verifier component
type TxSigVerifier interface {
	Verify(pk []byte, msg []byte, skBytes []byte) error
	IsInterfaceNil() bool
}

// PubkeyConverter can convert public key bytes from a human-readable form
type PubkeyConverter interface {
	Len() int
	Decode(humanReadable string) ([]byte, error)
	IsInterfaceNil() bool
}

// UsersHandler defines the methods available for a users handler
type UsersHandler interface {
	AddUser(address string)
	HasUser(address string) bool
	RemoveUser(address string)
	IsInterfaceNil() bool
}

// GuardianKeyGenerator defines the methods for a component able to generate unique HD keys for a guardian
type GuardianKeyGenerator interface {
	GenerateKeys(index uint32) ([]crypto.PrivateKey, error)
	IsInterfaceNil() bool
}

// KeyGenerator defines the methods for a component able to create a crypto.PrivateKey from a byte array
type KeyGenerator interface {
	PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error)
	IsInterfaceNil() bool
}
