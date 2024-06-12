package testscommon

import (
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/core"
)

type CryptoComponentsHolderStub struct {
	GetPublicKeyCalled      func() crypto.PublicKey
	GetPrivateKeyCalled     func() crypto.PrivateKey
	GetBech32Called         func() string
	GetAddressHandlerCalled func() core.AddressHandler
}

// GetPublicKey -
func (c *CryptoComponentsHolderStub) GetPublicKey() crypto.PublicKey {
	if c.GetPublicKeyCalled != nil {
		return c.GetPublicKeyCalled()
	}

	return nil
}

// GetPrivateKey -
func (c *CryptoComponentsHolderStub) GetPrivateKey() crypto.PrivateKey {
	if c.GetPrivateKeyCalled != nil {
		return c.GetPrivateKeyCalled()
	}

	return nil
}

// GetBech32 -
func (c *CryptoComponentsHolderStub) GetBech32() string {
	if c.GetBech32Called != nil {
		return c.GetBech32Called()
	}

	return ""
}

// GetAddressHandler -
func (c *CryptoComponentsHolderStub) GetAddressHandler() core.AddressHandler {
	if c.GetAddressHandlerCalled != nil {
		return c.GetAddressHandlerCalled()
	}

	return nil
}

// IsInterfaceNil -
func (c *CryptoComponentsHolderStub) IsInterfaceNil() bool {
	return c == nil
}
