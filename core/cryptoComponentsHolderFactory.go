package core

import (
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain/cryptoProvider"
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
)

var keyGen = crypto.NewKeyGenerator(ed25519.NewEd25519())

type CryptoComponentsHolderFactory struct{}

// Create creates a new instance of CryptoComponentsHolder
func (f *CryptoComponentsHolderFactory) Create(skBytes []byte) (erdCore.CryptoComponentsHolder, error) {
	return cryptoProvider.NewCryptoComponentsHolder(keyGen, skBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *CryptoComponentsHolderFactory) IsInterfaceNil() bool {
	return f == nil
}
