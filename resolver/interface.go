package resolver

import "github.com/multiversx/mx-sdk-go/core"

// CryptoComponentsHolderFactory is the interface that defines the methods that
// can be used to create a new instance of CryptoComponentsHolder
type CryptoComponentsHolderFactory interface {
	Create(privateKeyBytes []byte) (core.CryptoComponentsHolder, error)
	IsInterfaceNil() bool
}
