package factory

import (
	"io/ioutil"

	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	factoryMarshalizer "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/data"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/encryption"
	"github.com/multiversx/mx-multi-factor-auth-go-service/resolver"
)

// CreateServiceResolver will create a new service resolver component
func CreateServiceResolver(
	configs *config.Configs,
	cryptoComponents *cryptoComponentsHolder,
	httpClientWrapper core.HttpClientWrapper,
	registeredUsersDB core.StorageWithIndex,
	twoFactorHandler handlers.TOTPHandler,
	secureOtpHandler handlers.SecureOtpHandler,
) (core.ServiceResolver, error) {
	gogoMarshaller, err := factoryMarshalizer.NewMarshalizer(factoryMarshalizer.GogoProtobuf)
	if err != nil {
		return nil, err
	}

	jsonTxMarshaller, err := factoryMarshalizer.NewMarshalizer(factoryMarshalizer.TxJsonMarshalizer)
	if err != nil {
		return nil, err
	}

	jsonMarshaller, err := factoryMarshalizer.NewMarshalizer(factoryMarshalizer.JsonMarshalizer)
	if err != nil {
		return nil, err
	}

	builder, err := builders.NewTxBuilder(cryptoComponents.Signer())
	if err != nil {
		return nil, err
	}

	mnemonic, err := ioutil.ReadFile(configs.GeneralConfig.Guardian.MnemonicFile)
	if err != nil {
		return nil, err
	}
	argsGuardianKeyGenerator := core.ArgGuardianKeyGenerator{
		Mnemonic: data.Mnemonic(mnemonic),
		KeyGen:   cryptoComponents.KeyGenerator(),
	}
	guardianKeyGenerator, err := core.NewGuardianKeyGenerator(argsGuardianKeyGenerator)
	if err != nil {
		return nil, err
	}

	cryptoComponentsHolderFactory, err := core.NewCryptoComponentsHolderFactory(cryptoComponents.KeyGenerator())
	if err != nil {
		return nil, err
	}

	managedPrivateKey, err := guardianKeyGenerator.GenerateManagedKey()
	if err != nil {
		return nil, err
	}

	encryptor, err := encryption.NewEncryptor(jsonMarshaller, cryptoComponents.KeyGenerator(), managedPrivateKey)
	if err != nil {
		return nil, err
	}

	userEncryptor, err := resolver.NewUserEncryptor(encryptor)
	if err != nil {
		return nil, err
	}

	txHasher := keccak.NewKeccak()

	argsServiceResolver := resolver.ArgServiceResolver{
		UserEncryptor:                 userEncryptor,
		TOTPHandler:                   twoFactorHandler,
		SecureOtpHandler:              secureOtpHandler,
		HttpClientWrapper:             httpClientWrapper,
		KeysGenerator:                 guardianKeyGenerator,
		PubKeyConverter:               cryptoComponents.PubkeyConverter(),
		RegisteredUsersDB:             registeredUsersDB,
		UserDataMarshaller:            gogoMarshaller,
		TxMarshaller:                  jsonTxMarshaller,
		TxHasher:                      txHasher,
		SignatureVerifier:             cryptoComponents.Signer(),
		GuardedTxBuilder:              builder,
		KeyGen:                        cryptoComponents.KeyGenerator(),
		CryptoComponentsHolderFactory: cryptoComponentsHolderFactory,
		Config:                        configs.GeneralConfig.ServiceResolver,
	}
	return resolver.NewServiceResolver(argsServiceResolver)
}
