package resolver

import (
	"bytes"
	"context"
	"fmt"
	"time"

	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	erdData "github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

const emptyAddress = ""

// ArgServiceResolver is the DTO used to create a new instance of service resolver
type ArgServiceResolver struct {
	Proxy              blockchain.Proxy
	CredentialsHandler core.CredentialsHandler
	IndexHandler       core.IndexHandler
	KeysGenerator      core.KeysGenerator
	PubKeyConverter    core.PubkeyConverter
	RegisteredUsersDB  core.Storer
	ProvidersMap       map[string]core.Provider
	Marshaller         core.Marshaller
	RequestTime        time.Duration
}

// TODO add unittests on this structure
type serviceResolver struct {
	proxy              blockchain.Proxy
	credentialsHandler core.CredentialsHandler
	indexHandler       core.IndexHandler
	keysGenerator      core.KeysGenerator
	pubKeyConverter    core.PubkeyConverter
	registeredUsersDB  core.Storer
	providersMap       map[string]core.Provider
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
		proxy:              args.Proxy,
		credentialsHandler: args.CredentialsHandler,
		indexHandler:       args.IndexHandler,
		keysGenerator:      args.KeysGenerator,
		pubKeyConverter:    args.PubKeyConverter,
		registeredUsersDB:  args.RegisteredUsersDB,
		providersMap:       args.ProvidersMap,
		marshaller:         args.Marshaller,
		requestTime:        args.RequestTime,
	}, nil
}

func checkArgs(_ ArgServiceResolver) error {
	// TODO implement this

	return nil
}

// GetGuardianAddress returns the address of a unique guardian
func (resolver *serviceResolver) GetGuardianAddress(request requests.GetGuardianAddress) (string, error) {
	userAddress, err := resolver.validateCredentials(request.Credentials)
	if err != nil {
		return emptyAddress, err
	}

	_, ok := resolver.providersMap[request.Provider]
	if !ok {
		return emptyAddress, fmt.Errorf("%w, provider %s", ErrProviderDoesNotExists, request.Provider)
	}

	addressBytes := userAddress.AddressBytes()
	isRegistered := resolver.registeredUsersDB.Has(addressBytes)
	if isRegistered {
		return resolver.handleRegisteredAccount(addressBytes)
	}

	return resolver.handleNewAccount(addressBytes, request.Provider)
}

// RegisterUser creates a new OTP for the given provider
// and (optionally) returns some information required for the user to set up the OTP on his end (eg: QR code).
func (resolver *serviceResolver) RegisterUser(request requests.Register) ([]byte, error) {
	userAddress, err := resolver.validateCredentials(request.Credentials)
	if err != nil {
		return nil, err
	}

	err = resolver.validateGuardian(userAddress.AddressBytes(), request.Guardian)
	if err != nil {
		return nil, fmt.Errorf("%w for guardian %s", err, request.Guardian)
	}

	provider, ok := resolver.providersMap[request.Provider]
	if !ok {
		return nil, fmt.Errorf("%w, provider %s", ErrProviderDoesNotExists, request.Provider)
	}

	return provider.RegisterUser(userAddress.AddressAsBech32String())
}

// VerifyCodes validates the code received
func (resolver *serviceResolver) VerifyCodes(request requests.VerifyCodes) error {
	if len(request.Codes) == 0 {
		return ErrEmptyCodesSlice
	}

	userAddress, err := resolver.validateCredentials(request.Credentials)
	if err != nil {
		return err
	}

	for _, code := range request.Codes {
		provider, ok := resolver.providersMap[code.Provider]
		if !ok {
			return fmt.Errorf("%w, provider %s", ErrProviderDoesNotExists, code.Provider)
		}

		err = provider.VerifyCode(userAddress.AddressAsBech32String(), code.Code)
		if err != nil {
			return err
		}
	}

	guardianAddrBuff, err := resolver.pubKeyConverter.Decode(request.Guardian)
	if err != nil {
		return err
	}

	return resolver.updateGuardianState(userAddress.AddressBytes(), guardianAddrBuff)
}

func (resolver *serviceResolver) updateGuardianState(userAddress []byte, guardianAddress []byte) error {
	userInfo, err := resolver.getUserInfo(userAddress)
	if err != nil {
		return err
	}

	if bytes.Equal(guardianAddress, userInfo.MainGuardian.PublicKey) {
		userInfo.MainGuardian.State = core.Usable
	}
	if bytes.Equal(guardianAddress, userInfo.SecondaryGuardian.PublicKey) {
		userInfo.SecondaryGuardian.State = core.Usable
	}

	err = resolver.saveUserInfo(userAddress, userInfo)
	if err != nil {
		return err
	}

	return nil
}

