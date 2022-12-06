package resolver

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	erdData "github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/txcheck"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/providers"
)

const (
	emptyAddress   = ""
	minRequestTime = time.Second
)

// ArgServiceResolver is the DTO used to create a new instance of service resolver
type ArgServiceResolver struct {
	Provider           providers.Provider
	Proxy              blockchain.Proxy
	CredentialsHandler core.CredentialsHandler
	IndexHandler       core.IndexHandler
	KeysGenerator      core.KeysGenerator
	PubKeyConverter    core.PubkeyConverter
	RegisteredUsersDB  core.Storer
	Marshaller         core.Marshaller
	TxHasher           data.Hasher
	SignatureVerifier  core.TxSigVerifier
	GuardedTxBuilder   core.GuardedTxBuilder
	RequestTime        time.Duration
}

type serviceResolver struct {
	provider           providers.Provider
	proxy              blockchain.Proxy
	credentialsHandler core.CredentialsHandler
	indexHandler       core.IndexHandler
	keysGenerator      core.KeysGenerator
	pubKeyConverter    core.PubkeyConverter
	registeredUsersDB  core.Storer
	marshaller         core.Marshaller
	txHasher           data.Hasher
	requestTime        time.Duration
	signatureVerifier  core.TxSigVerifier
	guardedTxBuilder   core.GuardedTxBuilder
}

// NewServiceResolver returns a new instance of service resolver
func NewServiceResolver(args ArgServiceResolver) (*serviceResolver, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &serviceResolver{
		provider:           args.Provider,
		proxy:              args.Proxy,
		credentialsHandler: args.CredentialsHandler,
		indexHandler:       args.IndexHandler,
		keysGenerator:      args.KeysGenerator,
		pubKeyConverter:    args.PubKeyConverter,
		registeredUsersDB:  args.RegisteredUsersDB,
		marshaller:         args.Marshaller,
		txHasher:           args.TxHasher,
		requestTime:        args.RequestTime,
		signatureVerifier:  args.SignatureVerifier,
		guardedTxBuilder:   args.GuardedTxBuilder,
	}, nil
}

func checkArgs(args ArgServiceResolver) error {
	if check.IfNil(args.Provider) {
		return ErrNilProvider
	}
	if check.IfNil(args.Proxy) {
		return ErrNilProxy
	}
	if check.IfNil(args.CredentialsHandler) {
		return ErrNilCredentialsHandler
	}
	if check.IfNil(args.IndexHandler) {
		return ErrNilIndexHandler
	}
	if check.IfNil(args.KeysGenerator) {
		return ErrNilKeysGenerator
	}
	if check.IfNil(args.PubKeyConverter) {
		return ErrNilPubKeyConverter
	}
	if check.IfNil(args.RegisteredUsersDB) {
		return ErrNilStorer
	}
	if check.IfNil(args.Marshaller) {
		return ErrNilMarshaller
	}
	if check.IfNil(args.TxHasher) {
		return ErrNilHasher
	}
	if check.IfNil(args.SignatureVerifier) {
		return ErrNilSignatureVerifier
	}
	if check.IfNil(args.GuardedTxBuilder) {
		return ErrNilGuardedTxBuilder
	}
	if args.RequestTime < minRequestTime {
		return fmt.Errorf("%w for RequestTime, received %d, min expected %d", ErrInvalidValue, args.RequestTime, minRequestTime)
	}

	return nil
}

// getGuardianAddress returns the address of a unique guardian
func (resolver *serviceResolver) getGuardianAddress(userAddress erdCore.AddressHandler) (string, error) {
	addressBytes := userAddress.AddressBytes()
	err := resolver.registeredUsersDB.Has(addressBytes)
	if err != nil {
		return resolver.handleNewAccount(addressBytes)
	}

	return resolver.handleRegisteredAccount(addressBytes)
}

// RegisterUser creates a new OTP for the given provider
// and (optionally) returns some information required for the user to set up the OTP on his end (eg: QR code).
func (resolver *serviceResolver) RegisterUser(request requests.RegistrationPayload) ([]byte, string, error) {
	userAddress, err := resolver.validateCredentials(request.Credentials)
	if err != nil {
		return nil, "", err
	}

	guardianAddress, err := resolver.getGuardianAddress(userAddress)
	if err != nil {
		return nil, "", err
	}

	tag := resolver.extractUserTagForQRGeneration(request.Tag, userAddress.Pretty())
	qr, err := resolver.provider.RegisterUser(userAddress.AddressAsBech32String(), tag, guardianAddress)
	if err != nil {
		return nil, "", err
	}

	return qr, guardianAddress, nil
}

