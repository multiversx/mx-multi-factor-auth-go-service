package resolver

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	erdData "github.com/ElrondNetwork/elrond-sdk-erdgo/data"
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
	requestTime        time.Duration
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
		requestTime:        args.RequestTime,
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
	if args.RequestTime < minRequestTime {
		return fmt.Errorf("%w for RequestTime, received %d, min expected %d", ErrInvalidValue, args.RequestTime, minRequestTime)
	}

	return nil
}

// GetGuardianAddress returns the address of a unique guardian
func (resolver *serviceResolver) GetGuardianAddress(request requests.GetGuardianAddress) (string, error) {
	userAddress, err := resolver.validateCredentials(request.Credentials)
	if err != nil {
		return emptyAddress, err
	}

	addressBytes := userAddress.AddressBytes()
	err = resolver.registeredUsersDB.Has(addressBytes)
	if err != nil {
		return resolver.handleNewAccount(addressBytes)

	}

	return resolver.handleRegisteredAccount(addressBytes)
}

// RegisterUser creates a new OTP for the given provider
// and (optionally) returns some information required for the user to set up the OTP on his end (eg: QR code).
func (resolver *serviceResolver) RegisterUser(request requests.RegistrationPayload) ([]byte, error) {
	userAddress, err := resolver.validateCredentials(request.Credentials)
	if err != nil {
		return nil, err
	}

	err = resolver.validateGuardian(userAddress.AddressBytes(), request.Guardian)
	if err != nil {
		return nil, fmt.Errorf("%w for guardian %s", err, request.Guardian)
	}

	return resolver.provider.RegisterUser(userAddress.AddressAsBech32String(), request.Guardian)
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

func (resolver *serviceResolver) validateGuardian(userAddress []byte, guardian string) error {
	userInfo, err := resolver.getUserInfo(userAddress)
	if err != nil {
		return err
	}

	firstAddress := resolver.pubKeyConverter.Encode(userInfo.FirstGuardian.PublicKey)
	isNotUsableYet := userInfo.FirstGuardian.State == core.NotUsableYet
	if isNotUsableYet && guardian == firstAddress {
		return nil
	}

	secondAddress := resolver.pubKeyConverter.Encode(userInfo.SecondGuardian.PublicKey)
	isNotUsableYet = userInfo.SecondGuardian.State == core.NotUsableYet
	if isNotUsableYet && guardian == secondAddress {
		return nil
	}

	return ErrInvalidGuardian
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
	index := resolver.indexHandler.AllocateIndex()
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
	data, err := resolver.registeredUsersDB.Get(userAddress)
	if err != nil {
		return userInfo, err
	}

	err = resolver.marshaller.Unmarshal(&userInfo, data)
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
	data, err := resolver.marshaller.Marshal(userInfo)
	if err != nil {
		return err
	}

	err = resolver.registeredUsersDB.Put(userAddress, data)
	if err != nil {
		return err
	}

	return nil
}

func (resolver *serviceResolver) prepareNextGuardian(guardianData *erdData.GuardianData, userInfo *core.UserInfo) string {
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

func (resolver *serviceResolver) getOnChainGuardianState(guardianData *erdData.GuardianData, guardian core.GuardianInfo) core.OnChainGuardianState {
	guardianAddress := resolver.pubKeyConverter.Encode(guardian.PublicKey)
	isActiveGuardian := guardianData.ActiveGuardian.Address == guardianAddress
	if isActiveGuardian {
		return core.ActiveGuardian
	}

	isPendingGuardian := guardianData.PendingGuardian.Address == guardianAddress
	if isPendingGuardian {
		return core.PendingGuardian
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

// IsInterfaceNil return true if there is no value under the interface
func (resolver *serviceResolver) IsInterfaceNil() bool {
	return resolver == nil
}