func (resolver *serviceResolver) validateGuardian(userAddress []byte, guardian string) error {
	userInfo, err := resolver.getUserInfo(userAddress)
	if err != nil {
		return err
	}

	mainAddress := resolver.pubKeyConverter.Encode(userInfo.MainGuardian.PublicKey)
	if guardian == mainAddress {
		return nil
	}

	secondaryAddress := resolver.pubKeyConverter.Encode(userInfo.SecondaryGuardian.PublicKey)
	if guardian == secondaryAddress {
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

func (resolver *serviceResolver) handleNewAccount(userAddress []byte, provider string) (string, error) {
	index := resolver.indexHandler.GetIndex()
	privateKeys, err := resolver.keysGenerator.GenerateKeys(index)
	if err != nil {
		return emptyAddress, err
	}

	userInfo, err := resolver.computeDataAndSave(index, userAddress, privateKeys, provider)
	if err != nil {
		return emptyAddress, err
	}

	return resolver.pubKeyConverter.Encode(userInfo.MainGuardian.PublicKey), nil
}

func (resolver *serviceResolver) handleRegisteredAccount(userAddress []byte) (string, error) {
	// TODO properly decrypt keys from DB
	// temporary unmarshal them
	userInfo, err := resolver.getUserInfo(userAddress)
	if err != nil {
		return emptyAddress, err
	}

	if userInfo.MainGuardian.State == core.NotUsableYet {
		return resolver.pubKeyConverter.Encode(userInfo.MainGuardian.PublicKey), nil
	}

	if userInfo.SecondaryGuardian.State == core.NotUsableYet {
		return resolver.pubKeyConverter.Encode(userInfo.SecondaryGuardian.PublicKey), nil
	}

	accountAddress := erdData.NewAddressFromBytes(userAddress)

	ctxGetGuardianData, cancelGetGuardianData := context.WithTimeout(context.Background(), resolver.requestTime)
	defer cancelGetGuardianData()
	guardianData, err := resolver.proxy.GetGuardianData(ctxGetGuardianData, accountAddress)
	if err != nil {
		return emptyAddress, err
	}

	return resolver.getNextGuardianKey(guardianData, userInfo), nil
}


func (resolver *serviceResolver) getUserInfo(userAddress []byte) (*core.UserInfo, error) {
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

func (resolver *serviceResolver) computeDataAndSave(index uint32, userAddress []byte, privateKeys []crypto.PrivateKey, provider string) (*core.UserInfo, error) {
	// TODO properly encrypt keys
	// temporary marshal them and save
	mainGuardian, err := getGuardianInfoForKey(privateKeys[0])
	if err != nil {
		return &core.UserInfo{}, err
	}

	secondaryGuardian, err := getGuardianInfoForKey(privateKeys[1])
	if err != nil {
		return &core.UserInfo{}, err
	}

	userInfo := &core.UserInfo{
		Index:             index,
		MainGuardian:      mainGuardian,
		SecondaryGuardian: secondaryGuardian,
		Provider:          provider,
	}

	err = resolver.saveUserInfo(userAddress, userInfo)
	if err != nil {
		return &core.UserInfo{}, err
	}

	return userInfo, nil
}

func (resolver *serviceResolver) saveUserInfo(key []byte, userInfo *core.UserInfo) error {
	data, err := resolver.marshaller.Marshal(userInfo)
	if err != nil {
		return err
	}

	err = resolver.registeredUsersDB.Put(key, data)
	if err != nil {
		return err
	}

	return nil
}

func (resolver *serviceResolver) getNextGuardianKey(guardianData *erdData.GuardianData, userInfo *core.UserInfo) string {
	mainGuardianOnChainState := resolver.getOnChainGuardianState(guardianData, userInfo.MainGuardian)
	secondaryGuardianOnChainState := resolver.getOnChainGuardianState(guardianData, userInfo.SecondaryGuardian)
	isMainOnChain := mainGuardianOnChainState != core.MissingGuardian
	isSecondaryOnChain := secondaryGuardianOnChainState != core.MissingGuardian
	if !isMainOnChain && !isSecondaryOnChain {
		userInfo.MainGuardian.State = core.NotUsableYet
		userInfo.SecondaryGuardian.State = core.NotUsableYet
		return resolver.pubKeyConverter.Encode(userInfo.MainGuardian.PublicKey)
	}

	if isMainOnChain && isSecondaryOnChain {
		if mainGuardianOnChainState == core.PendingGuardian {
			userInfo.MainGuardian.State = core.NotUsableYet
			return resolver.pubKeyConverter.Encode(userInfo.MainGuardian.PublicKey)
		}

		userInfo.SecondaryGuardian.State = core.NotUsableYet
		return resolver.pubKeyConverter.Encode(userInfo.SecondaryGuardian.PublicKey)
	}

	if isMainOnChain {
		userInfo.SecondaryGuardian.State = core.NotUsableYet
		return resolver.pubKeyConverter.Encode(userInfo.SecondaryGuardian.PublicKey)
	}

	userInfo.MainGuardian.State = core.NotUsableYet
	return resolver.pubKeyConverter.Encode(userInfo.MainGuardian.PublicKey)
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
