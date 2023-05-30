package resolver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/core/sync"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/txcheck"
)

var log = logger.GetOrCreate("serviceresolver")

const (
	minRequestTimeInSec       = 1
	zeroBalance               = "0"
	minDelayBetweenOTPUpdates = 1
	minTransactionsAllowed    = 1
)

// ArgServiceResolver is the DTO used to create a new instance of service resolver
type ArgServiceResolver struct {
	UserEncryptor                 UserEncryptor
	TOTPHandler                   handlers.TOTPHandler
	FrozenOtpHandler              handlers.FrozenOtpHandler
	HttpClientWrapper             core.HttpClientWrapper
	KeysGenerator                 core.KeysGenerator
	PubKeyConverter               core.PubkeyConverter
	UserDataMarshaller            core.Marshaller
	TxMarshaller                  core.Marshaller
	TxHasher                      data.Hasher
	SignatureVerifier             builders.Signer
	GuardedTxBuilder              core.GuardedTxBuilder
	RegisteredUsersDB             core.StorageWithIndex
	KeyGen                        crypto.KeyGenerator
	CryptoComponentsHolderFactory CryptoComponentsHolderFactory
	Config                        config.ServiceResolverConfig
}

type serviceResolver struct {
	userEncryptor                 UserEncryptor
	totpHandler                   handlers.TOTPHandler
	frozenOtpHandler              handlers.FrozenOtpHandler
	httpClientWrapper             core.HttpClientWrapper
	keysGenerator                 core.KeysGenerator
	pubKeyConverter               core.PubkeyConverter
	userDataMarshaller            core.Marshaller
	txMarshaller                  core.Marshaller
	txHasher                      data.Hasher
	requestTime                   time.Duration
	signatureVerifier             builders.Signer
	guardedTxBuilder              core.GuardedTxBuilder
	registeredUsersDB             core.StorageWithIndex
	keyGen                        crypto.KeyGenerator
	cryptoComponentsHolderFactory CryptoComponentsHolderFactory
	config                        config.ServiceResolverConfig

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
		frozenOtpHandler:              args.FrozenOtpHandler,
		httpClientWrapper:             args.HttpClientWrapper,
		keysGenerator:                 args.KeysGenerator,
		pubKeyConverter:               args.PubKeyConverter,
		userDataMarshaller:            args.UserDataMarshaller,
		txMarshaller:                  args.TxMarshaller,
		txHasher:                      args.TxHasher,
		requestTime:                   time.Duration(args.Config.RequestTimeInSeconds) * time.Second,
		signatureVerifier:             args.SignatureVerifier,
		guardedTxBuilder:              args.GuardedTxBuilder,
		registeredUsersDB:             args.RegisteredUsersDB,
		keyGen:                        args.KeyGen,
		cryptoComponentsHolderFactory: args.CryptoComponentsHolderFactory,
		config:                        args.Config,
		userCritSection:               sync.NewKeyRWMutex(),
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
	if check.IfNil(args.FrozenOtpHandler) {
		return ErrNilFrozenOtpHandler
	}
	if check.IfNil(args.HttpClientWrapper) {
		return ErrNilHTTPClientWrapper
	}
	if check.IfNil(args.KeysGenerator) {
		return ErrNilKeysGenerator
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
	if args.Config.RequestTimeInSeconds < minRequestTimeInSec {
		return fmt.Errorf("%w for RequestTimeInSeconds, received %d, min expected %d", ErrInvalidValue, args.Config.RequestTimeInSeconds, minRequestTimeInSec)
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
	if args.Config.DelayBetweenOTPWritesInSec < minDelayBetweenOTPUpdates {
		return fmt.Errorf("%w for DelayBetweenOTPWritesInSec, got %d, min expected %d",
			ErrInvalidValue, args.Config.DelayBetweenOTPWritesInSec, minDelayBetweenOTPUpdates)
	}
	if args.Config.MaxTransactionsAllowedForSigning < minTransactionsAllowed {
		return fmt.Errorf("%w for MaxTransactionsAllowedForSigning, got %d, min expected %d",
			ErrInvalidValue, args.Config.MaxTransactionsAllowedForSigning, minTransactionsAllowed)
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
func (resolver *serviceResolver) VerifyCode(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error) {
	guardianAddr, err := resolver.pubKeyConverter.Decode(request.Guardian)
	if err != nil {
		return requests.DefaultOTPCodeVerifyData(), err
	}

	addressBytes := userAddress.AddressBytes()
	resolver.userCritSection.Lock(string(addressBytes))
	defer resolver.userCritSection.Unlock(string(addressBytes))

	userInfo, err := resolver.getUserInfo(addressBytes)
	if err != nil {
		return requests.DefaultOTPCodeVerifyData(), err
	}

	verifyCodeData, err := resolver.verifyCode(userInfo, userAddress.AddressAsBech32String(), userIp, request.Code, guardianAddr)
	if err != nil {
		return verifyCodeData, err
	}

	err = resolver.updateGuardianStateIfNeeded(userAddress.AddressBytes(), userInfo, guardianAddr)
	if err != nil {
		return verifyCodeData, err
	}

	log.Debug("code ok",
		"userAddress", userAddress.AddressAsBech32String(),
		"guardian", request.Guardian)

	return verifyCodeData, nil
}

// SignTransaction validates user's transaction, then adds guardian signature and returns the transaction
func (resolver *serviceResolver) SignTransaction(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error) {
	guardian, otpCodeVerifyData, err := resolver.validateTxRequestReturningGuardian(userIp, request.Code, []sdkData.Transaction{request.Tx})
	if err != nil {
		return nil, otpCodeVerifyData, err
	}

	guardianCryptoHolder, err := resolver.cryptoComponentsHolderFactory.Create(guardian.PrivateKey)
	if err != nil {
		return nil, otpCodeVerifyData, err
	}

	err = resolver.guardedTxBuilder.ApplyGuardianSignature(guardianCryptoHolder, &request.Tx)
	if err != nil {
		return nil, otpCodeVerifyData, err
	}

	txBytes, err := resolver.txMarshaller.Marshal(&request.Tx)
	return txBytes, otpCodeVerifyData, err
}

// SignMultipleTransactions validates user's transactions, then adds guardian signature and returns the transaction
func (resolver *serviceResolver) SignMultipleTransactions(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error) {
	guardian, otpCodeVerifyData, err := resolver.validateTxRequestReturningGuardian(userIp, request.Code, request.Txs)
	if err != nil {
		return nil, otpCodeVerifyData, err
	}

	guardianCryptoHolder, err := resolver.cryptoComponentsHolderFactory.Create(guardian.PrivateKey)
	if err != nil {
		return nil, otpCodeVerifyData, err
	}

	txsSlice := make([][]byte, 0)
	for index, tx := range request.Txs {
		err = resolver.guardedTxBuilder.ApplyGuardianSignature(guardianCryptoHolder, &tx)
		if err != nil {
			return nil, otpCodeVerifyData, fmt.Errorf("%w for transaction #%d", err, index)
		}

		txBuff, err := resolver.txMarshaller.Marshal(&tx)
		if err != nil {
			return nil, otpCodeVerifyData, fmt.Errorf("%w for transaction #%d", err, index)
		}

		txsSlice = append(txsSlice, txBuff)
	}

	return txsSlice, otpCodeVerifyData, nil
}

// RegisteredUsers returns the number of registered users
func (resolver *serviceResolver) RegisteredUsers() (uint32, error) {
	return resolver.registeredUsersDB.Count()
}

// TcsConfig returns the current configuration of the TCS
func (resolver *serviceResolver) TcsConfig() *core.TcsConfig {
	return &core.TcsConfig{
		OTPDelay:         resolver.config.DelayBetweenOTPWritesInSec,
		BackoffWrongCode: resolver.frozenOtpHandler.BackOffTime(),
	}
}

func (resolver *serviceResolver) validateUserAddress(userAddress string) error {
	ctx, cancel := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancel()
	account, err := resolver.httpClientWrapper.GetAccount(ctx, userAddress)
	if err != nil {
		return err
	}

	if !hasBalance(account.Balance) {
		return fmt.Errorf("%w for account %s", ErrNoBalance, userAddress)
	}

	return nil
}

func (resolver *serviceResolver) verifyCode(userInfo *core.UserInfo, userAddress string, userIp, userCode string, guardianAddr []byte) (*requests.OTPCodeVerifyData, error) {
	verifyCodeData, isAllowed := resolver.frozenOtpHandler.IsVerificationAllowed(userAddress, userIp)
	if !isAllowed {
		return verifyCodeData, ErrTooManyFailedAttempts
	}
	otpHandler, err := resolver.getUserOTPHandler(userInfo, guardianAddr)
	if err != nil {
		return verifyCodeData, err
	}

	err = otpHandler.Validate(userCode)
	if err != nil {
		return verifyCodeData, err
	}

	resolver.frozenOtpHandler.Reset(userAddress, userIp)

	return &requests.OTPCodeVerifyData{
		RemainingTrials: int(resolver.frozenOtpHandler.MaxFailures()),
		ResetAfter:      0,
	}, nil
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
		return nil, ErrNilUserInfo
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

	userInfo, err := resolver.getUserInfo(addressBytes)
	if err == storage.ErrKeyNotFound {
		return resolver.handleNewAccount(userAddress, otp)
	}
	if err != nil {
		return nil, err
	}

	return resolver.handleRegisteredAccount(userAddress, userInfo, otp)
}

func (resolver *serviceResolver) validateTxRequestReturningGuardian(
	userIp, code string, txs []sdkData.Transaction,
) (core.GuardianInfo, *requests.OTPCodeVerifyData, error) {
	if len(txs) > resolver.config.MaxTransactionsAllowedForSigning {
		return core.GuardianInfo{}, requests.DefaultOTPCodeVerifyData(), fmt.Errorf("%w, got %d, max allowed %d",
			ErrTooManyTransactionsToSign, len(txs), resolver.config.MaxTransactionsAllowedForSigning)
	}

	if len(txs) == 0 {
		return core.GuardianInfo{}, requests.DefaultOTPCodeVerifyData(), ErrNoTransactionToSign
	}

	userAddress, err := sdkData.NewAddressFromBech32String(txs[0].SndAddr)
	if err != nil {
		return core.GuardianInfo{}, requests.DefaultOTPCodeVerifyData(), err
	}

	err = resolver.validateTransactions(txs, userAddress)
	if err != nil {
		return core.GuardianInfo{}, requests.DefaultOTPCodeVerifyData(), err
	}

	// only validate the guardian for first tx, as all of them must have the same one
	guardianAddr, err := resolver.pubKeyConverter.Decode(txs[0].GuardianAddr)
	if err != nil {
		return core.GuardianInfo{}, requests.DefaultOTPCodeVerifyData(), err
	}

	addressBytes := userAddress.AddressBytes()
	resolver.userCritSection.RLock(string(addressBytes))
	userInfo, err := resolver.getUserInfo(addressBytes)
	resolver.userCritSection.RUnlock(string(addressBytes))
	if err != nil {
		return core.GuardianInfo{}, requests.DefaultOTPCodeVerifyData(), err
	}

	otpVerifyCodeData, err := resolver.verifyCode(userInfo, txs[0].SndAddr, userIp, code, guardianAddr)
	if err != nil {
		return core.GuardianInfo{}, otpVerifyCodeData, err
	}

	// only get the guardian for first tx, as all of them must have the same one
	guardianInfo, err := resolver.getGuardianForTx(txs[0], userInfo)
	if err != nil {
		return core.GuardianInfo{}, otpVerifyCodeData, err
	}

	return guardianInfo, otpVerifyCodeData, nil
}

func (resolver *serviceResolver) updateGuardianStateIfNeeded(userAddress []byte, userInfo *core.UserInfo, guardianAddress []byte) error {
	userInfoCopy := *userInfo
	if bytes.Equal(guardianAddress, userInfoCopy.FirstGuardian.PublicKey) {
		if userInfoCopy.FirstGuardian.State == core.NotUsable {
			userInfoCopy.FirstGuardian.State = core.Usable
			return resolver.marshalAndSaveEncrypted(userAddress, &userInfoCopy)
		}
	}
	if bytes.Equal(guardianAddress, userInfoCopy.SecondGuardian.PublicKey) {
		if userInfoCopy.SecondGuardian.State == core.NotUsable {
			userInfoCopy.SecondGuardian.State = core.Usable
			return resolver.marshalAndSaveEncrypted(userAddress, &userInfoCopy)
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
		return fmt.Errorf("%w, initial sender: %s, current tx sender: %s", ErrInvalidSender, addr, tx.SndAddr)
	}

	userSig, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return err
	}

	userPublicKey, err := resolver.keyGen.PublicKeyFromByteArray(userAddress.AddressBytes())
	if err != nil {
		return err
	}

	if resolver.config.SkipTxUserSigVerify {
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
	err := resolver.validateUserAddress(userAddress.AddressAsBech32String())
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

	userInfo, err := resolver.computeNewUserDataAndSave(index, addressBytes, privateKeys, otp)
	if err != nil {
		return nil, err
	}

	log.Debug("registering new user",
		"userAddress", userAddress.AddressAsBech32String(),
		"guardian", resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey),
		"index", index)

	return userInfo.FirstGuardian.PublicKey, nil
}

func (resolver *serviceResolver) handleRegisteredAccount(userAddress sdkCore.AddressHandler, userInfo *core.UserInfo, otp handlers.OTP) ([]byte, error) {
	nextGuardian, err := resolver.getNextGuardianAddress(userAddress.AddressAsBech32String(), userInfo)
	if err != nil {
		return nil, err
	}

	err = resolver.saveOTPForUserGuardian(userAddress, userInfo, otp, nextGuardian)
	if err != nil {
		return nil, err
	}

	return nextGuardian, nil
}

func (resolver *serviceResolver) getNextGuardianAddress(userAddress string, userInfo *core.UserInfo) ([]byte, error) {
	if userInfo.FirstGuardian.State == core.NotUsable {
		log.Debug("registering old user",
			"userAddress", userAddress,
			"newGuardian", resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey))
		return userInfo.FirstGuardian.PublicKey, nil
	}

	if userInfo.SecondGuardian.State == core.NotUsable {
		log.Debug("registering old user",
			"userAddress", userAddress,
			"newGuardian", resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey))
		return userInfo.SecondGuardian.PublicKey, nil
	}

	ctxGetGuardianData, cancelGetGuardianData := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancelGetGuardianData()
	guardianData, err := resolver.httpClientWrapper.GetGuardianData(ctxGetGuardianData, userAddress)
	if err != nil {
		return nil, err
	}

	nextGuardian := resolver.prepareNextGuardian(guardianData, userInfo)

	printableGuardianData := ""
	guardianDataBuff, err := json.Marshal(guardianData)
	if err == nil {
		printableGuardianData = string(guardianDataBuff)
	}

	log.Debug("registering old user",
		"userAddress", userAddress,
		"newGuardian", resolver.pubKeyConverter.Encode(nextGuardian),
		"fetched data from chain", printableGuardianData)

	return nextGuardian, nil
}

func (resolver *serviceResolver) saveOTPForUserGuardian(userAddress sdkCore.AddressHandler, userInfo *core.UserInfo, otp handlers.OTP, guardian []byte) error {
	err := resolver.addOTPToUserGuardian(userInfo, guardian, otp)
	if err != nil {
		return err
	}

	addressBytes := userAddress.AddressBytes()
	return resolver.marshalAndSaveEncrypted(addressBytes, userInfo)
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
	nextAllowedOTPChangeTimestamp := oldOTPInfo.LastTOTPChangeTimestamp + int64(resolver.config.DelayBetweenOTPWritesInSec)
	allowedToChangeOTP := nextAllowedOTPChangeTimestamp < currentTimestamp
	if !allowedToChangeOTP {
		return fmt.Errorf("%w, last update was %d seconds ago, retry in %d seconds",
			handlers.ErrRegistrationFailed,
			currentTimestamp-oldOTPInfo.LastTOTPChangeTimestamp,
			nextAllowedOTPChangeTimestamp-currentTimestamp,
		)
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

func (resolver *serviceResolver) computeNewUserDataAndSave(index uint32, userAddress []byte, privateKeys []crypto.PrivateKey, otp handlers.OTP) (*core.UserInfo, error) {
	firstGuardian, err := getGuardianInfoForKey(privateKeys[0])
	if err != nil {
		return nil, err
	}

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
