package resolver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/txcheck"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/sync"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage"
)

var log = logger.GetOrCreate("serviceresolver")

const (
	minRequestTimeInSec       = 1
	zeroBalance               = "0"
	minDelayBetweenOTPUpdates = 1
	minTransactionsAllowed    = 1
	zeroQRAge                 = 0
	extendedStr               = "extended"
	notExtendedStr            = "not extended"

	maxConcurrentUpdateAttempts = 3
)

// ArgServiceResolver is the DTO used to create a new instance of service resolver
type ArgServiceResolver struct {
	UserEncryptor                 UserEncryptor
	TOTPHandler                   handlers.TOTPHandler
	SecureOtpHandler              handlers.SecureOtpHandler
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
	secureOtpHandler              handlers.SecureOtpHandler
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
		secureOtpHandler:              args.SecureOtpHandler,
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
	if check.IfNil(args.SecureOtpHandler) {
		return ErrNilSecureOtpHandler
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
func (resolver *serviceResolver) RegisterUser(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
	tag := resolver.extractUserTagForSecretGeneration(request.Tag, userAddress.Pretty())
	otp, err := resolver.totpHandler.CreateTOTP(tag)
	if err != nil {
		return &requests.OTP{}, "", err
	}

	otpUrl, err := otp.Url()
	if err != nil {
		return &requests.OTP{}, "", err
	}

	otpInfo, err := parseUrl(otpUrl)
	if err != nil {
		return &requests.OTP{}, "", err
	}

	guardianAddress, otpAge, err := resolver.registerUser(userAddress, otp)
	if err != nil {
		return &requests.OTP{
			TimeSinceGeneration: otpAge,
		}, "", err
	}

	encodedAddr, err := resolver.pubKeyConverter.Encode(guardianAddress)
	if err != nil {
		return &requests.OTP{
			TimeSinceGeneration: otpAge,
		}, "", err
	}

	return otpInfo, encodedAddr, nil
}

// VerifyCode validates the code received
func (resolver *serviceResolver) VerifyCode(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error) {
	guardianAddr, err := resolver.pubKeyConverter.Decode(request.Guardian)
	if err != nil {
		return nil, err
	}

	addressBytes := userAddress.AddressBytes()
	resolver.userCritSection.Lock(string(addressBytes))
	defer resolver.userCritSection.Unlock(string(addressBytes))

	userInfo, version, err := resolver.getUserInfoWithVersion(addressBytes)
	if err != nil {
		return nil, err
	}

	bech32Addr, err := userAddress.AddressAsBech32String()
	if err != nil {
		return nil, err
	}

	verifyCodeData, err := resolver.checkAllowanceAndVerifyCode(userInfo, bech32Addr, userIp, request.Code, request.SecondCode, guardianAddr)
	if err != nil {
		return verifyCodeData, err
	}

	err = resolver.updateGuardianStateIfNeeded(addressBytes, *userInfo, version, guardianAddr)
	if err != nil {
		return verifyCodeData, err
	}

	log.Debug("code ok",
		"userAddress", bech32Addr,
		"guardian", request.Guardian)

	return verifyCodeData, nil
}

// SignMessage validates user's message, then adds guardian signature and returns the message.
func (resolver *serviceResolver) SignMessage(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error) {
	userAddress, err := sdkData.NewAddressFromBech32String(request.UserAddr)
	if err != nil {
		return nil, nil, err
	}
	guardian, otpCodeVerifyData, err := resolver.verifyCodesReturningGuardian(userAddress, request.GuardianAddr,
		userIp, request.Code, request.SecondCode)
	if err != nil {
		return nil, otpCodeVerifyData, err
	}

	guardianCryptoHolder, err := resolver.cryptoComponentsHolderFactory.Create(guardian.PrivateKey)
	if err != nil {
		return nil, otpCodeVerifyData, err
	}

	signedMessage, err := resolver.signatureVerifier.SignMessage([]byte(request.Message), guardianCryptoHolder.GetPrivateKey())
	if err != nil {
		return nil, otpCodeVerifyData, err
	}

	return signedMessage, otpCodeVerifyData, err

}

// SetSecurityModeNoExpire gets the user's guardian, verifies the codes and then sets the SecurityMode
func (resolver *serviceResolver) SetSecurityModeNoExpire(userIp string, request requests.SecurityModeNoExpire) (*requests.OTPCodeVerifyData, error) {
	userAddress, verifyCodeData, err := resolver.checkGuardianAndVerifyCode(userIp, request)
	if err != nil {
		return verifyCodeData, err
	}
	return verifyCodeData, resolver.secureOtpHandler.SetSecurityModeNoExpire(userAddress)
}

// UnsetSecurityModeNoExpire gets the user's guardian, verifies the codes and then unsets the SecurityMode
func (resolver *serviceResolver) UnsetSecurityModeNoExpire(userIp string, request requests.SecurityModeNoExpire) (*requests.OTPCodeVerifyData, error) {
	userAddress, verifyCodeData, err := resolver.checkGuardianAndVerifyCode(userIp, request)
	if err != nil {
		return verifyCodeData, err
	}
	return verifyCodeData, resolver.secureOtpHandler.UnsetSecurityModeNoExpire(userAddress)
}

// SignTransaction validates user's transaction, then adds guardian signature and returns the transaction
func (resolver *serviceResolver) SignTransaction(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error) {
	guardian, otpCodeVerifyData, err := resolver.validateTxRequestReturningGuardian(userIp, request.Code, request.SecondCode, []transaction.FrontendTransaction{request.Tx})
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
	guardian, otpCodeVerifyData, err := resolver.validateTxRequestReturningGuardian(userIp, request.Code, request.SecondCode, request.Txs)
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
		BackoffWrongCode: resolver.secureOtpHandler.FreezeBackOffTime(),
	}
}

// checkGuardianAndVerifyCode returns the user address in bech32 format
func (resolver *serviceResolver) checkGuardianAndVerifyCode(userIp string, request requests.SecurityModeNoExpire) (string, *requests.OTPCodeVerifyData, error) {
	userAddress, err := sdkData.NewAddressFromBech32String(request.UserAddr)
	if err != nil {
		return "", nil, err
	}

	bech32Addr, err := userAddress.AddressAsBech32String()
	if err != nil {
		return "", nil, err
	}

	ctxGetGuardianData, cancelGetGuardianData := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancelGetGuardianData()
	guardianData, err := resolver.httpClientWrapper.GetGuardianData(ctxGetGuardianData, bech32Addr)
	if err != nil {
		return "", nil, err
	}

	if check.IfNilReflect(guardianData.ActiveGuardian) {
		return "", nil, ErrAccountHasNoActiveGuardian
	}

	guardianAddrBytes, err := resolver.pubKeyConverter.Decode(guardianData.ActiveGuardian.Address)
	if err != nil {
		return "", nil, err
	}

	addressBytes := userAddress.AddressBytes()
	resolver.userCritSection.RLock(string(addressBytes))
	userInfo, err := resolver.getUserInfo(addressBytes)
	resolver.userCritSection.RUnlock(string(addressBytes))
	if err != nil {
		return "", nil, err
	}

	verifyCodeData, err := resolver.checkAllowanceAndVerifyCode(userInfo, bech32Addr, userIp, request.Code, request.SecondCode, guardianAddrBytes)
	return bech32Addr, verifyCodeData, err
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

// registerUser registers the user, retrying when the user data was changed concurrently by another instance
func (resolver *serviceResolver) registerUser(userAddress sdkCore.AddressHandler, otp handlers.OTP) ([]byte, int64, error) {
	addressBytes := userAddress.AddressBytes()

	resolver.userCritSection.Lock(string(addressBytes))
	defer resolver.userCritSection.Unlock(string(addressBytes))

	for attempt := 0; attempt < maxConcurrentUpdateAttempts; attempt++ {
		guardian, otpAge, err := resolver.tryRegisterUser(userAddress, otp)
		if !isConcurrentWriteError(err) {
			return guardian, otpAge, err
		}

		log.Debug("user data changed concurrently, retrying registration", "attempt", attempt)
	}

	return nil, zeroQRAge, ErrConcurrentUpdate
}

// tryRegisterUser reads the user once and saves it only if nobody changed it meanwhile
func (resolver *serviceResolver) tryRegisterUser(userAddress sdkCore.AddressHandler, otp handlers.OTP) ([]byte, int64, error) {
	userInfo, version, err := resolver.getUserInfoWithVersion(userAddress.AddressBytes())
	if errors.Is(err, storage.ErrKeyNotFound) {
		guardian, errNewAccount := resolver.handleNewAccount(userAddress, otp)
		return guardian, zeroQRAge, errNewAccount
	}
	if err != nil {
		return nil, zeroQRAge, err
	}

	return resolver.handleRegisteredAccount(userAddress, *userInfo, version, otp)
}

func isConcurrentWriteError(err error) bool {
	return errors.Is(err, storage.ErrKeyAlreadyExists) || errors.Is(err, storage.ErrVersionMismatch)
}

func (resolver *serviceResolver) validateTxRequestReturningGuardian(
	userIp, code string, secondCode string, txs []transaction.FrontendTransaction,
) (core.GuardianInfo, *requests.OTPCodeVerifyData, error) {
	if len(txs) > resolver.config.MaxTransactionsAllowedForSigning {
		return core.GuardianInfo{}, nil, fmt.Errorf("%w, got %d, max allowed %d",
			ErrTooManyTransactionsToSign, len(txs), resolver.config.MaxTransactionsAllowedForSigning)
	}

	if len(txs) == 0 {
		return core.GuardianInfo{}, nil, ErrNoTransactionToSign
	}

	userAddress, err := sdkData.NewAddressFromBech32String(txs[0].Sender)
	if err != nil {
		return core.GuardianInfo{}, nil, err
	}

	err = resolver.validateTransactions(txs, userAddress)
	if err != nil {
		return core.GuardianInfo{}, nil, err
	}

	return resolver.verifyCodesReturningGuardian(userAddress, txs[0].GuardianAddr, userIp, code, secondCode)
}

func (resolver *serviceResolver) verifyCodesReturningGuardian(
	userAddress sdkCore.AddressHandler,
	guardianAddr string,
	userIp,
	code,
	secondCode string,
) (core.GuardianInfo, *requests.OTPCodeVerifyData, error) {
	guardianAddrBytes, err := resolver.pubKeyConverter.Decode(guardianAddr)
	if err != nil {
		return core.GuardianInfo{}, nil, err
	}

	addressBytes := userAddress.AddressBytes()
	resolver.userCritSection.RLock(string(addressBytes))
	userInfo, err := resolver.getUserInfo(addressBytes)
	resolver.userCritSection.RUnlock(string(addressBytes))

	if err != nil {
		return core.GuardianInfo{}, nil, err
	}

	bech32Addr, err := userAddress.AddressAsBech32String()
	if err != nil {
		return core.GuardianInfo{}, nil, err
	}

	otpVerifyCodeData, err := resolver.checkAllowanceAndVerifyCode(
		userInfo,
		bech32Addr,
		userIp,
		code,
		secondCode,
		guardianAddrBytes,
	)
	if err != nil {
		return core.GuardianInfo{}, otpVerifyCodeData, err
	}

	guardianInfo, err := resolver.getGuardianInfoFromAddress(guardianAddr, userInfo)
	if err != nil {
		return core.GuardianInfo{}, otpVerifyCodeData, err
	}

	return guardianInfo, otpVerifyCodeData, nil
}

func (resolver *serviceResolver) extendSecurityMode(verifyCodeData *requests.OTPCodeVerifyData, userAddress string) {
	errExtendSecurityMode := resolver.secureOtpHandler.ExtendSecurityMode(userAddress)
	if errExtendSecurityMode == nil && verifyCodeData != nil && verifyCodeData.SecurityModeResetAfter != core.NoExpiryValue {
		verifyCodeData.SecurityModeResetAfter = int(resolver.secureOtpHandler.SecurityModeBackOffTime())
	}
}

func (resolver *serviceResolver) checkAllowanceAndVerifyCode(
	userInfo *core.UserInfo,
	userAddress string,
	userIp string,
	code string,
	secondCode string,
	guardianAddr []byte,
) (*requests.OTPCodeVerifyData, error) {
	verifyCodeData, err := resolver.secureOtpHandler.IsVerificationAllowedAndIncreaseTrials(userAddress, userIp)
	if err != nil {
		resolver.extendSecurityMode(verifyCodeData, userAddress)
		return verifyCodeData, err
	}

	err = resolver.verifyCode(userInfo, code, guardianAddr)
	if err != nil {
		resolver.extendSecurityMode(verifyCodeData, userAddress)
		return verifyCodeData, err
	}

	securityModeExtended, err := resolver.verifySecurityModeCode(
		userInfo,
		userAddress,
		code,
		secondCode,
		guardianAddr,
		verifyCodeData.SecurityModeRemainingTrials)
	if err != nil {
		log.Debug("second code verification failed", "user", userAddress, "error", err.Error())

		if securityModeExtended && verifyCodeData.SecurityModeResetAfter != core.NoExpiryValue {
			verifyCodeData.SecurityModeResetAfter = int(resolver.secureOtpHandler.SecurityModeBackOffTime())
		}
		return verifyCodeData, ErrWrongCode
	}

	err = resolver.markCodesAsUsed(guardianAddr, code, secondCode, verifyCodeData.SecurityModeRemainingTrials)
	if errors.Is(err, core.ErrCodeAlreadyUsed) {
		resolver.extendSecurityMode(verifyCodeData, userAddress)
		return verifyCodeData, ErrWrongCode
	}
	if err != nil {
		return verifyCodeData, err
	}

	errDec := resolver.secureOtpHandler.DecrementSecurityModeFailedTrials(userAddress)
	if errDec != nil {
		log.Warn("failed to decrement security mode failed trials", "user", userAddress, "error", errDec.Error())
	}

	resolver.secureOtpHandler.Reset(userAddress, userIp)

	return &requests.OTPCodeVerifyData{
		RemainingTrials:             int(resolver.secureOtpHandler.FreezeMaxFailures()),
		ResetAfter:                  0,
		SecurityModeRemainingTrials: verifyCodeData.SecurityModeRemainingTrials,
		SecurityModeResetAfter:      verifyCodeData.SecurityModeResetAfter,
	}, nil
}

func (resolver *serviceResolver) verifySecurityModeCode(
	userInfo *core.UserInfo,
	userAddress string,
	firstCode string,
	secondCode string,
	guardianAddr []byte,
	securityModeRemainingTrials int,
) (bool, error) {
	if securityModeRemainingTrials <= 0 {
		if secondCode == firstCode {
			errExtendSecurityMode := resolver.secureOtpHandler.ExtendSecurityMode(userAddress)
			securityModeExtended := errExtendSecurityMode == nil

			securityModeExtendedStr := extendedStr
			if !securityModeExtended {
				securityModeExtendedStr = notExtendedStr
			}

			if errExtendSecurityMode != nil {
				log.Error("failed to extend security mode", "error", errExtendSecurityMode)
			}

			return securityModeExtended, fmt.Errorf("%w with codeError %s, security mode %s", ErrSecondCodeInvalidInSecurityMode, ErrSameCode, securityModeExtendedStr)
		}

		err := resolver.verifyCode(userInfo, secondCode, guardianAddr)
		if err != nil {
			// if the second code is not correct, extend the ttl for security mode
			errExtendSecurityMode := resolver.secureOtpHandler.ExtendSecurityMode(userAddress)
			securityModeExtended := errExtendSecurityMode == nil

			securityModeExtendedStr := extendedStr
			if !securityModeExtended {
				securityModeExtendedStr = notExtendedStr
			}

			return securityModeExtended, fmt.Errorf("%w with codeError %s, security mode %s", ErrSecondCodeInvalidInSecurityMode, err, securityModeExtendedStr)
		}
	}

	return false, nil
}

func (resolver *serviceResolver) markCodesAsUsed(guardianAddr []byte, code string, secondCode string, securityModeRemainingTrials int) error {
	codes := []string{code}
	if securityModeRemainingTrials <= 0 {
		codes = append(codes, secondCode)
	}

	return resolver.secureOtpHandler.MarkCodesAsUsed(hex.EncodeToString(guardianAddr), codes)
}

// updateGuardianStateIfNeeded marks the verified guardian as usable, only if its otp did not change meanwhile
func (resolver *serviceResolver) updateGuardianStateIfNeeded(userAddress []byte, userInfo core.UserInfo, version int64, guardianAddress []byte) error {
	verifiedGuardian, found := getGuardianInfo(userInfo, guardianAddress)
	if !found {
		return nil
	}

	for attempt := 0; attempt < maxConcurrentUpdateAttempts; attempt++ {
		guardianInfo, found := getGuardianInfo(userInfo, guardianAddress)
		if !found {
			return nil
		}
		if !isSameOTPInfo(guardianInfo.OTPData, verifiedGuardian.OTPData) {
			return ErrConcurrentUpdate
		}
		if guardianInfo.State != core.NotUsable {
			return nil
		}

		guardianInfo.State = core.Usable
		updatedUserInfo := withGuardianInfo(userInfo, guardianInfo)
		err := resolver.marshalAndSaveEncryptedIfVersion(userAddress, &updatedUserInfo, version)
		if !errors.Is(err, storage.ErrVersionMismatch) {
			return err
		}

		log.Debug("guardian state updated concurrently, retrying", "attempt", attempt)
		freshUserInfo, freshVersion, err := resolver.getUserInfoWithVersion(userAddress)
		if err != nil {
			return err
		}
		userInfo, version = *freshUserInfo, freshVersion
	}

	return ErrConcurrentUpdate
}

// getGuardianInfo returns a copy of the user's guardian with the provided public key
func getGuardianInfo(userInfo core.UserInfo, guardianAddress []byte) (core.GuardianInfo, bool) {
	switch {
	case bytes.Equal(guardianAddress, userInfo.FirstGuardian.PublicKey):
		return userInfo.FirstGuardian, true
	case bytes.Equal(guardianAddress, userInfo.SecondGuardian.PublicKey):
		return userInfo.SecondGuardian, true
	default:
		return core.GuardianInfo{}, false
	}
}

// withGuardianInfo returns a copy of the user info with the guardian of the same public key replaced
func withGuardianInfo(userInfo core.UserInfo, guardianInfo core.GuardianInfo) core.UserInfo {
	switch {
	case bytes.Equal(guardianInfo.PublicKey, userInfo.FirstGuardian.PublicKey):
		userInfo.FirstGuardian = guardianInfo
	case bytes.Equal(guardianInfo.PublicKey, userInfo.SecondGuardian.PublicKey):
		userInfo.SecondGuardian = guardianInfo
	}

	return userInfo
}

func isSameOTPInfo(first core.OTPInfo, second core.OTPInfo) bool {
	return bytes.Equal(first.OTP, second.OTP) && first.LastTOTPChangeTimestamp == second.LastTOTPChangeTimestamp
}

func (resolver *serviceResolver) validateTransactions(txs []transaction.FrontendTransaction, userAddress sdkCore.AddressHandler) error {
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

func (resolver *serviceResolver) validateOneTransaction(tx transaction.FrontendTransaction, userAddress sdkCore.AddressHandler) error {
	addr, err := userAddress.AddressAsBech32String()
	if err != nil {
		return err
	}
	if tx.Sender != addr {
		return fmt.Errorf("%w, initial sender: %s, current tx sender: %s", ErrInvalidSender, addr, tx.Sender)
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

func (resolver *serviceResolver) getGuardianInfoFromAddress(guardianAddr string, userInfo *core.UserInfo) (core.GuardianInfo, error) {
	guardianForTx := core.GuardianInfo{}
	unknownGuardian := true
	firstGuardianAddr, err := resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey)
	if err != nil {
		return core.GuardianInfo{}, err
	}
	if guardianAddr == firstGuardianAddr {
		guardianForTx = userInfo.FirstGuardian
		unknownGuardian = false
	}
	secondGuardianAddr, err := resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey)
	if err != nil {
		return core.GuardianInfo{}, err
	}
	if guardianAddr == secondGuardianAddr {
		guardianForTx = userInfo.SecondGuardian
		unknownGuardian = false
	}

	if unknownGuardian {
		return core.GuardianInfo{}, fmt.Errorf("%w, guardian %s", ErrInvalidGuardian, guardianAddr)
	}

	if guardianForTx.State == core.NotUsable {
		return core.GuardianInfo{}, fmt.Errorf("%w, guardian %s", ErrGuardianNotUsable, guardianAddr)
	}

	return guardianForTx, nil
}

func (resolver *serviceResolver) handleNewAccount(userAddress sdkCore.AddressHandler, otp handlers.OTP) ([]byte, error) {
	bech32Addr, err := userAddress.AddressAsBech32String()
	if err != nil {
		return nil, err
	}
	err = resolver.validateUserAddress(bech32Addr)
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
		"userAddress", bech32Addr,
		"guardian", resolver.pubKeyConverter.SilentEncode(userInfo.FirstGuardian.PublicKey, log),
		"index", index)

	return userInfo.FirstGuardian.PublicKey, nil
}

func (resolver *serviceResolver) handleRegisteredAccount(userAddress sdkCore.AddressHandler, userInfo core.UserInfo, version int64, otp handlers.OTP) ([]byte, int64, error) {
	bech32Addr, err := userAddress.AddressAsBech32String()
	if err != nil {
		return nil, zeroQRAge, err
	}

	nextGuardian, userInfo, err := resolver.getNextGuardianAddress(bech32Addr, userInfo)
	if err != nil {
		return nil, zeroQRAge, err
	}

	userInfo, otpAge, err := resolver.addOTPToUserGuardian(userInfo, nextGuardian, otp)
	if err != nil {
		return nil, otpAge, err
	}

	err = resolver.marshalAndSaveEncryptedIfVersion(userAddress.AddressBytes(), &userInfo, version)
	if err != nil {
		return nil, otpAge, err
	}

	return nextGuardian, otpAge, nil
}

// getNextGuardianAddress returns the guardian to register next and the user info with the updated guardian states
func (resolver *serviceResolver) getNextGuardianAddress(userAddress string, userInfo core.UserInfo) ([]byte, core.UserInfo, error) {
	if userInfo.FirstGuardian.State == core.NotUsable {
		log.Debug("registering old user",
			"userAddress", userAddress,
			"newGuardian", resolver.pubKeyConverter.SilentEncode(userInfo.FirstGuardian.PublicKey, log))
		return userInfo.FirstGuardian.PublicKey, userInfo, nil
	}

	if userInfo.SecondGuardian.State == core.NotUsable {
		log.Debug("registering old user",
			"userAddress", userAddress,
			"newGuardian", resolver.pubKeyConverter.SilentEncode(userInfo.SecondGuardian.PublicKey, log))
		return userInfo.SecondGuardian.PublicKey, userInfo, nil
	}

	ctxGetGuardianData, cancelGetGuardianData := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancelGetGuardianData()
	guardianData, err := resolver.httpClientWrapper.GetGuardianData(ctxGetGuardianData, userAddress)
	if err != nil {
		return nil, userInfo, err
	}

	nextGuardian, userInfo := resolver.prepareNextGuardian(guardianData, userInfo)

	printableGuardianData := ""
	guardianDataBuff, err := json.Marshal(guardianData)
	if err == nil {
		printableGuardianData = string(guardianDataBuff)
	}

	log.Debug("registering old user",
		"userAddress", userAddress,
		"newGuardian", resolver.pubKeyConverter.SilentEncode(nextGuardian, log),
		"fetched data from chain", printableGuardianData)

	return nextGuardian, userInfo, nil
}

// addOTPToUserGuardian returns the user info with the new otp set for the guardian, if the otp change is allowed
func (resolver *serviceResolver) addOTPToUserGuardian(userInfo core.UserInfo, guardian []byte, otp handlers.OTP) (core.UserInfo, int64, error) {
	guardianInfo, found := getGuardianInfo(userInfo, guardian)
	if !found {
		return userInfo, zeroQRAge, ErrInvalidGuardian
	}

	currentTimestamp := time.Now().Unix()
	lastChangeTimestamp := guardianInfo.OTPData.LastTOTPChangeTimestamp
	otpAge := currentTimestamp - lastChangeTimestamp
	nextAllowedOTPChangeTimestamp := lastChangeTimestamp + int64(resolver.config.DelayBetweenOTPWritesInSec)
	allowedToChangeOTP := nextAllowedOTPChangeTimestamp < currentTimestamp
	if !allowedToChangeOTP {
		return userInfo, otpAge, fmt.Errorf("%w, last update was %d seconds ago, retry in %d seconds",
			handlers.ErrRegistrationFailed,
			otpAge,
			nextAllowedOTPChangeTimestamp-currentTimestamp,
		)
	}

	otpBytes, err := otp.ToBytes()
	if err != nil {
		return userInfo, zeroQRAge, err
	}

	guardianInfo.OTPData = core.OTPInfo{
		OTP:                     otpBytes,
		LastTOTPChangeTimestamp: currentTimestamp,
	}

	return withGuardianInfo(userInfo, guardianInfo), zeroQRAge, nil
}

func (resolver *serviceResolver) getUserInfo(userAddress []byte) (*core.UserInfo, error) {
	encryptedDataMarshalled, err := resolver.registeredUsersDB.Get(userAddress)
	if err != nil {
		return nil, err
	}

	return resolver.unmarshalAndDecryptUserInfo(encryptedDataMarshalled)
}

// getUserInfoWithVersion reads from the primary, to be used before conditional writes
func (resolver *serviceResolver) getUserInfoWithVersion(userAddress []byte) (*core.UserInfo, int64, error) {
	encryptedDataMarshalled, version, err := resolver.registeredUsersDB.GetWithVersion(userAddress)
	if err != nil {
		return nil, 0, err
	}

	userInfo, err := resolver.unmarshalAndDecryptUserInfo(encryptedDataMarshalled)
	if err != nil {
		return nil, 0, err
	}

	return userInfo, version, nil
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

func (resolver *serviceResolver) computeNewUserDataAndSave(index uint32, userAddress []byte, privateKeys []crypto.PrivateKey, otp handlers.OTP) (core.UserInfo, error) {
	firstGuardian, err := getGuardianInfoForKey(privateKeys[0])
	if err != nil {
		return core.UserInfo{}, err
	}

	firstGuardian.OTPData.OTP, err = otp.ToBytes()
	if err != nil {
		return core.UserInfo{}, err
	}
	firstGuardian.OTPData.LastTOTPChangeTimestamp = time.Now().Unix()

	secondGuardian, err := getGuardianInfoForKey(privateKeys[1])
	if err != nil {
		return core.UserInfo{}, err
	}

	userInfo := core.UserInfo{
		Index:          index,
		FirstGuardian:  firstGuardian,
		SecondGuardian: secondGuardian,
	}

	err = resolver.marshalAndInsertEncrypted(userAddress, &userInfo)
	if err != nil {
		return core.UserInfo{}, err
	}

	return userInfo, nil
}

func (resolver *serviceResolver) marshalAndInsertEncrypted(userAddress []byte, userInfo *core.UserInfo) error {
	encryptedDataBytes, err := resolver.encryptAndMarshalUserInfo(userInfo)
	if err != nil {
		return err
	}

	return resolver.registeredUsersDB.Insert(userAddress, encryptedDataBytes)
}

func (resolver *serviceResolver) marshalAndSaveEncryptedIfVersion(userAddress []byte, userInfo *core.UserInfo, version int64) error {
	encryptedDataBytes, err := resolver.encryptAndMarshalUserInfo(userInfo)
	if err != nil {
		return err
	}

	return resolver.registeredUsersDB.PutIfVersion(userAddress, encryptedDataBytes, version)
}

func (resolver *serviceResolver) prepareNextGuardian(guardianData *api.GuardianData, userInfo core.UserInfo) ([]byte, core.UserInfo) {
	firstGuardianOnChainState := resolver.getOnChainGuardianState(guardianData, userInfo.FirstGuardian)
	secondGuardianOnChainState := resolver.getOnChainGuardianState(guardianData, userInfo.SecondGuardian)
	isFirstOnChain := firstGuardianOnChainState != core.MissingGuardian
	isSecondOnChain := secondGuardianOnChainState != core.MissingGuardian
	if !isFirstOnChain && !isSecondOnChain {
		userInfo.FirstGuardian.State = core.NotUsable
		userInfo.SecondGuardian.State = core.NotUsable
		return userInfo.FirstGuardian.PublicKey, userInfo
	}

	if isFirstOnChain && isSecondOnChain {
		if firstGuardianOnChainState == core.PendingGuardian {
			userInfo.FirstGuardian.State = core.NotUsable
			return userInfo.FirstGuardian.PublicKey, userInfo
		}

		userInfo.SecondGuardian.State = core.NotUsable
		return userInfo.SecondGuardian.PublicKey, userInfo
	}

	if isFirstOnChain {
		userInfo.SecondGuardian.State = core.NotUsable
		return userInfo.SecondGuardian.PublicKey, userInfo
	}

	userInfo.FirstGuardian.State = core.NotUsable
	return userInfo.FirstGuardian.PublicKey, userInfo
}

func (resolver *serviceResolver) getOnChainGuardianState(guardianData *api.GuardianData, guardian core.GuardianInfo) core.OnChainGuardianState {
	if check.IfNilReflect(guardianData) {
		return core.MissingGuardian
	}

	guardianAddress, err := resolver.pubKeyConverter.Encode(guardian.PublicKey)
	if err != nil {
		return core.MissingGuardian
	}
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

func (resolver *serviceResolver) extractUserTagForSecretGeneration(tag string, prettyUserAddress string) string {
	if len(tag) > 0 {
		return tag
	}
	return prettyUserAddress
}

func parseUrl(otpUrl string) (*requests.OTP, error) {
	if len(otpUrl) == 0 {
		return &requests.OTP{}, ErrEmptyUrl
	}

	// example of valid url: otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example
	u, err := url.Parse(otpUrl)
	if err != nil {
		log.Warn("could not parse url")
		return &requests.OTP{}, fmt.Errorf("%w while parsing otp url", err)
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	otpInfo := &requests.OTP{}

	query := u.Query()
	err = decoder.Decode(otpInfo, query)
	if err != nil {
		log.Warn("could not extract schema from url")
		return &requests.OTP{}, fmt.Errorf("%w while extracting schema from url", err)
	}

	account, err := extractAccount(u.Path)
	if err != nil {
		log.Warn("could not parse path", "path", u.Path)
		return &requests.OTP{}, fmt.Errorf("%w while extracting account from path", err)
	}

	otpInfo.Scheme = u.Scheme
	otpInfo.Host = u.Host
	otpInfo.Account = account

	return otpInfo, nil
}

func extractAccount(path string) (string, error) {
	// path should be /issuer:account
	pathParts := strings.Split(path, ":")
	if len(pathParts) != 2 {
		return "", fmt.Errorf("%w while parsing path", ErrInvalidValue)
	}

	return pathParts[1], nil
}

// IsInterfaceNil return true if there is no value under the interface
func (resolver *serviceResolver) IsInterfaceNil() bool {
	return resolver == nil
}
