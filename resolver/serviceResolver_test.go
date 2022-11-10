package resolver

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/mock"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	erdMocks "github.com/ElrondNetwork/elrond-sdk-erdgo/testsCommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgServiceResolver {
	return ArgServiceResolver{
		Proxy: &erdMocks.ProxyStub{},
		CredentialsHandler: &testsCommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
			},
		},
		IndexHandler:      &testsCommon.IndexHandlerStub{},
		KeysGenerator:     &testsCommon.KeysGeneratorStub{},
		PubKeyConverter:   &mock.PubkeyConverterStub{},
		RegisteredUsersDB: testsCommon.NewStorerMock(),
		Provider:          &testsCommon.ProviderStub{},
		Marshaller:        &erdMocks.MarshalizerMock{},
		RequestTime:       time.Second,
	}
}

var (
	expectedErr      = errors.New("expected err")
	providedUserInfo = &core.UserInfo{
		Index: 7,
		FirstGuardian: core.GuardianInfo{
			PublicKey:  []byte("first public"),
			PrivateKey: []byte("first private"),
			State:      core.Usable,
		},
		SecondGuardian: core.GuardianInfo{
			PublicKey:  []byte("second public"),
			PrivateKey: []byte("second private"),
			State:      core.Usable,
		},
	}
)

func TestNewServiceResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil Proxy should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Proxy = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilProxy, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil CredentialsHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilCredentialsHandler, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil IndexHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.IndexHandler = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilIndexHandler, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil KeysGenerator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilKeysGenerator, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil PubKeyConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilPubKeyConverter, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil Storer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilStorer, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Provider = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilProvider, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilMarshaller, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("invalid request time should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RequestTime = minRequestTime - time.Nanosecond
		resolver, err := NewServiceResolver(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "RequestTime"))
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		resolver, err := NewServiceResolver(createMockArgs())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(resolver))
	})
}

