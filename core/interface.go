package core

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

// GuardedTxBuilder defines the component able to build and sign a guarded transaction
type GuardedTxBuilder interface {
	ApplyGuardianSignature(cryptoHolderGuardian core.CryptoComponentsHolder, tx *transaction.FrontendTransaction) error
	IsInterfaceNil() bool
}

// PubkeyConverter can convert public key bytes from a human-readable form
type PubkeyConverter interface {
	Len() int
	Decode(humanReadable string) ([]byte, error)
	Encode(pkBytes []byte) string
	IsInterfaceNil() bool
}

// ServiceResolver defines the methods available for a service
type ServiceResolver interface {
	RegisterUser(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error)
	VerifyCode(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error)
	SignMessage(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error)
	SignTransaction(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error)
	SignMultipleTransactions(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error)
	RegisteredUsers() (uint32, error)
	TcsConfig() *TcsConfig
	IsInterfaceNil() bool
}

// KeysGenerator defines the methods for a component able to generate unique HD keys
type KeysGenerator interface {
	GenerateManagedKey() (crypto.PrivateKey, error)
	GenerateKeys(index uint32) ([]crypto.PrivateKey, error)
	IsInterfaceNil() bool
}

// Storer provides storage services in a two layered storage construct, where the first layer is
// represented by a cache and second layer by a persistent storage (DB-like)
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

// IndexHandler defines the methods for a component which handles indexes
type IndexHandler interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	Close() error
	AllocateBucketIndex() (uint32, error)
	GetLastIndex() (uint32, error)
	IsInterfaceNil() bool
}

// StorageWithIndex defines the methods for a component that holds multiple BucketIndexHandler
type StorageWithIndex interface {
	AllocateIndex(address []byte) (uint32, error)
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	Close() error
	Count() (uint32, error)
	IsInterfaceNil() bool
}

// HttpClient defines the behavior of http client able to make http requests
type HttpClient interface {
	GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error)
	IsInterfaceNil() bool
}

// HttpClientWrapper defines the behavior of wrapper over HttpClient
type HttpClientWrapper interface {
	GetAccount(ctx context.Context, address string) (*data.Account, error)
	GetGuardianData(ctx context.Context, address string) (*api.GuardianData, error)
	IsInterfaceNil() bool
}

// NativeAuthWhitelistHandler defines the behavior of a component that handles whitelisted routes
type NativeAuthWhitelistHandler interface {
	IsWhitelisted(route string) bool
	IsInterfaceNil() bool
}

// StatusMetricsHandler defines the behavior of a component that handles status metrics
type StatusMetricsHandler interface {
	AddRequestData(path string, duration time.Duration, status int)
	GetAll() map[string]*requests.EndpointMetricsResponse
	GetMetricsForPrometheus() string
	IsInterfaceNil() bool
}
