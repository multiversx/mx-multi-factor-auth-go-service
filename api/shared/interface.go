package shared

import (
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/gin-gonic/gin"
)

// GroupHandler defines the actions needed to be performed by an gin API group
type GroupHandler interface {
	UpdateFacade(newFacade FacadeHandler) error
	RegisterRoutes(
		ws *gin.RouterGroup,
		apiConfig config.ApiRoutesConfig,
	)
	IsInterfaceNil() bool
}

// FacadeHandler defines all the methods that a facade should implement
type FacadeHandler interface {
	RestApiInterface() string
	PprofEnabled() bool
	VerifyCode(userAddress erdCore.AddressHandler, request requests.VerificationPayload) error
	RegisterUser(userAddress erdCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error)
	SignTransaction(userAddress erdCore.AddressHandler, request requests.SignTransaction) ([]byte, error)
	SignMultipleTransactions(userAddress erdCore.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error)
	IsInterfaceNil() bool
}

// UpgradeableHttpServerHandler defines the actions that an upgradeable http server need to do
type UpgradeableHttpServerHandler interface {
	StartHttpServer() error
	UpdateFacade(facade FacadeHandler) error
	Close() error
	IsInterfaceNil() bool
}
