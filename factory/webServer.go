package factory

import (
	"io"

	"github.com/multiversx/mx-multi-factor-auth-go-service/api/gin"
	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/facade"
	"github.com/multiversx/mx-sdk-go/authentication"
)

// StartWebServer creates and starts a web server able to respond with the metrics holder information
func StartWebServer(
	configs config.Configs,
	serviceResolver core.ServiceResolver,
	authServer authentication.AuthServer,
	tokenHandler authentication.AuthTokenHandler,
	whitelistHandler core.NativeAuthWhitelistHandler,
	statusMetricsHandler core.StatusMetricsHandler,
) (io.Closer, error) {
	argsFacade := facade.ArgsGuardianFacade{
		ServiceResolver:      serviceResolver,
		StatusMetricsHandler: statusMetricsHandler,
	}

	guardianFacade, err := facade.NewGuardianFacade(argsFacade)
	if err != nil {
		return nil, err
	}

	httpServerArgs := gin.ArgsNewWebServer{
		Facade:                     guardianFacade,
		Config:                     configs,
		AuthServer:                 authServer,
		TokenHandler:               tokenHandler,
		NativeAuthWhitelistHandler: whitelistHandler,
		StatusMetricsHandler:       statusMetricsHandler,
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
