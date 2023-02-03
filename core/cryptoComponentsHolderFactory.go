package core

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/core"
)

// CryptoComponentsHolderFactory is the implementation of the CryptoComponentsHolderFactory interface
type CryptoComponentsHolderFactory struct {
	keyGen crypto.KeyGenerator
}

// NewCryptoComponentsHolderFactory creates a new instance of CryptoComponentsHolderFactory
func NewCryptoComponentsHolderFactory(keyGen crypto.KeyGenerator) (*CryptoComponentsHolderFactory, error) {
	if check.IfNil(keyGen) {
		return nil, ErrNilKeyGenerator
	}

	return &CryptoComponentsHolderFactory{
		keyGen: keyGen,
	}, nil
}

// Create creates a new instance of CryptoComponentsHolder
func (f *CryptoComponentsHolderFactory) Create(skBytes []byte) (core.CryptoComponentsHolder, error) {
	return cryptoProvider.NewCryptoComponentsHolder(f.keyGen, skBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *CryptoComponentsHolderFactory) IsInterfaceNil() bool {
	return f == nil
}
