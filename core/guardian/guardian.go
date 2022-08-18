package guardian

import (
	"context"
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
)

type guardian struct {
	privateKey  []byte
	address     string
	proxy       blockchain.Proxy
	builder     interactors.GuardedTxBuilder
	requestTime time.Duration
}

func NewGuardian(config config.GuardianConfig, proxy blockchain.Proxy) (*guardian, error) {
	signer := blockchain.NewTxSigner()
	builder, err := builders.NewTxBuilder(signer)
	if err != nil {
		return nil, err
	}

	g := &guardian{
		builder:     builder,
		proxy:       proxy,
		requestTime: time.Second * time.Duration(config.RequestTimeInSeconds),
	}
	err = g.createElrondKeysAndAddresses(config)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *guardian) ValidateAndSend(transaction data.Transaction) (string, error) {
	if transaction.GuardianAddr != g.address {
		return "", errors.New("invalid guardian addr")
	}
	err := g.builder.ApplyGuardianSignature(g.privateKey, &transaction)
	if err != nil {
		return "", err
	}
	requestContext, cancel := context.WithTimeout(context.Background(), g.requestTime)
	defer cancel()
	hash, err := g.proxy.SendTransaction(requestContext, &transaction)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func (g *guardian) createElrondKeysAndAddresses(config config.GuardianConfig) error {
	wallet := interactors.NewWallet()
	var err error
	g.privateKey, err = wallet.LoadPrivateKeyFromPemFile(config.PrivateKeyFile)
	if err != nil {
		return err
	}

	address, err := wallet.GetAddressFromPrivateKey(g.privateKey)
	if err != nil {
		return err
	}
	g.address = address.AddressAsBech32String()

	return nil
}