// VerifyCode validates the code received
func (resolver *serviceResolver) VerifyCode(request requests.VerificationPayload) error {
	userAddress, err := resolver.validateCredentials(request.Credentials)
	if err != nil {
		return err
	}

	err = resolver.provider.ValidateCode(userAddress.AddressAsBech32String(), request.Guardian, request.Code)
	if err != nil {
		return err
	}

	guardianAddrBuff, err := resolver.pubKeyConverter.Decode(request.Guardian)
	if err != nil {
		return err
	}

	return resolver.updateGuardianStateIfNeeded(userAddress.AddressBytes(), guardianAddrBuff)
}

// SignTransaction validates user's transaction, then adds guardian signature and returns the transaction
func (resolver *serviceResolver) SignTransaction(request requests.SignTransaction) ([]byte, error) {
	guardian, err := resolver.validateTxRequestReturningGuardian(request.Credentials, request.Code, []erdData.Transaction{request.Tx})
	if err != nil {
		return nil, err
	}

	err = resolver.guardedTxBuilder.ApplyGuardianSignature(guardian.PrivateKey, &request.Tx)
	if err != nil {
		return nil, err
	}

	return resolver.marshaller.Marshal(&request.Tx)
}

// SignMultipleTransactions validates user's transactions, then adds guardian signature and returns the transaction
func (resolver *serviceResolver) SignMultipleTransactions(request requests.SignMultipleTransactions) ([][]byte, error) {
	guardian, err := resolver.validateTxRequestReturningGuardian(request.Credentials, request.Code, request.Txs)
	if err != nil {
		return nil, err
	}

	txsSlice := make([][]byte, 0)
	for _, tx := range request.Txs {
		err = resolver.guardedTxBuilder.ApplyGuardianSignature(guardian.PrivateKey, &tx)
		if err != nil {
			return nil, err
		}

		txBuff, err := resolver.marshaller.Marshal(&tx)
		if err != nil {
			return nil, err
		}

		txsSlice = append(txsSlice, txBuff)
	}

	return txsSlice, nil
}

func (resolver *serviceResolver) validateTxRequestReturningGuardian(credentials string, code string, txs []erdData.Transaction) (core.GuardianInfo, error) {
	userAddress, err := resolver.validateCredentials(credentials)
	if err != nil {
		return core.GuardianInfo{}, err
	}

	err = resolver.validateTransactions(txs, userAddress)
	if err != nil {
		return core.GuardianInfo{}, err
	}

	// only validate the guardian for first tx, as all of them must have the same one
	err = resolver.provider.ValidateCode(userAddress.AddressAsBech32String(), txs[0].GuardianAddr, code)
	if err != nil {
		return core.GuardianInfo{}, err
	}

	userInfo, err := resolver.getUserInfo(userAddress.AddressBytes())
	if err != nil {
		return core.GuardianInfo{}, err
	}

	// only get the guardian for first tx, as all of them must have the same one
	return resolver.getGuardianForTx(txs[0], userInfo)
}

func (resolver *serviceResolver) updateGuardianStateIfNeeded(userAddress []byte, guardianAddress []byte) error {
	userInfo, err := resolver.getUserInfo(userAddress)
	if err != nil {
		return err
	}

	if bytes.Equal(guardianAddress, userInfo.FirstGuardian.PublicKey) {
		if userInfo.FirstGuardian.State != core.NotUsableYet {
			return fmt.Errorf("%w for FirstGuardian, it is not in NotUsableYet state", ErrInvalidGuardianState)
		}
		userInfo.FirstGuardian.State = core.Usable
	}
	if bytes.Equal(guardianAddress, userInfo.SecondGuardian.PublicKey) {
		if userInfo.SecondGuardian.State != core.NotUsableYet {
			return fmt.Errorf("%w for SecondGuardian, it is not in NotUsableYet state", ErrInvalidGuardianState)
		}
		userInfo.SecondGuardian.State = core.Usable
	}

	return resolver.marshalAndSave(userAddress, userInfo)
}

