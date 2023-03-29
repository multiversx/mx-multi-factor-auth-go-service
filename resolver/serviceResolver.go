package resolver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/core/sync"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/builders"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/txcheck"
)

var log = logger.GetOrCreate("serviceresolver")

const (
	minRequestTime            = time.Second
	zeroBalance               = "0"
	minDelayBetweenOTPUpdates = 1
)

// ArgServiceResolver is the DTO used to create a new instance of service resolver
type ArgServiceResolver struct {
	UserEncryptor                 UserEncryptor
	TOTPHandler                   handlers.TOTPHandler
	Proxy                         blockchain.Proxy
	KeysGenerator                 core.KeysGenerator
	PubKeyConverter               core.PubkeyConverter
	UserDataMarshaller            core.Marshaller
	TxMarshaller                  core.Marshaller
	TxHasher                      data.Hasher
	SignatureVerifier             builders.Signer
	GuardedTxBuilder              core.GuardedTxBuilder
	RequestTime                   time.Duration
	RegisteredUsersDB             core.ShardedStorageWithIndex
	KeyGen                        crypto.KeyGenerator
	CryptoComponentsHolderFactory CryptoComponentsHolderFactory
	SkipTxUserSigVerify           bool
	DelayBetweenOTPUpdatesInSec   int64
}

type serviceResolver struct {
	userEncryptor                 UserEncryptor
	totpHandler                   handlers.TOTPHandler
	proxy                         blockchain.Proxy
	keysGenerator                 core.KeysGenerator
	pubKeyConverter               core.PubkeyConverter
	userDataMarshaller            core.Marshaller
	txMarshaller                  core.Marshaller
	txHasher                      data.Hasher
	requestTime                   time.Duration
	signatureVerifier             builders.Signer
	guardedTxBuilder              core.GuardedTxBuilder
	registeredUsersDB             core.ShardedStorageWithIndex
	keyGen                        crypto.KeyGenerator
	cryptoComponentsHolderFactory CryptoComponentsHolderFactory
	skipTxUserSigVerify           bool
	delayBetweenOTPUpdatesInSec   int64

	userCritSection sync.KeyRWMutexHandler
}

