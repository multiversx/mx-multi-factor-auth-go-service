package guardian

import (
	"context"
	"encoding/hex"
	"encoding/json"
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
	usersHandler    core.UsersHandler
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
		usersHandler:    core.NewUsersHandler(),
	}
	err = g.createElrondKeysAndAddresses(args.Config)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func checkArgs(args ArgGuardian) error {
	if check.IfNil(args.Proxy) {
		return core.ErrNilProxy
	}
	if args.Config.RequestTimeInSeconds < minRequestTimeInSeconds {
		return fmt.Errorf("%w for RequestTimeInSeconds, received %d, min expected %d",
			core.ErrInvalidValue, args.Config.RequestTimeInSeconds, minRequestTimeInSeconds)
	}
	if check.IfNil(args.PubKeyConverter) {
		return core.ErrNilPubkeyConverter
	}

	return nil
}

// ValidateAndSend will validate if the set guardian is its address
// it will apply his signature over transaction, and it will propagate the transaction
func (g *guardian) ValidateAndSend(transaction data.Transaction) (string, error) {
	err := g.validateTransaction(transaction)
	if err != nil {
		return "", err
	}

	err = g.builder.ApplyGuardianSignature(g.privateKey, &transaction)
	if err != nil {
		return "", err
	}

	return g.sendTransaction(transaction)
}

func (g *guardian) validateTransaction(transaction data.Transaction) error {
	if transaction.GuardianAddr != g.address {
		return core.ErrInvalidGuardianAddress
	}

	if !g.usersHandler.HasUser(transaction.SndAddr) {
		return core.ErrInvalidSenderAddress
	}

	err := g.verifyActiveGuardian(transaction.SndAddr)
	if err != nil {
		return err
	}

	err = g.verifySignature(transaction)
	if err != nil {
		return err
	}

	return nil
}

func (g *guardian) verifyActiveGuardian(userAddress string) error {
	userAddr, err := data.NewAddressFromBech32String(userAddress)
	if err != nil {
		return err
	}

	guardianDataCtx, cancel := context.WithTimeout(context.Background(), g.requestTime)
	defer cancel()
	guardianData, err := g.proxy.GetGuardianData(guardianDataCtx, userAddr)
	if err != nil {
		return err
	}

	if guardianData.ActiveGuardian.Address != g.address {
		return core.ErrInactiveGuardian
	}

	return nil
}

func (g *guardian) verifySignature(transaction data.Transaction) error {
	pkBytes, err := g.pubKeyConverter.Decode(transaction.SndAddr)
	if err != nil {
		return err
	}

	signature := transaction.Signature
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}

	transaction.Signature = ""
	transaction.GuardianSignature = ""
	buff, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	err = g.signer.Verify(pkBytes, buff, signatureBytes)
	if err != nil {
		return err
	}

	return nil
}

func (g *guardian) sendTransaction(transaction data.Transaction) (string, error) {
	sendTxCtx, cancel := context.WithTimeout(context.Background(), g.requestTime)
	defer cancel()

	return g.proxy.SendTransaction(sendTxCtx, &transaction)
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

// AddUser adds the provided address into the internal registered users' map
func (g *guardian) AddUser(address string) {
	g.usersHandler.AddUser(address)
}

// HasUser returns true if the provided address is registered for this guardian
func (g *guardian) HasUser(address string) bool {
	return g.usersHandler.HasUser(address)
}

// RemoveUser removes the provided address from the internal registered users' map
func (g *guardian) RemoveUser(address string) {
	g.usersHandler.RemoveUser(address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *guardian) IsInterfaceNil() bool {
	return g == nil
}
