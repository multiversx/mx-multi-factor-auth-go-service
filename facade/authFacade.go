package facade

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
)

// ArgsAuthFacade represents the DTO struct used in the auth facade constructor
type ArgsAuthFacade struct {
	ProvidersMap map[string]core.Provider
	Guardian     core.Guardian
	ApiInterface string
	PprofEnabled bool
}

type authFacade struct {
	serviceResolver core.ServiceResolver
	providersMap    map[string]core.Provider
	guardian        core.Guardian
	apiInterface    string
	pprofEnabled    bool
}

// NewAuthFacade returns a new instance of authFacade
func NewAuthFacade(args ArgsAuthFacade) (*authFacade, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &authFacade{
		providersMap: args.ProvidersMap,
		guardian:     args.Guardian,
		apiInterface: args.ApiInterface,
		pprofEnabled: args.PprofEnabled,
	}, nil
}

// checkArgs check the arguments of an ArgsNewWebServer
func checkArgs(args ArgsAuthFacade) error {
	if len(args.ProvidersMap) == 0 {
		return ErrEmptyProvidersMap
	}

	for providerType, provider := range args.ProvidersMap {
		if check.IfNil(provider) {
			return fmt.Errorf("%s:%s", ErrNilProvider, providerType)
		}
	}

	if check.IfNil(args.Guardian) {
		return ErrNilGuardian
	}
	return nil
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
func (af *authFacade) VerifyCode(request requests.VerifyCode) error {
	return af.serviceResolver.VerifyCode(request)
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