func TestServiceResolver_GetGuardianAddress(t *testing.T) {
	t.Parallel()

	t.Run("validate credentials fails - verify error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testsCommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("validate credentials fails - get account address error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testsCommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return nil, expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("validate credentials fails - get account error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Proxy = &erdMocks.ProxyStub{
			GetAccountCalled: func(address erdgoCore.AddressHandler) (*data.Account, error) {
				return nil, expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})

	// First time registering
	t.Run("first time registering, but keys generator fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testsCommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for first private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testsCommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return nil, expectedErr
						},
					},
					&erdMocks.PrivateKeyStub{},
				}, nil
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for first public key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testsCommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{
						GeneratePublicCalled: func() crypto.PublicKey {
							return &erdMocks.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return nil, expectedErr
								},
							}
						},
					},
					&erdMocks.PrivateKeyStub{},
				}, nil
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for second private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testsCommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{},
					&erdMocks.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return nil, expectedErr
						},
					},
				}, nil
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for second public key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testsCommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{},
					&erdMocks.PrivateKeyStub{
						GeneratePublicCalled: func() crypto.PublicKey {
							return &erdMocks.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return nil, expectedErr
								},
							}
						},
					},
				}, nil
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but computeDataAndSave fails on Marshal", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testsCommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{},
					&erdMocks.PrivateKeyStub{},
				}, nil
			},
		}
		args.Marshaller = &erdMocks.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but computeDataAndSave fails while saving to db", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testsCommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{},
					&erdMocks.PrivateKeyStub{},
				}, nil
			},
		}
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			HasCalled: func(key []byte) error {
				return errors.New("missing key")
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("first time registering should work", func(t *testing.T) {
		t.Parallel()

		providedAddress := "provided address"
		args := createMockArgs()
		args.KeysGenerator = &testsCommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{},
					&erdMocks.PrivateKeyStub{},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return providedAddress
			},
		}

		checkGetGuardianAddressResults(t, args, nil, providedAddress)
	})

	// Second time registering
	t.Run("second time registering, get from db returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("second time registering, Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testsCommon.StorerStub{}
		args.Marshaller = &erdMocks.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("second time registering, Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testsCommon.StorerStub{}
		args.Marshaller = &erdMocks.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("second time registering, first not usable yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				assert.Equal(t, providedUserInfoCopy.FirstGuardian.PublicKey, pkBytes)
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, nil, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, first usable but second not yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				assert.Equal(t, providedUserInfoCopy.SecondGuardian.PublicKey, pkBytes)
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, nil, string(providedUserInfoCopy.SecondGuardian.PublicKey))
	})
	t.Run("second time registering, both usable but proxy returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return nil, expectedErr
			},
		}

		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("second time registering, both missing from chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
			PutCalled: func(key, data []byte) error {
				userInfoCopy := *providedUserInfo
				userInfoCopy.FirstGuardian.State = core.NotUsableYet
				userInfoCopy.SecondGuardian.State = core.NotUsableYet
				buff, _ := args.Marshaller.Marshal(&userInfoCopy)
				assert.Equal(t, string(buff), string(data))
				return nil
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: "active guardian",
					},
					PendingGuardian: &data.Guardian{
						Address: "pending guardian",
					},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, nil, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, both on chain and first pending should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
			PutCalled: func(key, data []byte) error {
				userInfoCopy := *providedUserInfo
				userInfoCopy.FirstGuardian.State = core.NotUsableYet
				buff, _ := args.Marshaller.Marshal(&userInfoCopy)
				assert.Equal(t, string(buff), string(data))
				return nil
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: string(providedUserInfo.SecondGuardian.PublicKey),
					},
					PendingGuardian: &data.Guardian{
						Address: string(providedUserInfo.FirstGuardian.PublicKey),
					},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, nil, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, both on chain and first active should return second", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
			PutCalled: func(key, data []byte) error {
				userInfoCopy := *providedUserInfo
				userInfoCopy.SecondGuardian.State = core.NotUsableYet
				buff, _ := args.Marshaller.Marshal(&userInfoCopy)
				assert.Equal(t, string(buff), string(data))
				return nil
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: string(providedUserInfo.FirstGuardian.PublicKey),
					},
					PendingGuardian: &data.Guardian{
						Address: string(providedUserInfo.SecondGuardian.PublicKey),
					},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, nil, string(providedUserInfo.SecondGuardian.PublicKey))
	})
	t.Run("second time registering, only first on chain should return second", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
			PutCalled: func(key, data []byte) error {
				userInfoCopy := *providedUserInfo
				userInfoCopy.SecondGuardian.State = core.NotUsableYet
				buff, _ := args.Marshaller.Marshal(&userInfoCopy)
				assert.Equal(t, string(buff), string(data))
				return nil
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: string(providedUserInfo.FirstGuardian.PublicKey),
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, nil, string(providedUserInfo.SecondGuardian.PublicKey))
	})
	t.Run("second time registering, only second on chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
			PutCalled: func(key, data []byte) error {
				userInfoCopy := *providedUserInfo
				userInfoCopy.FirstGuardian.State = core.NotUsableYet
				buff, _ := args.Marshaller.Marshal(&userInfoCopy)
				assert.Equal(t, string(buff), string(data))
				return nil
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: string(providedUserInfo.SecondGuardian.PublicKey),
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, nil, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, final put returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian:  &data.Guardian{},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)

	})
}

