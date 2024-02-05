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

// CryptoComponentsHolder will hold core crypto components
type CryptoComponentsHolder struct {
	KeyGenerator    crypto.KeyGenerator
	Signer          builders.Signer
	PubKeyConverter core.PubkeyConverter
}

// CreateCoreCryptoComponents will create core crypto components
func CreateCoreCryptoComponents() (*CryptoComponentsHolder, error) {
	pkConv, err := pubkeyConverter.NewBech32PubkeyConverter(userAddressLength, log)
	if err != nil {
		return nil, err
	}

	keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
	signer := cryptoProvider.NewSigner()

	return &CryptoComponentsHolder{
		KeyGenerator:    keyGen,
		Signer:          signer,
		PubKeyConverter: pkConv,
	}, nil
}
