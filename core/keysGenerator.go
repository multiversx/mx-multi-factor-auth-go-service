package core

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

// ArgGuardianKeyGenerator is the DTO used to create a new instance of guardian key generator
type ArgGuardianKeyGenerator struct {
	MainKey      string
	SecondaryKey string
	KeyGen       KeyGenerator
}

type guardianKeyGenerator struct {
	mainKey      string
	secondaryKey string
	keyGen       KeyGenerator
}

// NewGuardianKeyGenerator returns a new instance of guardian key generator
func NewGuardianKeyGenerator(args ArgGuardianKeyGenerator) (*guardianKeyGenerator, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &guardianKeyGenerator{
		mainKey:      args.MainKey,
		secondaryKey: args.SecondaryKey,
		keyGen:       args.KeyGen,
	}, nil
}

func checkArgs(args ArgGuardianKeyGenerator) error {
	if len(args.MainKey) == 0 {
		return fmt.Errorf("%w for main key", ErrInvalidValue)
	}
	if len(args.SecondaryKey) == 0 {
		return fmt.Errorf("%w for secondary key", ErrInvalidValue)
	}
	if check.IfNil(args.KeyGen) {
		return ErrNilKeyGenerator
	}

	return nil
}

// GenerateKeys generates two HD keys based on the provided index and the managed keys
func (generator *guardianKeyGenerator) GenerateKeys(index uint32) ([]crypto.PrivateKey, error) {
	wallet := interactors.NewWallet()
	mainPrivateKeyBytes := wallet.GetPrivateKeyFromMnemonic(data.Mnemonic(generator.mainKey), 0, index)
	mainKey, err := generator.keyGen.PrivateKeyFromByteArray(mainPrivateKeyBytes)
	if err != nil {
		return nil, err
	}

	secondaryPrivateKeyBytes := wallet.GetPrivateKeyFromMnemonic(data.Mnemonic(generator.secondaryKey), 0, index)
	secondaryKey, err := generator.keyGen.PrivateKeyFromByteArray(secondaryPrivateKeyBytes)
	if err != nil {
		return nil, err
	}

	return []crypto.PrivateKey{mainKey, secondaryKey}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (generator *guardianKeyGenerator) IsInterfaceNil() bool {
	return generator == nil
}
