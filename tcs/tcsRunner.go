package tcs

import (
	"crypto"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	factoryMarshalizer "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-storage-go/storageUnit"
	"github.com/multiversx/mx-multi-factor-auth-go-service/api/middleware"
	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/factory"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/encryption"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/frozenOtp"
	storageFactory "github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage/factory"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/twofactor"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/twofactor/sec51"
	"github.com/multiversx/mx-multi-factor-auth-go-service/metrics"
	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
	"github.com/multiversx/mx-multi-factor-auth-go-service/resolver"
	"github.com/multiversx/mx-sdk-go/authentication/native"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core/http"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	userAddressLength = 32
	hashType          = crypto.SHA1
)

var log = logger.GetOrCreate("tcsRunner")

type tcsRunner struct {
	configs *config.Configs
}

// NewTcsRunner will create a new tcs runner instance
func NewTcsRunner(cfgs *config.Configs) (*tcsRunner, error) {
	if cfgs == nil {
		return nil, ErrNilConfigs
	}

	return &tcsRunner{
		configs: cfgs,
	}, nil
}

// Start will trigger the tcs service
func (tr *tcsRunner) Start() error {
	pkConv, err := pubkeyConverter.NewBech32PubkeyConverter(userAddressLength, log)
	if err != nil {
		return err
	}

	statusMetricsHandler := metrics.NewStatusMetrics()

	shardedStorageFactory := storageFactory.NewStorageWithIndexFactory(tr.configs.GeneralConfig, tr.configs.ExternalConfig, statusMetricsHandler)
	registeredUsersDB, err := shardedStorageFactory.Create()
	if err != nil {
		return err
	}

	defer func() {
		log.LogIfError(registeredUsersDB.Close())
	}()

	gogoMarshaller, err := factoryMarshalizer.NewMarshalizer(factoryMarshalizer.GogoProtobuf)
	if err != nil {
		return err
	}

	jsonTxMarshaller, err := factoryMarshalizer.NewMarshalizer(factoryMarshalizer.TxJsonMarshalizer)
	if err != nil {
		return err
	}

	jsonMarshaller, err := factoryMarshalizer.NewMarshalizer(factoryMarshalizer.JsonMarshalizer)
	if err != nil {
		return err
	}

	otpProvider := sec51.NewSec51Wrapper(tr.configs.GeneralConfig.TwoFactor.Digits, tr.configs.GeneralConfig.TwoFactor.Issuer)
	twoFactorHandler, err := twofactor.NewTwoFactorHandler(otpProvider, hashType)
	if err != nil {
		return err
	}

	rateLimiter, err := redis.CreateRedisRateLimiter(tr.configs.ExternalConfig.Redis, tr.configs.GeneralConfig.TwoFactor)
	if err != nil {
		return err
	}

	frozenOtpArgs := frozenOtp.ArgsFrozenOtpHandler{
		RateLimiter: rateLimiter,
	}
	frozenOtpHandler, err := frozenOtp.NewFrozenOtpHandler(frozenOtpArgs)
	if err != nil {
		return err
	}

	keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
	mnemonic, err := ioutil.ReadFile(tr.configs.GeneralConfig.Guardian.MnemonicFile)
	if err != nil {
		return err
	}
	argsGuardianKeyGenerator := core.ArgGuardianKeyGenerator{
		Mnemonic: data.Mnemonic(mnemonic),
		KeyGen:   keyGen,
	}
	guardianKeyGenerator, err := core.NewGuardianKeyGenerator(argsGuardianKeyGenerator)
	if err != nil {
		return err
	}

	signer := cryptoProvider.NewSigner()
	builder, err := builders.NewTxBuilder(signer)
	if err != nil {
		return err
	}

	cryptoComponentsHolderFactory, err := core.NewCryptoComponentsHolderFactory(keyGen)
	if err != nil {
		return err
	}

	managedPrivateKey, err := guardianKeyGenerator.GenerateManagedKey()
	if err != nil {
		return err
	}

	encryptor, err := encryption.NewEncryptor(jsonMarshaller, keyGen, managedPrivateKey)
	if err != nil {
		return err
	}

	userEncryptor, err := resolver.NewUserEncryptor(encryptor)
	if err != nil {
		return err
	}

	httpClient := http.NewHttpClientWrapper(nil, tr.configs.ExternalConfig.Api.NetworkAddress)
	httpClientWrapper, err := core.NewHttpClientWrapper(httpClient)
	if err != nil {
		return err
	}

	argsServiceResolver := resolver.ArgServiceResolver{
		UserEncryptor:                 userEncryptor,
		TOTPHandler:                   twoFactorHandler,
		FrozenOtpHandler:              frozenOtpHandler,
		HttpClientWrapper:             httpClientWrapper,
		KeysGenerator:                 guardianKeyGenerator,
		PubKeyConverter:               pkConv,
		RegisteredUsersDB:             registeredUsersDB,
		UserDataMarshaller:            gogoMarshaller,
		TxMarshaller:                  jsonTxMarshaller,
		TxHasher:                      keccak.NewKeccak(),
		SignatureVerifier:             signer,
		GuardedTxBuilder:              builder,
		KeyGen:                        keyGen,
		CryptoComponentsHolderFactory: cryptoComponentsHolderFactory,
		Config:                        tr.configs.GeneralConfig.ServiceResolver,
	}
	serviceResolver, err := resolver.NewServiceResolver(argsServiceResolver)
	if err != nil {
		return err
	}

	nativeAuthServerCacher, err := storageUnit.NewCache(tr.configs.GeneralConfig.NativeAuthServer.Cache)
	if err != nil {
		return err
	}

	tokenHandler := native.NewAuthTokenHandler()
	args := native.ArgsNativeAuthServer{
		HttpClientWrapper: httpClient,
		TokenHandler:      tokenHandler,
		Signer:            signer,
		PubKeyConverter:   pkConv,
		KeyGenerator:      keyGen,
		TimestampsCacher:  nativeAuthServerCacher,
	}

	nativeAuthServer, err := native.NewNativeAuthServer(args)
	if err != nil {
		return err
	}

	nativeAuthWhitelistHandler := middleware.NewNativeAuthWhitelistHandler(tr.configs.ApiRoutesConfig.APIPackages)

	webServer, err := factory.StartWebServer(*tr.configs, serviceResolver, nativeAuthServer, tokenHandler, nativeAuthWhitelistHandler, statusMetricsHandler)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("application closing, calling Close on all subcomponents...")

	var lastErr error

	err = webServer.Close()
	if err != nil {
		lastErr = err
	}

	return lastErr
}
