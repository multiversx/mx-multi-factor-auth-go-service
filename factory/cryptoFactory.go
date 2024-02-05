package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
)

const (
	userAddressLength = 32
)

var log = logger.GetOrCreate("factory")

// cryptoComponentsHolder will hold core crypto components
type cryptoComponentsHolder struct {
	keyGenerator    crypto.KeyGenerator
	signer          builders.Signer
	pubKeyConverter core.PubkeyConverter
}

// CreateCoreCryptoComponents will create core crypto components
func CreateCoreCryptoComponents() (*cryptoComponentsHolder, error) {
	pkConv, err := pubkeyConverter.NewBech32PubkeyConverter(userAddressLength, log)
	if err != nil {
		return nil, err
	}

	keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
	signer := cryptoProvider.NewSigner()

	return &cryptoComponentsHolder{
		keyGenerator:    keyGen,
		signer:          signer,
		pubKeyConverter: pkConv,
	}, nil
}

// KeyGenerator returns key generator component
func (cch *cryptoComponentsHolder) KeyGenerator() crypto.KeyGenerator {
	return cch.keyGenerator
}

// Signer returns signer component
func (cch *cryptoComponentsHolder) Singer() builders.Signer {
	return cch.signer
}

// PubkeyConverter returns signer component
func (cch *cryptoComponentsHolder) PubkeyConverter() core.PubkeyConverter {
	return cch.pubKeyConverter
}
