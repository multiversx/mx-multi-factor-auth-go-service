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
	"github.com/ElrondNetwork/elrond-go-crypto/encryption/x25519"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain/cryptoProvider"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
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
	Provider          providers.Provider
	Proxy             blockchain.Proxy
	KeysGenerator     core.KeysGenerator
	PubKeyConverter   core.PubkeyConverter
	GogoMarshaller    core.Marshaller
	JsonMarshaller    core.Marshaller
	JsonTxMarshaller  core.Marshaller
	TxHasher          data.Hasher
	SignatureVerifier builders.Signer
	GuardedTxBuilder  core.GuardedTxBuilder
	RequestTime       time.Duration
	RegisteredUsersDB core.ShardedStorageWithIndex
	KeyGen            crypto.KeyGenerator
}

type serviceResolver struct {
	provider          providers.Provider
	proxy             blockchain.Proxy
	keysGenerator     core.KeysGenerator
	pubKeyConverter   core.PubkeyConverter
	gogoMarshaller    core.Marshaller
	jsonMarshaller    core.Marshaller
	jsonTxMarshaller  core.Marshaller
	txHasher          data.Hasher
	requestTime       time.Duration
	signatureVerifier builders.Signer
	guardedTxBuilder  core.GuardedTxBuilder
	registeredUsersDB core.ShardedStorageWithIndex
	managedPrivateKey crypto.PrivateKey
	keyGen            crypto.KeyGenerator

	newCryptoComponentsHolderHandler func(keyGen crypto.KeyGenerator, skBytes []byte) (erdCore.CryptoComponentsHolder, error)
}

