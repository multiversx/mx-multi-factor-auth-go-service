package guardian

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
)

const minRequestTimeInSeconds = 1

type guardian struct {
	privateKey  []byte
	address     string
	proxy       blockchain.Proxy
	builder     interactors.GuardedTxBuilder
	requestTime time.Duration
}

// NewGuardian returns a new instance of guardian
func NewGuardian(config config.GuardianConfig, proxy blockchain.Proxy) (*guardian, error) {
	err := checkArgs(config, proxy)
	if err != nil {
		return nil, err
	}

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

func checkArgs(config config.GuardianConfig, proxy blockchain.Proxy) error {
	if check.IfNil(proxy) {
		return ErrNilProxy
	}
	if config.RequestTimeInSeconds < minRequestTimeInSeconds {
		return fmt.Errorf("%w in checkArgs for value RequestTimeInSeconds", ErrInvalidValue)
	}
	return nil
}

// ValidateAndSend will validate if the set guardian is its address
// it will apply his signature over transaction, and it will propagate the transaction
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

	return g.proxy.SendTransaction(requestContext, &transaction)
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

// GetAddress returns the address of the guardian
func (g *guardian) GetAddress() string {
	return g.address
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *guardian) IsInterfaceNil() bool {
	return g == nil
}
