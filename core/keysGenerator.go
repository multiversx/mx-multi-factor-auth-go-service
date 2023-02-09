package core

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/interactors"
)

const managedKeyIndex = 0

// ArgGuardianKeyGenerator is the DTO used to create a new instance of guardian key generator
type ArgGuardianKeyGenerator struct {
	Mnemonic data.Mnemonic
	KeyGen   crypto.KeyGenerator
}

type guardianKeyGenerator struct {
	mnemonic data.Mnemonic
	keyGen   crypto.KeyGenerator
}

// NewGuardianKeyGenerator returns a new instance of guardian key generator
func NewGuardianKeyGenerator(args ArgGuardianKeyGenerator) (*guardianKeyGenerator, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &guardianKeyGenerator{
		mnemonic: args.Mnemonic,
		keyGen:   args.KeyGen,
	}, nil
}

func checkArgs(args ArgGuardianKeyGenerator) error {
	if len(args.Mnemonic) == 0 {
		return fmt.Errorf("%w for mnemonic", ErrInvalidValue)
	}
	if check.IfNil(args.KeyGen) {
		return ErrNilKeyGenerator
	}

	return nil
}

// GenerateManagedKey generates one HD key based on a constant index, which will only be used by the service
func (generator *guardianKeyGenerator) GenerateManagedKey() (crypto.PrivateKey, error) {
	wallet := interactors.NewWallet()
	privateKeyBytes := wallet.GetPrivateKeyFromMnemonic(generator.mnemonic, 0, managedKeyIndex)
	return generator.keyGen.PrivateKeyFromByteArray(privateKeyBytes)
}

// GenerateKeys generates two HD keys based on the provided index and the managed keys
func (generator *guardianKeyGenerator) GenerateKeys(index uint32) ([]crypto.PrivateKey, error) {
	if index == managedKeyIndex {
		return nil, fmt.Errorf("%w for index %d", ErrInvalidValue, index)
	}

	wallet := interactors.NewWallet()
	firstIndex := index
	firstPrivateKeyBytes := wallet.GetPrivateKeyFromMnemonic(generator.mnemonic, 0, firstIndex)
	firstKey, err := generator.keyGen.PrivateKeyFromByteArray(firstPrivateKeyBytes)
	if err != nil {
		return nil, err
	}

	secondIndex := firstIndex + 1
	secondPrivateKeyBytes := wallet.GetPrivateKeyFromMnemonic(generator.mnemonic, 0, secondIndex)
	secondKey, err := generator.keyGen.PrivateKeyFromByteArray(secondPrivateKeyBytes)
	if err != nil {
		return nil, err
	}

	return []crypto.PrivateKey{firstKey, secondKey}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (generator *guardianKeyGenerator) IsInterfaceNil() bool {
	return generator == nil
}
