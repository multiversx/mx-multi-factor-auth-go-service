package main

import (
	"crypto"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/api/middleware"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/factory"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/encryption"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/frozenOtp"
	storageFactory "github.com/multiversx/multi-factor-auth-go-service/handlers/storage/factory"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/twofactor"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/twofactor/sec51"
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
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core/http"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/urfave/cli"
)

const (
	filePathPlaceholder = "[path]"
	defaultLogsPath     = "logs"
	logFilePrefix       = "multi-factor-auth-go-service"
	logMaxSizeInMB      = 1024
	userAddressLength   = 32
	hashType            = crypto.SHA1
)

var log = logger.GetOrCreate("main")

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//
//	go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
//
// windows:
//
//	for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//	go build -i -v -ldflags="-X main.appVersion=%VERS%"
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

	configs, err := readConfigs(flagsConfig)
	if err != nil {
		return err
	}

	if !check.IfNil(fileLogging) {
		err = fileLogging.ChangeFileLifeSpan(time.Second*time.Duration(configs.GeneralConfig.Logs.LogFileLifeSpanInSec), logMaxSizeInMB)
		if err != nil {
			return err
		}
	}

	pkConv, err := pubkeyConverter.NewBech32PubkeyConverter(userAddressLength, log)
	if err != nil {
		return err
	}

	shardedStorageFactory := storageFactory.NewStorageWithIndexFactory(configs.GeneralConfig, configs.ExternalConfig)
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

	otpProvider := sec51.NewSec51Wrapper(configs.GeneralConfig.TwoFactor.Digits, configs.GeneralConfig.TwoFactor.Issuer)
	twoFactorHandler, err := twofactor.NewTwoFactorHandler(otpProvider, hashType)
	if err != nil {
		return err
	}

	frozenOtpArgs := frozenOtp.ArgsFrozenOtpHandler{
		MaxFailures: uint8(configs.GeneralConfig.TwoFactor.MaxFailures),
		BackoffTime: time.Second * time.Duration(configs.GeneralConfig.TwoFactor.BackoffTimeInSeconds),
	}
	frozenOtpHandler, err := frozenOtp.NewFrozenOtpHandler(frozenOtpArgs)
	if err != nil {
		return err
	}

	keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
	mnemonic, err := ioutil.ReadFile(configs.GeneralConfig.Guardian.MnemonicFile)
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

	httpClient := http.NewHttpClientWrapper(nil, configs.ExternalConfig.Api.NetworkAddress)
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
		Config:                        configs.GeneralConfig.ServiceResolver,
	}
	serviceResolver, err := resolver.NewServiceResolver(argsServiceResolver)
	if err != nil {
		return err
	}

	nativeAuthServerCacher, err := storageUnit.NewCache(configs.GeneralConfig.NativeAuthServer.Cache)
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

	_, err = native.NewNativeAuthServer(args)
	if err != nil {
		return err
	}
	nativeAuthServer := &mock.AuthServerStub{}

	nativeAuthWhitelistHandler := middleware.NewNativeAuthWhitelistHandler(configs.ApiRoutesConfig.APIPackages)

	webServer, err := factory.StartWebServer(*configs, serviceResolver, nativeAuthServer, tokenHandler, nativeAuthWhitelistHandler)
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

func readConfigs(flagsConfig config.ContextFlagsConfig) (*config.Configs, error) {
	cfg, err := loadConfig(flagsConfig.ConfigurationFile)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", flagsConfig.ConfigurationFile)

	apiRoutesConfig, err := loadApiConfig(flagsConfig.ConfigurationApiFile)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", flagsConfig.ConfigurationApiFile)

	externalConfig, err := loadExternalConfig(flagsConfig.ConfigurationExternalFile)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", flagsConfig.ConfigurationExternalFile)

	return &config.Configs{
		GeneralConfig:   cfg,
		ExternalConfig:  externalConfig,
		ApiRoutesConfig: apiRoutesConfig,
		FlagsConfig:     flagsConfig,
	}, nil
}

func loadConfig(filepath string) (config.Config, error) {
	cfg := config.Config{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

// loadApiConfig returns a ApiRoutesConfig by reading the config file provided
func loadApiConfig(filepath string) (config.ApiRoutesConfig, error) {
	cfg := config.ApiRoutesConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.ApiRoutesConfig{}, err
	}

	return cfg, nil
}

func loadExternalConfig(filepath string) (config.ExternalConfig, error) {
	cfg := config.ExternalConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.ExternalConfig{}, err
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
