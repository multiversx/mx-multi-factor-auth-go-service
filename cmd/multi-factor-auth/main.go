package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go-core/marshal/factory"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/file"
	"github.com/ElrondNetwork/elrond-go-storage/storageUnit"
	elrondFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/authentication/native"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain/cryptoProvider"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/bucket"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/factory"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers/storage"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/providers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/resolver"
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
	machineID := elrondCore.GetAnonymizedMachineID(app.Name)
	app.Version = fmt.Sprintf("%s/%s/%s-%s/%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, machineID)
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
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

	argsProxy := blockchain.ArgsElrondProxy{
		ProxyURL:            cfg.Proxy.NetworkAddress,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       cfg.Proxy.ProxyFinalityCheck,
		AllowedDeltaToFinal: cfg.Proxy.ProxyMaxNoncesDelta,
		CacheExpirationTime: time.Second * time.Duration(cfg.Proxy.ProxyCacherExpirationSeconds),
		EntityType:          erdgoCore.RestAPIEntityType(cfg.Proxy.ProxyRestAPIEntityType),
	}

	proxy, err := blockchain.NewElrondProxy(argsProxy)
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

	defer func() {
		log.LogIfError(otpStorer.Close())
	}()

	twoFactorHandler := handlers.NewTwoFactorHandler(cfg.TwoFactor.Digits, cfg.TwoFactor.Issuer)

	argsStorageHandler := storage.ArgDBOTPHandler{
		DB:          otpStorer,
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

	keyGen := crypto.NewKeyGenerator(ed25519.NewEd25519())
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

	marshalizer, err := factoryMarshalizer.NewMarshalizer(cfg.General.Marshalizer)
	if err != nil {
		return err
	}

	argsServiceResolver := resolver.ArgServiceResolver{
		Provider:                      provider,
		Proxy:                         proxy,
		KeysGenerator:                 guardianKeyGenerator,
		PubKeyConverter:               pkConv,
		RegisteredUsersDB:             registeredUsersDB,
		Marshaller:                    marshalizer,
		TxHasher:                      keccak.NewKeccak(),
		SignatureVerifier:             signer,
		GuardedTxBuilder:              builder,
		RequestTime:                   time.Duration(cfg.ServiceResolver.RequestTimeInSeconds) * time.Second,
		KeyGen:                        keyGen,
		CryptoComponentsHolderFactory: &core.CryptoComponentsHolderFactory{},
	}
	serviceResolver, err := resolver.NewServiceResolver(argsServiceResolver)
	if err != nil {
		return err
	}

	tokenHandler := native.NewAuthTokenHandler()
	args := native.ArgsNativeAuthServer{
		Proxy:           proxy,
		TokenHandler:    tokenHandler,
		Signer:          signer,
		PubKeyConverter: pkConv,
		KeyGenerator:    keyGen,
	}

	nativeAuthServer, err := native.NewNativeAuthServer(args)
	if err != nil {
		return err
	}

	webServer, err := factory.StartWebServer(configs, serviceResolver, nativeAuthServer, tokenHandler)
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
	err := elrondCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

// LoadApiConfig returns a ApiRoutesConfig by reading the config file provided
func loadApiConfig(filepath string) (config.ApiRoutesConfig, error) {
	cfg := config.ApiRoutesConfig{}
	err := elrondCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.ApiRoutesConfig{}, err
	}

	return cfg, nil
}

func attachFileLogger(log logger.Logger, flagsConfig config.ContextFlagsConfig) (elrondFactory.FileLoggingHandler, error) {
	var fileLogging elrondFactory.FileLoggingHandler
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
