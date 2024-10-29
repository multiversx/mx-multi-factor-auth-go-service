package shared

import (
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-sdk-go/core"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	tcsCore "github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

// GroupHandler defines the actions needed to be performed by a gin API group
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
	VerifyCode(userAddress core.AddressHandler, userIp string, request requests.VerificationPayload) (*requests.OTPCodeVerifyData, error)
	RegisterUser(userAddress core.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error)
	SignMessage(userIp string, request requests.SignMessage) ([]byte, *requests.OTPCodeVerifyData, error)
	SignTransaction(userIp string, request requests.SignTransaction) ([]byte, *requests.OTPCodeVerifyData, error)
	SignMultipleTransactions(userIp string, request requests.SignMultipleTransactions) ([][]byte, *requests.OTPCodeVerifyData, error)
	SetSecurityModeNoExpire(userIp string, request requests.SecurityModeNoExpire) (*requests.OTPCodeVerifyData, error)
	UnsetSecurityModeNoExpire(userIp string, request requests.SecurityModeNoExpire) (*requests.OTPCodeVerifyData, error)
	RegisteredUsers() (uint32, error)
	TcsConfig() *tcsCore.TcsConfig
	GetMetrics() map[string]*requests.EndpointMetricsResponse
	GetMetricsForPrometheus() string
	IsInterfaceNil() bool
}

// UpgradeableHttpServerHandler defines the actions that an upgradeable http server need to do
type UpgradeableHttpServerHandler interface {
	StartHttpServer() error
	UpdateFacade(facade FacadeHandler) error
	Close() error
	IsInterfaceNil() bool
}
