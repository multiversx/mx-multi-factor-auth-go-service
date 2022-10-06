package core

import (
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
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
	Encode(pkBytes []byte) string
	IsInterfaceNil() bool
}

// UsersHandler defines the methods available for a users handler
type UsersHandler interface {
	AddUser(address string)
	HasUser(address string) bool
	RemoveUser(address string)
	IsInterfaceNil() bool
}

// ServiceResolver defines the methods available for a service
type ServiceResolver interface {
	GetGuardianAddress(request requests.GetGuardianAddress) (string, error)
}

// CredentialsHandler defines the methods available for a credentials handler
type CredentialsHandler interface {
	Verify(credentials string) error
	GetAccountAddress(credentials string) (core.AddressHandler, error)
}

// IndexHandler defines the methods for a component able to provide unique indexes
type IndexHandler interface {
	GetIndex() uint64
}

// KeysGenerator defines the methods for a component able to generate unique HD keys
type KeysGenerator interface {
	GenerateKeys(index uint64) []crypto.PrivateKey
}

// Storer provides storage services for a persistent storage(DB-like)
type Storer interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) bool
	Remove(key []byte) error
	Close() error
	IsInterfaceNil() bool
}

// Marshaller defines the 2 basic operations: serialize (marshal) and deserialize (unmarshal)
type Marshaller interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(obj interface{}, buff []byte) error
	IsInterfaceNil() bool
}
