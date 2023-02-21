package facade

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-chain-core-go/core/check"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

// ArgsGuardianFacade represents the DTO struct used in the auth facade constructor
type ArgsGuardianFacade struct {
	ServiceResolver core.ServiceResolver
	ApiInterface    string
	PprofEnabled    bool
}

type guardianFacade struct {
	serviceResolver core.ServiceResolver
	apiInterface    string
	pprofEnabled    bool
}

// NewGuardianFacade returns a new instance of guardianFacade
func NewGuardianFacade(args ArgsGuardianFacade) (*guardianFacade, error) {
	if check.IfNil(args.ServiceResolver) {
		return nil, ErrNilServiceResolver
	}
	if len(args.ApiInterface) == 0 {
		return nil, fmt.Errorf("%w for ApiInterface", ErrInvalidValue)
	}

	return &guardianFacade{
		serviceResolver: args.ServiceResolver,
		apiInterface:    args.ApiInterface,
		pprofEnabled:    args.PprofEnabled,
	}, nil
}

// RestApiInterface returns the interface on which the rest API should start on, based on the flags provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
//  the value is explicitly set to off, in which case it will not start at all
func (af *guardianFacade) RestApiInterface() string {
	return af.apiInterface
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (af *guardianFacade) PprofEnabled() bool {
	return af.pprofEnabled
}

// VerifyCode validates the code received
func (af *guardianFacade) VerifyCode(userAddress sdkCore.AddressHandler, request requests.VerificationPayload) error {
	return af.serviceResolver.VerifyCode(userAddress, request)
}

// RegisterUser creates a new OTP and (optionally) returns some information required
// for the user to set up the OTP on his end (eg: QR code).
func (af *guardianFacade) RegisterUser(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
	return af.serviceResolver.RegisterUser(userAddress, request)
}

// SignTransaction validates user's transaction, then signs it from guardian and returns the transaction
func (af *guardianFacade) SignTransaction(userAddress sdkCore.AddressHandler, request requests.SignTransaction) ([]byte, error) {
	return af.serviceResolver.SignTransaction(userAddress, request)
}

// SignMultipleTransactions validates user's transactions, then adds guardian signature and returns the transaction
func (af *guardianFacade) SignMultipleTransactions(userAddress sdkCore.AddressHandler, request requests.SignMultipleTransactions) ([][]byte, error) {
	return af.serviceResolver.SignMultipleTransactions(userAddress, request)
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *guardianFacade) IsInterfaceNil() bool {
	return af == nil
}
