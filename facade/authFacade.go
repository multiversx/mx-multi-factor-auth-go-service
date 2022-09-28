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
	providersMap map[string]core.Provider
	guardian     core.Guardian
	apiInterface string
	pprofEnabled bool
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

// ValidateAndSend validates the request and trigger the guardian to sign and send the given transaction
// if verification passed
func (af *authFacade) ValidateAndSend(request requests.SendTransaction) (string, error) {
	if len(request.Codes) == 0 {
		return "", ErrEmptyCodesArray
	}
	if request.Account != request.Tx.SndAddr {
		return "", ErrInvalidSenderAddress
	}

	for _, code := range request.Codes {
		provider, exists := af.providersMap[code.Provider]
		if !exists {
			return "", fmt.Errorf("%s: %s", code.Provider, ErrProviderDoesNotExists)
		}
		isValid, err := provider.Validate(request.Account, code.Code)
		if err != nil {
			return "", fmt.Errorf("%s: %s", code.Provider, err.Error())
		}
		if !isValid {
			return "", ErrRequestNotValid
		}
	}

	hash, err := af.guardian.ValidateAndSend(request.Tx)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// RegisterUser creates a new OTP for the given provider
// and (optionally) returns some information required for the user to set up the OTP on his end (eg: QR code).
func (af *authFacade) RegisterUser(request requests.Register) ([]byte, error) {
	provider, exists := af.providersMap[request.Provider]
	if !exists {
		return make([]byte, 0), fmt.Errorf("%w for provider %s", ErrProviderDoesNotExists, request.Provider)
	}

	guardianAddress := af.guardian.GetAddress()
	if request.Guardian != guardianAddress {
		return make([]byte, 0), fmt.Errorf("%w for guardian %s", ErrInvalidGuardian, request.Guardian)
	}

	qrBytes, err := provider.RegisterUser(request.Account)
	if err != nil {
		return make([]byte, 0), err
	}

	err = af.guardian.AddUser(request.Account)
	if err != nil {
		return make([]byte, 0), err
	}

	return qrBytes, nil
}

// GetGuardianAddress returns the address of the guardian
func (af *authFacade) GetGuardianAddress() string {
	return af.guardian.GetAddress()
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *authFacade) IsInterfaceNil() bool {
	return af == nil
}