// NewServiceResolver returns a new instance of service resolver
func NewServiceResolver(args ArgServiceResolver) (*serviceResolver, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	resolver := &serviceResolver{
		userEncryptor:                 args.UserEncryptor,
		totpHandler:                   args.TOTPHandler,
		proxy:                         args.Proxy,
		keysGenerator:                 args.KeysGenerator,
		pubKeyConverter:               args.PubKeyConverter,
		userDataMarshaller:            args.UserDataMarshaller,
		txMarshaller:                  args.TxMarshaller,
		txHasher:                      args.TxHasher,
		requestTime:                   args.RequestTime,
		signatureVerifier:             args.SignatureVerifier,
		guardedTxBuilder:              args.GuardedTxBuilder,
		registeredUsersDB:             args.RegisteredUsersDB,
		keyGen:                        args.KeyGen,
		cryptoComponentsHolderFactory: args.CryptoComponentsHolderFactory,
		skipTxUserSigVerify:           args.SkipTxUserSigVerify,
		delayBetweenOTPUpdatesInSec:   args.DelayBetweenOTPUpdatesInSec,
	}
	resolver.userCritSection = sync.NewKeyRWMutex()
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func checkArgs(args ArgServiceResolver) error {
	if check.IfNil(args.UserEncryptor) {
		return ErrNilUserEncryptor
	}
	if check.IfNil(args.TOTPHandler) {
		return ErrNilTOTPHandler
	}
	if check.IfNil(args.Proxy) {
		return ErrNilProxy
	}
	if check.IfNil(args.PubKeyConverter) {
		return ErrNilPubKeyConverter
	}
	if check.IfNil(args.UserDataMarshaller) {
		return fmt.Errorf("%w for userData marshaller", ErrNilMarshaller)
	}
	if check.IfNil(args.TxMarshaller) {
		return fmt.Errorf("%w for tx marshaller", ErrNilMarshaller)
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
	if check.IfNil(args.CryptoComponentsHolderFactory) {
		return ErrNilCryptoComponentsHolderFactory
	}
	if args.DelayBetweenOTPUpdatesInSec < minDelayBetweenOTPUpdates {
		return fmt.Errorf("%w for DelayBetweenOTPUpdatesInSec, got %d, min expected %d",
			ErrInvalidValue, args.DelayBetweenOTPUpdatesInSec, minDelayBetweenOTPUpdates)
	}

	return nil
}

// RegisterUser creates a new OTP for the given provider
// and (optionally) returns some information required for the user to set up the OTP on his end (eg: QR code).
func (resolver *serviceResolver) RegisterUser(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
	tag := resolver.extractUserTagForQRGeneration(request.Tag, userAddress.Pretty())
	otp, err := resolver.totpHandler.CreateTOTP(tag)
	if err != nil {
		return nil, "", err
	}

	qr, err := otp.QR()
	if err != nil {
		return nil, "", err
	}

	guardianAddress, err := resolver.getGuardianAddressAndRegisterIfNewUser(userAddress, otp)
	if err != nil {
		return nil, "", err
	}

	return qr, resolver.pubKeyConverter.Encode(guardianAddress), nil
}

// VerifyCode validates the code received
func (resolver *serviceResolver) VerifyCode(userAddress sdkCore.AddressHandler, request requests.VerificationPayload) error {
	guardianAddr, err := resolver.pubKeyConverter.Decode(request.Guardian)
	if err != nil {
		return err
	}

	userInfo, err := resolver.getUserInfo(userAddress.AddressBytes())
	if err != nil {
		return err
	}

	err = resolver.verifyCode(userInfo, request.Code, guardianAddr)
	if err != nil {
		return err
	}

	return resolver.updateGuardianStateIfNeeded(userAddress.AddressBytes(), guardianAddr)
}

// SignTransaction validates user's transaction, then adds guardian signature and returns the transaction
func (resolver *serviceResolver) SignTransaction(userAddress sdkCore.AddressHandler, request requests.SignTransaction) ([]byte, error) {
	guardian, err := resolver.validateTxRequestReturningGuardian(userAddress, request.Code, []sdkData.Transaction{request.Tx})
	if err != nil {
		return nil, err
	}

	guardianCryptoHolder, err := resolver.cryptoComponentsHolderFactory.Create(guardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	err = resolver.guardedTxBuilder.ApplyGuardianSignature(guardianCryptoHolder, &request.Tx)
	if err != nil {
		return nil, err
	}

	return resolver.txMarshaller.Marshal(&request.Tx)
}

// SignMultipleTransactions validates user's transactions, then adds guardian signature and returns the transaction
func (resolver *serviceResolver) SignMultipleTransactions(userAddress sdkCore.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
	guardian, err := resolver.validateTxRequestReturningGuardian(userAddress, request.Code, request.Txs)
	if err != nil {
		return nil, err
	}

	guardianCryptoHolder, err := resolver.cryptoComponentsHolderFactory.Create(guardian.PrivateKey)
	if err != nil {
		return nil, err
	}

	txsSlice := make([][]byte, 0)
	for index, tx := range request.Txs {
		err = resolver.guardedTxBuilder.ApplyGuardianSignature(guardianCryptoHolder, &tx)
		if err != nil {
			return nil, fmt.Errorf("%w for transaction #%d", err, index)
		}

		txBuff, err := resolver.txMarshaller.Marshal(&tx)
		if err != nil {
			return nil, fmt.Errorf("%w for transaction #%d", err, index)
		}

		txsSlice = append(txsSlice, txBuff)
	}

	return txsSlice, nil
}

// RegisteredUsers returns the number of registered users
func (resolver *serviceResolver) RegisteredUsers() (uint32, error) {
	return resolver.registeredUsersDB.Count()
}

func (resolver *serviceResolver) validateUserAddress(userAddress sdkCore.AddressHandler) error {
	ctx, cancel := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancel()
	account, err := resolver.proxy.GetAccount(ctx, userAddress)
	if err != nil {
		return err
	}

	if !hasBalance(account.Balance) {
		return fmt.Errorf("%w for account %s", ErrNoBalance, userAddress.AddressAsBech32String())
	}

	return nil
}

func (resolver *serviceResolver) verifyCode(userInfo *core.UserInfo, userCode string, guardianAddr []byte) error {
	otpHandler, err := resolver.getUserOTPHandler(userInfo, guardianAddr)
	if err != nil {
		return err
	}

	return otpHandler.Validate(userCode)
}

func (resolver *serviceResolver) getUserOTPHandler(userInfo *core.UserInfo, guardianAddr []byte) (handlers.OTP, error) {
	otpInfo, err := extractOtpForGuardian(userInfo, guardianAddr)
	if err != nil {
		return nil, err
	}

	return resolver.totpHandler.TOTPFromBytes(otpInfo.OTP)
}

func extractOtpForGuardian(userInfo *core.UserInfo, guardian []byte) (*core.OTPInfo, error) {
	if userInfo == nil {
		return nil, handlers.ErrNilUserInfo
	}

	if bytes.Equal(userInfo.FirstGuardian.PublicKey, guardian) {
		return &userInfo.FirstGuardian.OTPData, nil
	}

	if bytes.Equal(userInfo.SecondGuardian.PublicKey, guardian) {
		return &userInfo.SecondGuardian.OTPData, nil
	}

	return nil, ErrInvalidGuardian
}

func hasBalance(balance string) bool {
	missingBalance := len(balance) == 0
	hasZeroBalance := balance == zeroBalance
	return !missingBalance && !hasZeroBalance
}

// getGuardianAddressAndRegisterIfNewUser returns the address of a unique guardian
func (resolver *serviceResolver) getGuardianAddressAndRegisterIfNewUser(userAddress sdkCore.AddressHandler, otp handlers.OTP) ([]byte, error) {
	addressBytes := userAddress.AddressBytes()

	resolver.userCritSection.Lock(string(addressBytes))
	defer resolver.userCritSection.Unlock(string(addressBytes))

	err := resolver.registeredUsersDB.Has(addressBytes)
	if err != nil {
		return resolver.handleNewAccount(userAddress, otp)
	}

	return resolver.handleRegisteredAccount(userAddress, otp)
}

// TODO: add limits for the number of transactions that can be verified at once
func (resolver *serviceResolver) validateTxRequestReturningGuardian(userAddress sdkCore.AddressHandler, code string, txs []sdkData.Transaction) (core.GuardianInfo, error) {
	err := resolver.validateTransactions(txs, userAddress)
	if err != nil {
		return core.GuardianInfo{}, err
	}

	// only validate the guardian for first tx, as all of them must have the same one
	guardianAddr, err := resolver.pubKeyConverter.Decode(txs[0].GuardianAddr)
	if err != nil {
		return core.GuardianInfo{}, err
	}

	userInfo, err := resolver.getUserInfo(userAddress.AddressBytes())
	if err != nil {
		return core.GuardianInfo{}, err
	}

	err = resolver.verifyCode(userInfo, code, guardianAddr)
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
		if userInfo.FirstGuardian.State == core.NotUsable {
			userInfo.FirstGuardian.State = core.Usable
			return resolver.marshalAndSaveEncrypted(userAddress, userInfo)
		}
	}
	if bytes.Equal(guardianAddress, userInfo.SecondGuardian.PublicKey) {
		if userInfo.SecondGuardian.State == core.NotUsable {
			userInfo.SecondGuardian.State = core.Usable
			return resolver.marshalAndSaveEncrypted(userAddress, userInfo)
		}
	}

	return nil
}

func (resolver *serviceResolver) validateTransactions(txs []sdkData.Transaction, userAddress sdkCore.AddressHandler) error {
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

func (resolver *serviceResolver) validateOneTransaction(tx sdkData.Transaction, userAddress sdkCore.AddressHandler) error {
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

	if resolver.skipTxUserSigVerify {
		return nil
	}

	return txcheck.VerifyTransactionSignature(
		&tx,
		userPublicKey,
		userSig,
		resolver.signatureVerifier,
		resolver.txMarshaller,
		resolver.txHasher,
	)
}

func (resolver *serviceResolver) getGuardianForTx(tx sdkData.Transaction, userInfo *core.UserInfo) (core.GuardianInfo, error) {
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

func (resolver *serviceResolver) handleNewAccount(userAddress sdkCore.AddressHandler, otp handlers.OTP) ([]byte, error) {
	err := resolver.validateUserAddress(userAddress)
	if err != nil {
		return nil, err
	}

	addressBytes := userAddress.AddressBytes()

	index, err := resolver.registeredUsersDB.AllocateIndex(addressBytes)
	if err != nil {
		return nil, err
	}

	privateKeys, err := resolver.keysGenerator.GenerateKeys(index)
	if err != nil {
		return nil, err
	}

	userInfo, err := resolver.computeDataAndSave(index, addressBytes, privateKeys, otp)
	if err != nil {
		return nil, err
	}

	log.Debug("registering new user",
		"userAddress", userAddress.AddressAsBech32String(),
		"guardian", resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey),
		"index", index)

	return userInfo.FirstGuardian.PublicKey, nil
}

func (resolver *serviceResolver) handleRegisteredAccount(userAddress sdkCore.AddressHandler, otp handlers.OTP) ([]byte, error) {
	addressBytes := userAddress.AddressBytes()
	userInfo, err := resolver.getUserInfo(addressBytes)
	if err != nil {
		return nil, err
	}

	if userInfo.FirstGuardian.State == core.NotUsable {
		log.Debug("registering old user",
			"userAddress", userAddress.AddressAsBech32String(),
			"newGuardian", resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey))
		return userInfo.FirstGuardian.PublicKey, nil
	}

	if userInfo.SecondGuardian.State == core.NotUsable {
		log.Debug("registering old user",
			"userAddress", userAddress.AddressAsBech32String(),
			"newGuardian", resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey))
		return userInfo.SecondGuardian.PublicKey, nil
	}

	accountAddress := sdkData.NewAddressFromBytes(addressBytes)

	ctxGetGuardianData, cancelGetGuardianData := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancelGetGuardianData()
	guardianData, err := resolver.proxy.GetGuardianData(ctxGetGuardianData, accountAddress)
	if err != nil {
		return nil, err
	}

	nextGuardian := resolver.prepareNextGuardian(guardianData, userInfo)
	err = resolver.addOTPToUserGuardian(userInfo, nextGuardian, otp)
	if err != nil {
		return nil, err
	}

	err = resolver.marshalAndSaveEncrypted(addressBytes, userInfo)
	if err != nil {
		return nil, err
	}

	printableGuardianData := ""
	guardianDataBuff, err := json.Marshal(guardianData)
	if err == nil {
		printableGuardianData = string(guardianDataBuff)
	}

	log.Debug("registering old user",
		"userAddress", userAddress.AddressAsBech32String(),
		"newGuardian", resolver.pubKeyConverter.Encode(nextGuardian),
		"fetched data from chain", printableGuardianData)

	return nextGuardian, nil
}

func (resolver *serviceResolver) addOTPToUserGuardian(userInfo *core.UserInfo, guardian []byte, otp handlers.OTP) error {
	if userInfo == nil {
		return ErrNilUserInfo
	}

	var selectedGuardianInfo *core.GuardianInfo
	if bytes.Equal(userInfo.FirstGuardian.PublicKey, guardian) {
		selectedGuardianInfo = &userInfo.FirstGuardian
	}

	if bytes.Equal(userInfo.SecondGuardian.PublicKey, guardian) {
		selectedGuardianInfo = &userInfo.SecondGuardian
	}

	if selectedGuardianInfo == nil {
		return ErrInvalidGuardian
	}

	var err error
	currentTimestamp := time.Now().Unix()
	oldOTPInfo := &selectedGuardianInfo.OTPData
	allowedToChangeOTP := oldOTPInfo.LastTOTPChangeTimestamp+resolver.delayBetweenOTPUpdatesInSec < currentTimestamp
	if !allowedToChangeOTP {
		return fmt.Errorf("%w, last update was %d seconds ago",
			handlers.ErrRegistrationFailed, currentTimestamp-oldOTPInfo.LastTOTPChangeTimestamp)
	}

	otpBytes, err := otp.ToBytes()
	if err != nil {
		return err
	}

	selectedGuardianInfo.OTPData.OTP = otpBytes
	selectedGuardianInfo.OTPData.LastTOTPChangeTimestamp = currentTimestamp

	return nil
}

func (resolver *serviceResolver) getUserInfo(userAddress []byte) (*core.UserInfo, error) {
	encryptedDataMarshalled, err := resolver.registeredUsersDB.Get(userAddress)
	if err != nil {
		return nil, err
	}

	return resolver.unmarshalAndDecryptUserInfo(encryptedDataMarshalled)
}

func (resolver *serviceResolver) encryptAndMarshalUserInfo(userInfo *core.UserInfo) ([]byte, error) {
	encryptedUserInfo, err := resolver.userEncryptor.EncryptUserInfo(userInfo)
	if err != nil {
		return nil, err
	}

	return resolver.userDataMarshaller.Marshal(encryptedUserInfo)
}

func (resolver *serviceResolver) unmarshalAndDecryptUserInfo(encryptedDataMarshalled []byte) (*core.UserInfo, error) {
	userInfo := &core.UserInfo{}
	err := resolver.userDataMarshaller.Unmarshal(userInfo, encryptedDataMarshalled)
	if err != nil {
		return nil, err
	}

	return resolver.userEncryptor.DecryptUserInfo(userInfo)
}

func (resolver *serviceResolver) computeDataAndSave(index uint32, userAddress []byte, privateKeys []crypto.PrivateKey, otp handlers.OTP) (*core.UserInfo, error) {
	firstGuardian, err := getGuardianInfoForKey(privateKeys[0])
	if err != nil {
		return nil, err
	}

	// ToBytes does also the encryption for the otp data
	firstGuardian.OTPData.OTP, err = otp.ToBytes()
	if err != nil {
		return nil, err
	}
	firstGuardian.OTPData.LastTOTPChangeTimestamp = time.Now().Unix()

	secondGuardian, err := getGuardianInfoForKey(privateKeys[1])
	if err != nil {
		return nil, err
	}

	userInfo := &core.UserInfo{
		Index:          index,
		FirstGuardian:  firstGuardian,
		SecondGuardian: secondGuardian,
	}

	err = resolver.marshalAndSaveEncrypted(userAddress, userInfo)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (resolver *serviceResolver) marshalAndSaveEncrypted(userAddress []byte, userInfo *core.UserInfo) error {
	encryptedDataBytes, err := resolver.encryptAndMarshalUserInfo(userInfo)
	if err != nil {
		return err
	}

	err = resolver.registeredUsersDB.Put(userAddress, encryptedDataBytes)
	if err != nil {
		return err
	}

	return nil
}

func (resolver *serviceResolver) prepareNextGuardian(guardianData *api.GuardianData, userInfo *core.UserInfo) []byte {
	firstGuardianOnChainState := resolver.getOnChainGuardianState(guardianData, userInfo.FirstGuardian)
	secondGuardianOnChainState := resolver.getOnChainGuardianState(guardianData, userInfo.SecondGuardian)
	isFirstOnChain := firstGuardianOnChainState != core.MissingGuardian
	isSecondOnChain := secondGuardianOnChainState != core.MissingGuardian
	if !isFirstOnChain && !isSecondOnChain {
		userInfo.FirstGuardian.State = core.NotUsable
		userInfo.SecondGuardian.State = core.NotUsable
		return userInfo.FirstGuardian.PublicKey
	}

	if isFirstOnChain && isSecondOnChain {
		if firstGuardianOnChainState == core.PendingGuardian {
			userInfo.FirstGuardian.State = core.NotUsable
			return userInfo.FirstGuardian.PublicKey
		}

		userInfo.SecondGuardian.State = core.NotUsable
		return userInfo.SecondGuardian.PublicKey
	}

	if isFirstOnChain {
		userInfo.SecondGuardian.State = core.NotUsable
		return userInfo.SecondGuardian.PublicKey
	}

	userInfo.FirstGuardian.State = core.NotUsable
	return userInfo.FirstGuardian.PublicKey
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

	OTPData := core.OTPInfo{
		OTP:                     []byte{},
		LastTOTPChangeTimestamp: 0,
	}
	return core.GuardianInfo{
		PublicKey:  pkBytes,
		PrivateKey: privateKeyBytes,
		State:      core.NotUsable,
		OTPData:    OTPData,
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
