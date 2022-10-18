package facade

import (
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

// VerifyCodes validates the code received
func (af *authFacade) VerifyCodes(request requests.VerifyCodes) error {
	return af.serviceResolver.VerifyCodes(request)
}

// RegisterUser creates a new OTP for the given provider
// and (optionally) returns some information required for the user to set up the OTP on his end (eg: QR code).
func (af *authFacade) RegisterUser(request requests.Register) ([]byte, error) {
	return af.serviceResolver.RegisterUser(request)
}

// GetGuardianAddress returns the address of a unique guardian
func (af *authFacade) GetGuardianAddress(request requests.GetGuardianAddress) (string, error) {
	return af.serviceResolver.GetGuardianAddress(request)
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *authFacade) IsInterfaceNil() bool {
	return af == nil
}
