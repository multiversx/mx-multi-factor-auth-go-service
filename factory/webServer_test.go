package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/authentication/native/mock"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestStartWebServer(t *testing.T) {
	t.Parallel()

	cfg := config.Configs{
		GeneralConfig: config.Config{
			Guardian: config.GuardianConfig{
				MnemonicFile:         "testdata/elrond.mnemonic",
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

	webServer, err := StartWebServer(cfg, &testscommon.ServiceResolverStub{}, &mock.AuthServerStub{}, &mock.AuthTokenHandlerStub{})
	assert.Nil(t, err)
	assert.NotNil(t, webServer)

	err = webServer.Close()
	assert.Nil(t, err)
}
