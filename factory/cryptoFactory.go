package factory

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
)

const bech32Format = "bech32"

// cryptoComponentsHolder will hold core crypto components
type cryptoComponentsHolder struct {
	keyGenerator    crypto.KeyGenerator
	signer          builders.Signer
	pubKeyConverter core.PubkeyConverter
}

// CreateCoreCryptoComponents will create core crypto components
func CreateCoreCryptoComponents(conf config.PubkeyConfig) (*cryptoComponentsHolder, error) {
	pkConv, err := createPubkeyConverter(conf)
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

func createPubkeyConverter(config config.PubkeyConfig) (core.PubkeyConverter, error) {
	switch config.Type {
	case bech32Format:
		return pubkeyConverter.NewBech32PubkeyConverter(config.Length, config.Hrp)
	default:
		return nil, fmt.Errorf("%w unrecognized type %s", core.ErrInvalidPubkeyConverterType, config.Type)
	}
}

// KeyGenerator returns key generator component
func (cch *cryptoComponentsHolder) KeyGenerator() crypto.KeyGenerator {
	return cch.keyGenerator
}

// Signer returns signer component
func (cch *cryptoComponentsHolder) Signer() builders.Signer {
	return cch.signer
}

// PubkeyConverter returns pubkey converter component
func (cch *cryptoComponentsHolder) PubkeyConverter() core.PubkeyConverter {
	return cch.pubKeyConverter
}
