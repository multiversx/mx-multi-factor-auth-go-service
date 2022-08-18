package facade

import (
	"fmt"

	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/providers"
)

// ArgsAuthFacade represents the DTO struct used in the relayer facade constructor
type ArgsAuthFacade struct {
	Providers    map[string]providers.Provider
	Guardian     core.Guardian
	ApiInterface string
	PprofEnabled bool
}

type authFacade struct {
	providers    map[string]providers.Provider
	guardian     core.Guardian
	apiInterface string
	pprofEnabled bool
}

// NewAuthFacade is the implementation of the relayer facade
func NewAuthFacade(args ArgsAuthFacade) (*authFacade, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &authFacade{
		providers:    args.Providers,
		guardian:     args.Guardian,
		apiInterface: args.ApiInterface,
		pprofEnabled: args.PprofEnabled,
	}, nil
}

// checkArgs check the arguments of an ArgsNewWebServer
func checkArgs(args ArgsAuthFacade) error {
	// TODO: check args
	return nil
}

// RestApiInterface returns the interface on which the rest API should start on, based on the flags provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
//  the value is explicitly set to off, in which case it will not start at all
func (rf *authFacade) RestApiInterface() string {
	return rf.apiInterface
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (rf *authFacade) PprofEnabled() bool {
	return rf.pprofEnabled
}

// Validate returns rarity for the specified nft.
func (rf *authFacade) Validate(request providers.GuardianValidateRequest) (string, error) {
	provider, exists := rf.providers["totp"]
	if !exists {
		return "", fmt.Errorf("%s: provider does not exists", "totp")
	}
	isValid, err := provider.Validate(request.Account, request.Codes.Totp)
	if err != nil {
		return "", fmt.Errorf("%s: %s", provider, err.Error())
	}
	if !isValid {
		return "", nil
	}

	hash, err := rf.guardian.ValidateAndSend(request.Tx)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// Register returns rarity for the specified nft.
func (rf *authFacade) RegisterUser(request providers.GuardianRegisterRequest) ([]byte, error) {
	provider, exists := rf.providers[request.Type]
	if !exists {
		return nil, fmt.Errorf("%s: provider does not exists", request.Type)
	}
	return provider.RegisterUser(request.Account)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rf *authFacade) IsInterfaceNil() bool {
	return rf == nil
}
