package tcs

import (
	"os"
	"os/signal"
	"syscall"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-storage-go/storageUnit"
	"github.com/multiversx/mx-sdk-go/authentication/native"
	"github.com/multiversx/mx-sdk-go/core/http"

	"github.com/multiversx/mx-multi-factor-auth-go-service/api/middleware"
	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/factory"
	storageFactory "github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage/factory"
	"github.com/multiversx/mx-multi-factor-auth-go-service/metrics"
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
	cryptoComponents, err := factory.CreateCoreCryptoComponents(tr.configs.GeneralConfig.PubKey)
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

	httpClient := http.NewHttpClientWrapper(nil, tr.configs.ExternalConfig.Api.NetworkAddress)
	httpClientWrapper, err := core.NewHttpClientWrapper(httpClient)
	if err != nil {
		return err
	}

	twoFactorHandler, err := factory.CreateOTPHandler(tr.configs)
	if err != nil {
		return err
	}

	secureOtpHandler, err := factory.CreateSecureOTPHandler(tr.configs)
	if err != nil {
		return err
	}

	serviceResolver, err := factory.CreateServiceResolver(tr.configs, cryptoComponents, httpClientWrapper, registeredUsersDB, twoFactorHandler, secureOtpHandler)
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
		Signer:            cryptoComponents.Signer(),
		PubKeyConverter:   cryptoComponents.PubkeyConverter(),
		KeyGenerator:      cryptoComponents.KeyGenerator(),
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

	return webServer.Close()
}
