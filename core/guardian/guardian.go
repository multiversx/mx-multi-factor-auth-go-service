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
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

const minRequestTimeInSeconds = 1

type ArgGuardian struct {
	Config          config.GuardianConfig
	Proxy           blockchain.Proxy
	PubKeyConverter core.PubkeyConverter
}

type guardian struct {
	privateKey      []byte
	address         string
	proxy           blockchain.Proxy
	builder         interactors.GuardedTxBuilder
	requestTime     time.Duration
	signer          core.TxSigVerifier
	pubKeyConverter core.PubkeyConverter
}

// NewGuardian returns a new instance of guardian
func NewGuardian(args ArgGuardian) (*guardian, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	signer := blockchain.NewTxSigner()
	builder, err := builders.NewTxBuilder(signer)
	if err != nil {
		return nil, err
	}

	g := &guardian{
		builder:         builder,
		proxy:           args.Proxy,
		requestTime:     time.Second * time.Duration(args.Config.RequestTimeInSeconds),
		signer:          signer,
		pubKeyConverter: args.PubKeyConverter,
	}
	err = g.createElrondKeysAndAddresses(args.Config)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func checkArgs(args ArgGuardian) error {
	if check.IfNil(args.Proxy) {
		return ErrNilProxy
	}
	if args.Config.RequestTimeInSeconds < minRequestTimeInSeconds {
		return fmt.Errorf("%w in checkArgs for value RequestTimeInSeconds", ErrInvalidValue)
	}
	if check.IfNil(args.PubKeyConverter) {
		return ErrNilPubkeyConverter
	}

	return nil
}

// ValidateAndSend will validate if the set guardian is its address
// it will apply his signature over transaction, and it will propagate the transaction
func (g *guardian) ValidateAndSend(transaction data.Transaction) (string, error) {
	if transaction.GuardianAddr != g.address {
		return "", errors.New("invalid guardian addr")
	}

	pkBytes, err := g.pubKeyConverter.Decode(transaction.SndAddr)
	if err != nil {
		return "", err
	}

	err = g.signer.Verify(pkBytes, transaction.Data, []byte(transaction.Signature))
	if err != nil {
		return "", err
	}

	err = g.builder.ApplyGuardianSignature(g.privateKey, &transaction)
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