func (resolver *serviceResolver) validateTransactions(txs []erdData.Transaction, userAddress erdCore.AddressHandler) error {
	expectedGuardian := txs[0].GuardianAddr
	for _, tx := range txs {
		if tx.GuardianAddr != expectedGuardian {
			return ErrGuardianMismatch
		}

		err := resolver.validateOneTransaction(tx, userAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

func (resolver *serviceResolver) validateOneTransaction(tx erdData.Transaction, userAddress erdCore.AddressHandler) error {
	addr := userAddress.AddressAsBech32String()
	if tx.SndAddr != addr {
		return fmt.Errorf("%w, sender from credentials: %s, tx sender: %s", ErrInvalidSender, addr, tx.SndAddr)
	}

	userSig, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return err
	}

	return txcheck.VerifyTransactionSignature(
		&tx,
		userAddress.AddressBytes(),
		userSig,
		resolver.signatureVerifier,
		resolver.marshaller,
		resolver.txHasher,
	)
}

func (resolver *serviceResolver) getGuardianForTx(tx erdData.Transaction, userInfo *core.UserInfo) (core.GuardianInfo, error) {
	guardianForTx := core.GuardianInfo{}
	unknownGuardian := true
	firstGuardianAddr := resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey)
	if tx.GuardianAddr == firstGuardianAddr {
		guardianForTx = userInfo.FirstGuardian
		unknownGuardian = false
	}
	secondGuardianAddr := resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey)
	if tx.GuardianAddr == secondGuardianAddr {
		guardianForTx = userInfo.SecondGuardian
		unknownGuardian = false
	}

	if unknownGuardian {
		return core.GuardianInfo{}, fmt.Errorf("%w, guardian %s", ErrInvalidGuardian, tx.GuardianAddr)
	}

	if guardianForTx.State == core.NotUsableYet {
		return core.GuardianInfo{}, fmt.Errorf("%w, guardian %s", ErrGuardianNotYetUsable, tx.GuardianAddr)
	}

	return guardianForTx, nil
}

func (resolver *serviceResolver) validateCredentials(credentials string) (erdCore.AddressHandler, error) {
	err := resolver.credentialsHandler.Verify(credentials)
	if err != nil {
		return nil, err
	}

	accountAddress, err := resolver.credentialsHandler.GetAccountAddress(credentials)
	if err != nil {
		return nil, err
	}

	ctxGetAccount, cancelGetAccount := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancelGetAccount()
	_, err = resolver.proxy.GetAccount(ctxGetAccount, accountAddress)
	if err != nil {
		return nil, err
	}

	return accountAddress, nil
}

func (resolver *serviceResolver) handleNewAccount(userAddress []byte) (string, error) {
	index, err := resolver.indexHandler.AllocateIndex()
	if err != nil {
		return emptyAddress, err
	}

	privateKeys, err := resolver.keysGenerator.GenerateKeys(index)
	if err != nil {
		return emptyAddress, err
	}

	userInfo, err := resolver.computeDataAndSave(index, userAddress, privateKeys)
	if err != nil {
		return emptyAddress, err
	}

	return resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey), nil
}

func (resolver *serviceResolver) handleRegisteredAccount(userAddress []byte) (string, error) {
	userInfo, err := resolver.getUserInfo(userAddress)
	if err != nil {
		return emptyAddress, err
	}

	if userInfo.FirstGuardian.State == core.NotUsableYet {
		return resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey), nil
	}

	if userInfo.SecondGuardian.State == core.NotUsableYet {
		return resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey), nil
	}

	accountAddress := erdData.NewAddressFromBytes(userAddress)

	ctxGetGuardianData, cancelGetGuardianData := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancelGetGuardianData()
	guardianData, err := resolver.proxy.GetGuardianData(ctxGetGuardianData, accountAddress)
	if err != nil {
		return emptyAddress, err
	}

	nextGuardian := resolver.prepareNextGuardian(guardianData, userInfo)

	err = resolver.marshalAndSave(userAddress, userInfo)
	if err != nil {
		return emptyAddress, err
	}

	return nextGuardian, nil
}

func (resolver *serviceResolver) getUserInfo(userAddress []byte) (*core.UserInfo, error) {
	// TODO properly decrypt keys from DB
	// temporary unmarshal them
	userInfo := &core.UserInfo{}
	userInfoMarshalled, err := resolver.registeredUsersDB.Get(userAddress)
	if err != nil {
		return userInfo, err
	}

	err = resolver.marshaller.Unmarshal(&userInfo, userInfoMarshalled)
	if err != nil {
		return userInfo, err
	}

	return userInfo, nil
}

