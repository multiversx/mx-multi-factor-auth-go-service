package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common/logging"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/guardian"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/factory"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/providers"
	"github.com/urfave/cli"
	_ "github.com/urfave/cli"
)

const (
	filePathPlaceholder = "[path]"
	defaultLogsPath     = "logs"
	logFilePrefix       = "multi-factor-auth-go-service"
	logMaxSizeInMB      = 1024
	issuer              = "ElrondNetwork" //TODO: add issuer & digits into config.toml
	digits              = 6
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

	argsGuardian := guardian.ArgGuardian{
		Config:          cfg.Guardian,
		Proxy:           proxy,
		PubKeyConverter: pkConv,
	}
	guard, err := guardian.NewGuardian(argsGuardian)
	if err != nil {
		return err
	}

	providersMap := make(map[string]core.Provider)
	totp := providers.NewTimebasedOnetimePassword(issuer, digits)
	providersMap["totp"] = totp

	webServer, err := factory.StartWebServer(configs, providersMap, guard)
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
		args := logging.ArgsFileLogging{
			WorkingDir:      flagsConfig.WorkingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		}
		fileLogging, err = logging.NewFileLogging(args)
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
