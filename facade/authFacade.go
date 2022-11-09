package facade

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// ArgsAuthFacade represents the DTO struct used in the auth facade constructor
type ArgsAuthFacade struct {
	ServiceResolver core.ServiceResolver
	ApiInterface    string
	PprofEnabled    bool
}

type authFacade struct {
	serviceResolver core.ServiceResolver
	apiInterface    string
	pprofEnabled    bool
}

// NewAuthFacade returns a new instance of authFacade
func NewAuthFacade(args ArgsAuthFacade) (*authFacade, error) {
	if check.IfNil(args.ServiceResolver) {
		return nil, ErrNilServiceResolver
	}
	if len(args.ApiInterface) == 0 {
		return nil, fmt.Errorf("%w for ApiInterface", ErrInvalidValue)
	}

	return &authFacade{
		serviceResolver: args.ServiceResolver,
		apiInterface:    args.ApiInterface,
		pprofEnabled:    args.PprofEnabled,
	}, nil
}

// RestApiInterface returns the interface on which the rest API should start on, based on the flags provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
//  the value is explicitly set to off, in which case it will not start at all
func (af *authFacade) RestApiInterface() string {
	return af.apiInterface
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (af *authFacade) PprofEnabled() bool {
	return af.pprofEnabled
}

// VerifyCode validates the code received
func (af *authFacade) VerifyCode(request requests.VerificationPayload) error {
	return af.serviceResolver.VerifyCode(request)
}

// RegisterUser creates a new OTP and (optionally) returns some information required
// for the user to set up the OTP on his end (eg: QR code).
func (af *authFacade) RegisterUser(request requests.RegistrationPayload) ([]byte, error) {
	return af.serviceResolver.RegisterUser(request)
}

// GetGuardianAddress returns the address of a unique guardian
func (af *authFacade) GetGuardianAddress(request requests.GetGuardianAddress) (string, error) {
	return af.serviceResolver.GetGuardianAddress(request)
}

// SendTransaction validates user's transaction, then signs it from guardian and returns the transaction
func (af *authFacade) SendTransaction(request requests.SendTransaction) ([]byte, error) {
	return af.serviceResolver.SendTransaction(request)
}

// SendMultipleTransactions validates user's transactions, then adds guardian signature and returns the transaction
func (af *authFacade) SendMultipleTransactions(request requests.SendMultipleTransaction) ([][]byte, error) {
	return af.serviceResolver.SendMultipleTransactions(request)
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *authFacade) IsInterfaceNil() bool {
	return af == nil
}
