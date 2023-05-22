package factory

import (
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon/middleware"
	"github.com/multiversx/mx-sdk-go/authentication/native/mock"
	"github.com/stretchr/testify/assert"
)

func TestStartWebServer(t *testing.T) {
	t.Parallel()

	cfg := config.Configs{
		GeneralConfig: config.Config{
			Guardian: config.GuardianConfig{
				MnemonicFile:         "testdata/multiversx.mnemonic",
				RequestTimeInSeconds: 2,
			},
			Logs:      config.LogsConfig{},
			Antiflood: config.AntifloodConfig{},
		},
		ApiRoutesConfig: config.ApiRoutesConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			RestApiInterface: core.WebServerOffString,
		},
	}

	webServer, err := StartWebServer(
		cfg,
		&testscommon.ServiceResolverStub{},
		&mock.AuthServerStub{},
		&mock.AuthTokenHandlerStub{},
		&middleware.NativeAuthWhitelistHandlerStub{},
		&testscommon.StatusMetricsStub{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, webServer)

	err = webServer.Close()
	assert.Nil(t, err)
}
