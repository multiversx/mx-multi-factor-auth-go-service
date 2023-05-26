package facade

import (
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-chain-core-go/core/check"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

// ArgsGuardianFacade represents the DTO struct used in the auth facade constructor
type ArgsGuardianFacade struct {
	ServiceResolver      core.ServiceResolver
	StatusMetricsHandler core.StatusMetricsHandler
}

type guardianFacade struct {
	serviceResolver core.ServiceResolver
	statusMetrics   core.StatusMetricsHandler
}

// NewGuardianFacade returns a new instance of guardianFacade
func NewGuardianFacade(args ArgsGuardianFacade) (*guardianFacade, error) {
	if check.IfNil(args.ServiceResolver) {
		return nil, ErrNilServiceResolver
	}
	if check.IfNil(args.StatusMetricsHandler) {
		return nil, core.ErrNilMetricsHandler
	}

	return &guardianFacade{
		serviceResolver: args.ServiceResolver,
		statusMetrics:   args.StatusMetricsHandler,
	}, nil
}

// VerifyCode validates the code received
func (gf *guardianFacade) VerifyCode(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) error {
	return gf.serviceResolver.VerifyCode(userAddress, userIp, request)
}

// RegisterUser creates a new OTP and (optionally) returns some information required
// for the user to set up the OTP on his end (eg: QR code).
func (gf *guardianFacade) RegisterUser(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) (*requests.OTP, string, error) {
	return gf.serviceResolver.RegisterUser(userAddress, request)
}

// SignTransaction validates user's transaction, then signs it from guardian and returns the transaction
func (gf *guardianFacade) SignTransaction(userIp string, request requests.SignTransaction) ([]byte, error) {
	return gf.serviceResolver.SignTransaction(userIp, request)
}

// SignMultipleTransactions validates user's transactions, then adds guardian signature and returns the transaction
func (gf *guardianFacade) SignMultipleTransactions(userIp string, request requests.SignMultipleTransactions) ([][]byte, error) {
	return gf.serviceResolver.SignMultipleTransactions(userIp, request)
}

// RegisteredUsers returns the number of registered users
func (gf *guardianFacade) RegisteredUsers() (uint32, error) {
	return gf.serviceResolver.RegisteredUsers()
}

// TcsConfig returns the current configuration of the TCS
func (gf *guardianFacade) TcsConfig() *core.TcsConfig {
	return gf.serviceResolver.TcsConfig()
}

// GetMetrics will return metrics in json format
func (gf *guardianFacade) GetMetrics() map[string]*requests.EndpointMetricsResponse {
	return gf.statusMetrics.GetAll()
}

// GetMetricsForPrometheus will return metrics in prometheus format
func (gf *guardianFacade) GetMetricsForPrometheus() string {
	return gf.statusMetrics.GetMetricsForPrometheus()
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *guardianFacade) IsInterfaceNil() bool {
	return af == nil
}
