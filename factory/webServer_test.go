package factory

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/guardian"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/providers"
	"github.com/stretchr/testify/assert"
)

func TestStartWebServer(t *testing.T) {
	t.Parallel()

	cfg := config.Configs{
		GeneralConfig: config.Config{
			Guardian: config.GuardianConfig{
				PrivateKeyFile:       "testdata/alice.pem",
				RequestTimeInSeconds: 2,
			},
			Proxy: config.ProxyConfig{
				NetworkAddress:               "http://localhost:7950",
				ProxyCacherExpirationSeconds: 600,
				ProxyRestAPIEntityType:       "proxy",
				ProxyMaxNoncesDelta:          7,
				ProxyFinalityCheck:           true,
			},
			Logs:      config.LogsConfig{},
			Antiflood: config.AntifloodConfig{},
		},
		ApiRoutesConfig: config.ApiRoutesConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			RestApiInterface: core.WebServerOffString,
		},
	}

	argsProxy := blockchain.ArgsElrondProxy{
		ProxyURL:            cfg.GeneralConfig.Proxy.NetworkAddress,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       cfg.GeneralConfig.Proxy.ProxyFinalityCheck,
		AllowedDeltaToFinal: cfg.GeneralConfig.Proxy.ProxyMaxNoncesDelta,
		CacheExpirationTime: time.Second * time.Duration(cfg.GeneralConfig.Proxy.ProxyCacherExpirationSeconds),
		EntityType:          erdgoCore.RestAPIEntityType(cfg.GeneralConfig.Proxy.ProxyRestAPIEntityType),
	}

	proxy, _ := blockchain.NewElrondProxy(argsProxy)
	pkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, &mock.LoggerMock{})
	argsGuardian := guardian.ArgGuardian{
		Config:          cfg.GeneralConfig.Guardian,
		Proxy:           proxy,
		PubKeyConverter: pkConv,
	}
	guard, _ := guardian.NewGuardian(argsGuardian)

	providersMap := make(map[string]core.Provider)
	totp := providers.NewTimebasedOnetimePassword("ElrondNetwork", 6)
	providersMap["totp"] = totp

	webServer, err := StartWebServer(cfg, providersMap, guard)
	assert.Nil(t, err)
	assert.NotNil(t, webServer)

	err = webServer.Close()
	assert.Nil(t, err)
}