func (resolver *serviceResolver) computeDataAndSave(index uint32, userAddress []byte, privateKeys []crypto.PrivateKey) (*core.UserInfo, error) {
	firstGuardian, err := getGuardianInfoForKey(privateKeys[0])
	if err != nil {
		return &core.UserInfo{}, err
	}

	secondGuardian, err := getGuardianInfoForKey(privateKeys[1])
	if err != nil {
		return &core.UserInfo{}, err
	}

	userInfo := &core.UserInfo{
		Index:          index,
		FirstGuardian:  firstGuardian,
		SecondGuardian: secondGuardian,
	}

	err = resolver.marshalAndSave(userAddress, userInfo)
	if err != nil {
		return &core.UserInfo{}, err
	}

	return userInfo, nil
}

func (resolver *serviceResolver) marshalAndSave(userAddress []byte, userInfo *core.UserInfo) error {
	// TODO properly encrypt keys
	// temporary marshal them and save
	userInfoMarshalled, err := resolver.marshaller.Marshal(userInfo)
	if err != nil {
		return err
	}

	err = resolver.registeredUsersDB.Put(userAddress, userInfoMarshalled)
	if err != nil {
		return err
	}

	return nil
}

func (resolver *serviceResolver) prepareNextGuardian(guardianData *api.GuardianData, userInfo *core.UserInfo) string {
	firstGuardianOnChainState := resolver.getOnChainGuardianState(guardianData, userInfo.FirstGuardian)
	secondGuardianOnChainState := resolver.getOnChainGuardianState(guardianData, userInfo.SecondGuardian)
	isFirstOnChain := firstGuardianOnChainState != core.MissingGuardian
	isSecondOnChain := secondGuardianOnChainState != core.MissingGuardian
	if !isFirstOnChain && !isSecondOnChain {
		userInfo.FirstGuardian.State = core.NotUsableYet
		userInfo.SecondGuardian.State = core.NotUsableYet
		return resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey)
	}

	if isFirstOnChain && isSecondOnChain {
		if firstGuardianOnChainState == core.PendingGuardian {
			userInfo.FirstGuardian.State = core.NotUsableYet
			return resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey)
		}

		userInfo.SecondGuardian.State = core.NotUsableYet
		return resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey)
	}

	if isFirstOnChain {
		userInfo.SecondGuardian.State = core.NotUsableYet
		return resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey)
	}

	userInfo.FirstGuardian.State = core.NotUsableYet
	return resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey)
}

func (resolver *serviceResolver) getOnChainGuardianState(guardianData *api.GuardianData, guardian core.GuardianInfo) core.OnChainGuardianState {
	if check.IfNilReflect(guardianData) {
		return core.MissingGuardian
	}

	guardianAddress := resolver.pubKeyConverter.Encode(guardian.PublicKey)
	if !check.IfNilReflect(guardianData.ActiveGuardian) {
		isActiveGuardian := guardianData.ActiveGuardian.Address == guardianAddress
		if isActiveGuardian {
			return core.ActiveGuardian
		}
	}

	if !check.IfNilReflect(guardianData.PendingGuardian) {
		isPendingGuardian := guardianData.PendingGuardian.Address == guardianAddress
		if isPendingGuardian {
			return core.PendingGuardian
		}
	}

	return core.MissingGuardian
}

func getGuardianInfoForKey(privateKey crypto.PrivateKey) (core.GuardianInfo, error) {
	privateKeyBytes, err := privateKey.ToByteArray()
	if err != nil {
		return core.GuardianInfo{}, err
	}

	pk := privateKey.GeneratePublic()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return core.GuardianInfo{}, err
	}

	return core.GuardianInfo{
		PublicKey:  pkBytes,
		PrivateKey: privateKeyBytes,
		State:      core.NotUsableYet,
	}, nil
}

func (resolver *serviceResolver) extractUserTagForQRGeneration(tag string, prettyUserAddress string) string {
	if len(tag) > 0 {
		return tag
	}
	return prettyUserAddress
}

// IsInterfaceNil return true if there is no value under the interface
func (resolver *serviceResolver) IsInterfaceNil() bool {
	return resolver == nil
}
