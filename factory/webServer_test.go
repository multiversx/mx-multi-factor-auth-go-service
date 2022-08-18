package factory

import (
	"testing"

	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/providers"
	"github.com/stretchr/testify/assert"
)

//TODO: modify and to tests for WebServer

func TestStartWebServer(t *testing.T) {
	t.Parallel()

	cfg := config.Configs{
		GeneralConfig: config.Config{
			Logs:      config.LogsConfig{},
			Antiflood: config.AntifloodConfig{},
			Computer:  config.ComputerConfig{},
		},

		ApiRoutesConfig: config.ApiRoutesConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			RestApiInterface: core.WebServerOffString,
		},
	}

	webServer, err := StartWebServer(cfg, providers.NewRarityCalculator())
	assert.Nil(t, err)
	assert.NotNil(t, webServer)

	err = webServer.Close()
	assert.Nil(t, err)
}