func TestServiceResolver_RegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("validate credentials fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testsCommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkRegisterUserResults(t, args, requests.RegistrationPayload{}, expectedErr, nil)
	})
	t.Run("validate guardian fails - getUserInfo error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkRegisterUserResults(t, args, requests.RegistrationPayload{}, expectedErr, nil)
	})
	t.Run("validate guardian fails - encode returns none of the 2 guardians", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return "different guardian"
			},
		}
		checkRegisterUserResults(t, args, requests.RegistrationPayload{}, ErrInvalidGuardian, nil)
	})
	t.Run("should error for first with usable state", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		numCalls := 0
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				numCalls++
				if numCalls == 1 {
					return string(providedUserInfoCopy.FirstGuardian.PublicKey)
				}
				return string(providedUserInfoCopy.SecondGuardian.PublicKey)
			},
		}
		req := requests.RegistrationPayload{
			Guardian: string(providedUserInfoCopy.FirstGuardian.PublicKey),
		}
		checkRegisterUserResults(t, args, req, ErrInvalidGuardian, nil)
	})
	t.Run("should work for first", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		numCalls := 0
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				numCalls++
				if numCalls == 1 {
					return string(providedUserInfoCopy.FirstGuardian.PublicKey)
				}
				return string(providedUserInfoCopy.SecondGuardian.PublicKey)
			},
		}
		req := requests.RegistrationPayload{
			Guardian: string(providedUserInfoCopy.FirstGuardian.PublicKey),
		}
		expectedQR := []byte("expected qr")
		args.Provider = &testsCommon.ProviderStub{
			RegisterUserCalled: func(account, guardian string) ([]byte, error) {
				return expectedQR, nil
			},
		}
		checkRegisterUserResults(t, args, req, nil, expectedQR)
	})
	t.Run("should error for second with usable state", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		numCalls := 0
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				numCalls++
				if numCalls == 1 {
					return string(providedUserInfoCopy.FirstGuardian.PublicKey)
				}
				return string(providedUserInfoCopy.SecondGuardian.PublicKey)
			},
		}
		req := requests.RegistrationPayload{
			Guardian: string(providedUserInfoCopy.SecondGuardian.PublicKey),
		}
		checkRegisterUserResults(t, args, req, ErrInvalidGuardian, nil)
	})
	t.Run("should work for second", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		numCalls := 0
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				numCalls++
				if numCalls == 1 {
					return string(providedUserInfoCopy.FirstGuardian.PublicKey)
				}
				return string(providedUserInfoCopy.SecondGuardian.PublicKey)
			},
		}
		req := requests.RegistrationPayload{
			Guardian: string(providedUserInfoCopy.SecondGuardian.PublicKey),
		}
		expectedQR := []byte("expected qr")
		args.Provider = &testsCommon.ProviderStub{
			RegisterUserCalled: func(account, guardian string) ([]byte, error) {
				return expectedQR, nil
			},
		}
		checkRegisterUserResults(t, args, req, nil, expectedQR)
	})
}

func TestServiceResolver_VerifyCode(t *testing.T) {
	t.Parallel()

	providedRequest := requests.VerificationPayload{
		Code: "secret code",
	}
	t.Run("validate credentials fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testsCommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, expectedErr)
	})
	t.Run("verify code and update otp returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Provider = &testsCommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian, userCode string) error {
				return expectedErr
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, expectedErr)
	})
	t.Run("decode returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, expectedErr)
	})
	t.Run("update guardian state if needed fails - get user info error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, expectedErr)
	})
	t.Run("update guardian state if needed fails - trying to update first guardian but already usable", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.Usable
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.FirstGuardian.PublicKey, nil
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, ErrInvalidGuardianState)
	})
	t.Run("update guardian state if needed fails - trying to update second guardian but already usable", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.Usable
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.SecondGuardian.PublicKey, nil
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, ErrInvalidGuardianState)
	})
	t.Run("should work for first guardian", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.FirstGuardian.PublicKey, nil
			},
		}
		wasCalled := false
		args.Provider = &testsCommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian, userCode string) error {
				assert.Equal(t, providedRequest.Code, userCode)
				wasCalled = true
				return nil
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, nil)
		assert.True(t, wasCalled)
	})
	t.Run("should work for second guardian", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testsCommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.SecondGuardian.PublicKey, nil
			},
		}
		wasCalled := false
		args.Provider = &testsCommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian, userCode string) error {
				assert.Equal(t, providedRequest.Code, userCode)
				wasCalled = true
				return nil
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, nil)
		assert.True(t, wasCalled)
	})
}

func checkGetGuardianAddressResults(t *testing.T, args ArgServiceResolver, expectedErr error, expectedAddress string) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	addr, err := resolver.GetGuardianAddress(requests.GetGuardianAddress{})
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, expectedAddress, addr)
}

func checkRegisterUserResults(t *testing.T, args ArgServiceResolver, providedRequest requests.RegistrationPayload, expectedErr error, expectedCode []byte) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	qrCode, err := resolver.RegisterUser(providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedCode, qrCode)
}

func checkVerifyCodeResults(t *testing.T, args ArgServiceResolver, providedRequest requests.VerificationPayload, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	err := resolver.VerifyCode(providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
}
