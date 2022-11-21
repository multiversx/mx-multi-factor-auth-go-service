package resolver

import erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"

// GetGuardianAddress -
func (resolver *serviceResolver) GetGuardianAddress(userAddress erdCore.AddressHandler) (string, error) {
	return resolver.getGuardianAddress(userAddress)
}
