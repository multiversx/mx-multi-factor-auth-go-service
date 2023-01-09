package core

import (
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain/cryptoProvider"
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
)

// CryptoComponentsHolderFactory is the implementation of the CryptoComponentsHolderFactory interface
type CryptoComponentsHolderFactory struct {
	keyGen crypto.KeyGenerator
}

// NewCryptoComponentsHolderFactory creates a new instance of CryptoComponentsHolderFactory
func NewCryptoComponentsHolderFactory(keyGen crypto.KeyGenerator) *CryptoComponentsHolderFactory {
	return &CryptoComponentsHolderFactory{
		keyGen: keyGen,
	}
}

// Create creates a new instance of CryptoComponentsHolder
func (f *CryptoComponentsHolderFactory) Create(skBytes []byte) (erdCore.CryptoComponentsHolder, error) {
	return cryptoProvider.NewCryptoComponentsHolder(f.keyGen, skBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *CryptoComponentsHolderFactory) IsInterfaceNil() bool {
	return f == nil
}