// NewServiceResolver returns a new instance of service resolver
func NewServiceResolver(args ArgServiceResolver) (*serviceResolver, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	resolver := &serviceResolver{
		provider:          args.Provider,
		proxy:             args.Proxy,
		keysGenerator:     args.KeysGenerator,
		pubKeyConverter:   args.PubKeyConverter,
		gogoMarshaller:    args.GogoMarshaller,
		jsonMarshaller:    args.JsonMarshaller,
		jsonTxMarshaller:  args.JsonTxMarshaller,
		txHasher:          args.TxHasher,
		requestTime:       args.RequestTime,
		signatureVerifier: args.SignatureVerifier,
		guardedTxBuilder:  args.GuardedTxBuilder,
		registeredUsersDB: args.RegisteredUsersDB,
		keyGen:            args.KeyGen,
	}
	resolver.newCryptoComponentsHolderHandler = resolver.newCryptoComponentsHolder
	resolver.managedPrivateKey, err = resolver.keysGenerator.GenerateManagedKey()
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func checkArgs(args ArgServiceResolver) error {
	if check.IfNil(args.Provider) {
		return ErrNilProvider
	}
	if check.IfNil(args.Proxy) {
		return ErrNilProxy
	}
	if check.IfNil(args.KeysGenerator) {
		return ErrNilKeysGenerator
	}
	if check.IfNil(args.PubKeyConverter) {
		return ErrNilPubKeyConverter
	}
	if check.IfNil(args.GogoMarshaller) {
		return ErrNilMarshaller
	}
	if check.IfNil(args.JsonMarshaller) {
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
	if check.IfNil(args.RegisteredUsersDB) {
		return fmt.Errorf("%w for registered users", ErrNilDB)
	}
	if check.IfNil(args.KeyGen) {
		return ErrNilKeyGenerator
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
func (resolver *serviceResolver) RegisterUser(userAddress erdCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
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
func (resolver *serviceResolver) VerifyCode(userAddress erdCore.AddressHandler, request requests.VerificationPayload) error {
	err := resolver.provider.ValidateCode(userAddress.AddressAsBech32String(), request.Guardian, request.Code)
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
func (resolver *serviceResolver) SignTransaction(userAddress erdCore.AddressHandler, request requests.SignTransaction) ([]byte, error) {
	guardian, err := resolver.validateTxRequestReturningGuardian(userAddress, request.Code, []erdData.Transaction{request.Tx})
	if err != nil {
		return nil, err
	}

	guardianCryptoHolder, err := resolver.newCryptoComponentsHolderHandler(resolver.keyGen, guardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	err = resolver.guardedTxBuilder.ApplyGuardianSignature(guardianCryptoHolder, &request.Tx)
	if err != nil {
		return nil, err
	}

	return resolver.jsonTxMarshaller.Marshal(&request.Tx)
}

// SignMultipleTransactions validates user's transactions, then adds guardian signature and returns the transaction
func (resolver *serviceResolver) SignMultipleTransactions(userAddress erdCore.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
	guardian, err := resolver.validateTxRequestReturningGuardian(userAddress, request.Code, request.Txs)
	if err != nil {
		return nil, err
	}

	guardianCryptoHolder, err := resolver.newCryptoComponentsHolderHandler(resolver.keyGen, guardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	txsSlice := make([][]byte, 0)
	for _, tx := range request.Txs {
		err = resolver.guardedTxBuilder.ApplyGuardianSignature(guardianCryptoHolder, &tx)
		if err != nil {
			return nil, err
		}

		txBuff, err := resolver.jsonTxMarshaller.Marshal(&tx)
		if err != nil {
			return nil, err
		}

		txsSlice = append(txsSlice, txBuff)
	}

	return txsSlice, nil
}

func (resolver *serviceResolver) validateTxRequestReturningGuardian(userAddress erdCore.AddressHandler, code string, txs []erdData.Transaction) (core.GuardianInfo, error) {
	err := resolver.validateTransactions(txs, userAddress)
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
		if userInfo.FirstGuardian.State != core.NotUsable {
			return fmt.Errorf("%w for FirstGuardian, it is not in NotUsable state", ErrInvalidGuardianState)
		}
		userInfo.FirstGuardian.State = core.Usable
	}
	if bytes.Equal(guardianAddress, userInfo.SecondGuardian.PublicKey) {
		if userInfo.SecondGuardian.State != core.NotUsable {
			return fmt.Errorf("%w for SecondGuardian, it is not in NotUsable state", ErrInvalidGuardianState)
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

	userPublicKey, err := resolver.keyGen.PublicKeyFromByteArray(userAddress.AddressBytes())
	if err != nil {
		return err
	}

	return txcheck.VerifyTransactionSignature(
		&tx,
		userPublicKey,
		userSig,
		resolver.signatureVerifier,
		resolver.jsonTxMarshaller,
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

	if guardianForTx.State == core.NotUsable {
		return core.GuardianInfo{}, fmt.Errorf("%w, guardian %s", ErrGuardianNotUsable, tx.GuardianAddr)
	}

	return guardianForTx, nil
}

func (resolver *serviceResolver) handleNewAccount(userAddress []byte) (string, error) {
	index, err := resolver.registeredUsersDB.AllocateIndex(userAddress)
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

	if userInfo.FirstGuardian.State == core.NotUsable {
		return resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey), nil
	}

	if userInfo.SecondGuardian.State == core.NotUsable {
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
	userInfo := &core.UserInfo{}
	encryptedData := &x25519.EncryptedData{}
	encryptedDataMarshalled, err := resolver.registeredUsersDB.Get(userAddress)
	if err != nil {
		return userInfo, err
	}

	err = resolver.jsonMarshaller.Unmarshal(encryptedData, encryptedDataMarshalled)
	if err != nil {
		return userInfo, err
	}

	userInfoMarshalled, err := encryptedData.Decrypt(resolver.managedPrivateKey)
	if err != nil {
		return userInfo, err
	}

	err = resolver.gogoMarshaller.Unmarshal(userInfo, userInfoMarshalled)
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
	userInfoMarshalled, err := resolver.gogoMarshaller.Marshal(userInfo)
	if err != nil {
		return err
	}

	encryptedData := &x25519.EncryptedData{}
	encryptionSk, _ := resolver.keyGen.GeneratePair()
	err = encryptedData.Encrypt(userInfoMarshalled, resolver.managedPrivateKey.GeneratePublic(), encryptionSk)
	if err != nil {
		return err
	}

	encryptedDataBytes, err := resolver.jsonMarshaller.Marshal(encryptedData)
	if err != nil {
		return err
	}

	err = resolver.registeredUsersDB.Put(userAddress, encryptedDataBytes)
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
		userInfo.FirstGuardian.State = core.NotUsable
		userInfo.SecondGuardian.State = core.NotUsable
		return resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey)
	}

	if isFirstOnChain && isSecondOnChain {
		if firstGuardianOnChainState == core.PendingGuardian {
			userInfo.FirstGuardian.State = core.NotUsable
			return resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey)
		}

		userInfo.SecondGuardian.State = core.NotUsable
		return resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey)
	}

	if isFirstOnChain {
		userInfo.SecondGuardian.State = core.NotUsable
		return resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey)
	}

	userInfo.FirstGuardian.State = core.NotUsable
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
		State:      core.NotUsable,
	}, nil
}

func (resolver *serviceResolver) extractUserTagForQRGeneration(tag string, prettyUserAddress string) string {
	if len(tag) > 0 {
		return tag
	}
	return prettyUserAddress
}

func (resolver *serviceResolver) newCryptoComponentsHolder(keyGen crypto.KeyGenerator, skBytes []byte) (erdCore.CryptoComponentsHolder, error) {
	return cryptoProvider.NewCryptoComponentsHolder(keyGen, skBytes)
}

// IsInterfaceNil return true if there is no value under the interface
func (resolver *serviceResolver) IsInterfaceNil() bool {
	return resolver == nil
}
