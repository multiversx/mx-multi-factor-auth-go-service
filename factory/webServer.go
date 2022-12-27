package factory

import (
	"io"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/authentication"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/api/gin"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/facade"
)

// StartWebServer creates and starts a web server able to respond with the metrics holder information
func StartWebServer(configs config.Configs, serviceResolver core.ServiceResolver, authServer authentication.AuthServer, tokenHandler authentication.AuthTokenHandler) (io.Closer, error) {
	argsFacade := facade.ArgsAuthFacade{
		ServiceResolver: serviceResolver,
		ApiInterface:    configs.FlagsConfig.RestApiInterface,
		PprofEnabled:    configs.FlagsConfig.EnablePprof,
	}

	authFacade, err := facade.NewAuthFacade(argsFacade)
	if err != nil {
		return nil, err
	}

	httpServerArgs := gin.ArgsNewWebServer{
		Facade:          authFacade,
		ApiConfig:       configs.ApiRoutesConfig,
		AntiFloodConfig: configs.GeneralConfig.Antiflood.WebServer,
		AuthServer:      authServer,
		TokenHandler:    tokenHandler,
	}

	httpServerWrapper, err := gin.NewWebServerHandler(httpServerArgs)
	if err != nil {
		return nil, err
	}

	err = httpServerWrapper.StartHttpServer()
	if err != nil {
		return nil, err
	}

	return httpServerWrapper, nil
}
