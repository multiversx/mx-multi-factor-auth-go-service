package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/bucket"
	"github.com/multiversx/multi-factor-auth-go-service/factory"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/multi-factor-auth-go-service/providers"
	"github.com/multiversx/multi-factor-auth-go-service/resolver"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	factoryMarshalizer "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	chainFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/multiversx/mx-chain-storage-go/storageUnit"
	"github.com/multiversx/mx-sdk-go/authentication/native"
	"github.com/multiversx/mx-sdk-go/authentication/native/mock"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/core/http"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/urfave/cli"
	_ "github.com/urfave/cli"
)

const (
	filePathPlaceholder = "[path]"
	defaultLogsPath     = "logs"
	logFilePrefix       = "multi-factor-auth-go-service"
	logMaxSizeInMB      = 1024
	userAddressLength   = 32
)

var log = logger.GetOrCreate("main")

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//            go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
// windows:
//            for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//            go build -i -v -ldflags="-X main.appVersion=%VERS%"
var appVersion = "undefined"

func main() {
	app := cli.NewApp()
	app.Name = "Relay CLI app"
	app.Usage = "This is the entry point for the multi-factor authentication service written in go"
	app.Flags = getFlags()
	machineID := chainCore.GetAnonymizedMachineID(app.Name)
	app.Version = fmt.Sprintf("%s/%s/%s-%s/%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, machineID)
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startService(c, app.Version)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startService(ctx *cli.Context, version string) error {
	flagsConfig := getFlagsConfig(ctx)

	fileLogging, errLogger := attachFileLogger(log, flagsConfig)
	if errLogger != nil {
		return errLogger
	}

	log.Info("starting multi-factor authentication service", "version", version, "pid", os.Getpid())

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	cfg, err := loadConfig(flagsConfig.ConfigurationFile)
	if err != nil {
		return err
	}

	apiRoutesConfig, err := loadApiConfig(flagsConfig.ConfigurationApiFile)
	if err != nil {
		return err
	}
	log.Debug("config", "file", flagsConfig.ConfigurationApiFile)

	if !check.IfNil(fileLogging) {
		err = fileLogging.ChangeFileLifeSpan(time.Second*time.Duration(cfg.Logs.LogFileLifeSpanInSec), logMaxSizeInMB)
		if err != nil {
			return err
		}
	}

	configs := config.Configs{
		GeneralConfig:   cfg,
		ApiRoutesConfig: apiRoutesConfig,
		FlagsConfig:     flagsConfig,
	}

	argsProxy := blockchain.ArgsProxy{
		ProxyURL:            cfg.Proxy.NetworkAddress,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       cfg.Proxy.ProxyFinalityCheck,
		AllowedDeltaToFinal: cfg.Proxy.ProxyMaxNoncesDelta,
		CacheExpirationTime: time.Second * time.Duration(cfg.Proxy.ProxyCacherExpirationSeconds),
		EntityType:          sdkCore.RestAPIEntityType(cfg.Proxy.ProxyRestAPIEntityType),
	}

	proxy, err := blockchain.NewProxy(argsProxy)
	if err != nil {
		return err
	}

	pkConv, err := pubkeyConverter.NewBech32PubkeyConverter(userAddressLength, log)
	if err != nil {
		return err
	}

	otpStorer, err := storageUnit.NewStorageUnitFromConf(cfg.OTP.Cache, cfg.OTP.DB)
	if err != nil {
		return err
	}

	mongodbStorer, err := createMongoDBStorerHandler(cfg.MongoDB)
	if err != nil {
		return err
	}
	defer func() {
		log.LogIfError(mongodbStorer.Close())
	}()

	defer func() {
		log.LogIfError(otpStorer.Close())
	}()

	twoFactorHandler := handlers.NewTwoFactorHandler(cfg.TwoFactor.Digits, cfg.TwoFactor.Issuer)

	argsStorageHandler := storage.ArgDBOTPHandler{
		DB:          mongodbStorer,
		TOTPHandler: twoFactorHandler,
	}
	otpStorageHandler, err := storage.NewDBOTPHandler(argsStorageHandler)
	if err != nil {
		return err
	}

	argsProvider := providers.ArgTimeBasedOneTimePassword{
		TOTPHandler:       twoFactorHandler,
		OTPStorageHandler: otpStorageHandler,
	}
	provider, err := providers.NewTimeBasedOnetimePassword(argsProvider)
	if err != nil {
		return err
	}

	keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
	mnemonic, err := ioutil.ReadFile(cfg.Guardian.MnemonicFile)
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

	registeredUsersDB, err := createRegisteredUsersDB(cfg)
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

	cryptoComponentsHolderFactory, err := core.NewCryptoComponentsHolderFactory(keyGen)
	if err != nil {
		return err
	}

	argsServiceResolver := resolver.ArgServiceResolver{
		Provider:                      provider,
		Proxy:                         proxy,
		KeysGenerator:                 guardianKeyGenerator,
		PubKeyConverter:               pkConv,
		RegisteredUsersDB:             registeredUsersDB,
		UserDataMarshaller:            gogoMarshaller,
		EncryptionMarshaller:          jsonMarshaller,
		TxMarshaller:                  jsonTxMarshaller,
		TxHasher:                      keccak.NewKeccak(),
		SignatureVerifier:             signer,
		GuardedTxBuilder:              builder,
		RequestTime:                   time.Duration(cfg.ServiceResolver.RequestTimeInSeconds) * time.Second,
		KeyGen:                        keyGen,
		CryptoComponentsHolderFactory: cryptoComponentsHolderFactory,
		SkipTxUserSigVerify:           cfg.ServiceResolver.SkipTxUserSigVerify,
	}
	serviceResolver, err := resolver.NewServiceResolver(argsServiceResolver)
	if err != nil {
		return err
	}

	tokenHandler := native.NewAuthTokenHandler()
	httpClientWrapper := http.NewHttpClientWrapper(nil, cfg.Api.NetworkAddress)
	args := native.ArgsNativeAuthServer{
		HttpClientWrapper: httpClientWrapper,
		TokenHandler:      tokenHandler,
		Signer:            signer,
		PubKeyConverter:   pkConv,
		KeyGenerator:      keyGen,
	}

	_, err = native.NewNativeAuthServer(args)
	if err != nil {
		return err
	}
	nativeAuthServerMock := &mock.AuthServerStub{}

	webServer, err := factory.StartWebServer(configs, serviceResolver, nativeAuthServerMock, tokenHandler)
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

// TODO: addapt and use this for redis setup
// func createRedisStorerHandler(cfg config.RedisConfig) (core.Storer, error) {
// 	redisClient, err := redis.CreateRedisClient(cfg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return storage.NewRedisStorerHandler(redisClient)
// }

func createMongoDBStorerHandler(cfg config.MongoDBConfig) (core.Storer, error) {
	client, err := mongodb.NewMongoDBClient(cfg)
	if err != nil {
		return nil, err
	}

	return storage.NewMongoDBStorerHandler(client, mongodb.UsersCollection)
}

func createRegisteredUsersDB(cfg config.Config) (core.ShardedStorageWithIndex, error) {
	bucketIDProvider, err := bucket.NewBucketIDProvider(cfg.Buckets.NumberOfBuckets)
	if err != nil {
		return nil, err
	}

	bucketIndexHandlers := make(map[uint32]core.BucketIndexHandler, cfg.Buckets.NumberOfBuckets)
	var bucketStorer core.Storer
	for i := uint32(0); i < cfg.Buckets.NumberOfBuckets; i++ {
		cacheCfg := cfg.Users.Cache
		cacheCfg.Name = fmt.Sprintf("%s_%d", cacheCfg.Name, i)
		dbCfg := cfg.Users.DB
		dbCfg.FilePath = fmt.Sprintf("%s_%d", dbCfg.FilePath, i)

		bucketStorer, err = storageUnit.NewStorageUnitFromConf(cacheCfg, dbCfg)
		if err != nil {
			return nil, err
		}

		bucketIndexHandlers[i], err = bucket.NewBucketIndexHandler(bucketStorer)
		if err != nil {
			return nil, err
		}
	}

	argsShardedStorageWithIndex := bucket.ArgShardedStorageWithIndex{
		BucketIDProvider: bucketIDProvider,
		BucketHandlers:   bucketIndexHandlers,
	}

	return bucket.NewShardedStorageWithIndex(argsShardedStorageWithIndex)
}

func loadConfig(filepath string) (config.Config, error) {
	cfg := config.Config{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

// LoadApiConfig returns a ApiRoutesConfig by reading the config file provided
func loadApiConfig(filepath string) (config.ApiRoutesConfig, error) {
	cfg := config.ApiRoutesConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.ApiRoutesConfig{}, err
	}

	return cfg, nil
}

func attachFileLogger(log logger.Logger, flagsConfig config.ContextFlagsConfig) (chainFactory.FileLoggingHandler, error) {
	var fileLogging chainFactory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		args := file.ArgsFileLogging{
			WorkingDir:      flagsConfig.WorkingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		}
		fileLogging, err = file.NewFileLogging(args)
		if err != nil {
			return nil, fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)
	logger.ToggleLoggerName(flagsConfig.EnableLogName)
	logLevelFlagValue := flagsConfig.LogLevel
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}

	if flagsConfig.DisableAnsiColor {
		err = logger.RemoveLogObserver(os.Stdout)
		if err != nil {
			return nil, err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			return nil, err
		}
	}
	log.Trace("logger updated", "level", logLevelFlagValue, "disable ANSI color", flagsConfig.DisableAnsiColor)

	return fileLogging, nil
}
