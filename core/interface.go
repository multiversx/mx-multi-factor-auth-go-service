package core

import (
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// TxSigVerifier defines the methods available for a transaction signature verifier component
type TxSigVerifier interface {
	Verify(pk []byte, msg []byte, sigBytes []byte) error
	IsInterfaceNil() bool
}

// GuardedTxBuilder defines the component able to build and sign a guarded transaction
type GuardedTxBuilder interface {
	ApplyGuardianSignature(skGuardianBytes []byte, tx *data.Transaction) error
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
	RegisterUser(request requests.RegistrationPayload) ([]byte, string, error)
	VerifyCode(request requests.VerificationPayload) error
	SignTransaction(request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactions(request requests.SignMultipleTransactions) ([][]byte, error)
	IsInterfaceNil() bool
}

// CredentialsHandler defines the methods available for a credentials handler
type CredentialsHandler interface {
	Verify(credentials string) error
	GetAccountAddress(credentials string) (core.AddressHandler, error)
	IsInterfaceNil() bool
}

// IndexHandler defines the methods for a component able to provide unique indexes
type IndexHandler interface {
	AllocateIndex(address []byte) (uint32, error)
	IsInterfaceNil() bool
}

// KeysGenerator defines the methods for a component able to generate unique HD keys
type KeysGenerator interface {
	GenerateKeys(index uint32) ([]crypto.PrivateKey, error)
	IsInterfaceNil() bool
}

// Storer provides storage services in a two layered storage construct, where the first layer is
// represented by a cache and second layer by a persitent storage (DB-like)
type Storer interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	SearchFirst(key []byte) ([]byte, error)
	Remove(key []byte) error
	ClearCache()
	Close() error
	IsInterfaceNil() bool
}

// Marshaller defines the 2 basic operations: serialize (marshal) and deserialize (unmarshal)
type Marshaller interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(obj interface{}, buff []byte) error
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

// BucketIDProvider defines the methods for a component able to extract a bucket id from an address
type BucketIDProvider interface {
	GetBucketForAddress(address []byte) uint32
	IsInterfaceNil() bool
}

// BucketIndexHandler defines the methods for a component which handles a bucket
type BucketIndexHandler interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	Close() error
	UpdateIndexReturningNext() (uint32, error)
	IsInterfaceNil() bool
}

// BucketIndexHolder defines the methods for a component that holds multiple BucketIndexHandler
type BucketIndexHolder interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	Close() error
	UpdateIndexReturningNext(address []byte) (uint32, error)
	NumberOfBuckets() uint32
	IsInterfaceNil() bool
}
