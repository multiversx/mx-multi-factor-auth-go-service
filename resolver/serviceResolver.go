package resolver

import (
	"context"
	"fmt"
	"sync"
	"time"

	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
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
	RegisteredUsersDB  map[string]core.Storer
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
	registeredUsersDB  map[string]core.Storer
	marshaller         core.Marshaller
	requestTime        time.Duration
	registeredUsersMut sync.RWMutex
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

	resolver.registeredUsersMut.RLock()
	storer, ok := resolver.registeredUsersDB[request.Provider]
	resolver.registeredUsersMut.RUnlock()
	if !ok {
		return emptyAddress, fmt.Errorf("%w, provider %s", ErrProviderDoesNotExists, request.Provider)
	}

	isRegistered := storer.Has(userAddress)
	if isRegistered {
		return resolver.handleRegisteredAccount(userAddress, storer)
	}

	return resolver.handleNewAccount(userAddress, storer)
}

func (resolver *serviceResolver) validateCredentials(credentials string) ([]byte, error) {
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

	return accountAddress.AddressBytes(), nil
}

func (resolver *serviceResolver) handleNewAccount(userAddress []byte, storer core.Storer) (string, error) {
	index := resolver.indexHandler.GetIndex()
	privateKeys := resolver.keysGenerator.GenerateKeys(index)

	userInfo, err := resolver.computeDataAndSave(index, userAddress, privateKeys, storer)
	if err != nil {
		return emptyAddress, err
	}

	return resolver.pubKeyConverter.Encode(userInfo.MainGuardian.PublicKey), nil
}

func (resolver *serviceResolver) handleRegisteredAccount(userAddress []byte, storer core.Storer) (string, error) {
	// TODO properly decrypt keys from DB
	// temporary unmarshal them
	data, err := storer.Get(userAddress)
	if err != nil {
		return emptyAddress, err
	}

	userInfo := &core.UserInfo{}
	err = resolver.marshaller.Unmarshal(userInfo, data)
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

func (resolver *serviceResolver) computeDataAndSave(index uint64, userAddress []byte, privateKeys []crypto.PrivateKey, storer core.Storer) (*core.UserInfo, error) {
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
	}

	data, err := resolver.marshaller.Marshal(userInfo)
	if err != nil {
		return &core.UserInfo{}, err
	}

	err = storer.Put(userAddress, data)
	if err != nil {
		return &core.UserInfo{}, err
	}

	return userInfo, nil
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
