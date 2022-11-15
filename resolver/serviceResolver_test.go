package resolver

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/mock"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	erdMocks "github.com/ElrondNetwork/elrond-sdk-erdgo/testsCommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgServiceResolver {
	return ArgServiceResolver{
		Provider: &testscommon.ProviderStub{},
		Proxy:    &erdMocks.ProxyStub{},
		CredentialsHandler: &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
			},
		},
		IndexHandler:      &testscommon.IndexHandlerStub{},
		KeysGenerator:     &testscommon.KeysGeneratorStub{},
		PubKeyConverter:   &mock.PubkeyConverterStub{},
		RegisteredUsersDB: testscommon.NewStorerMock(),
		Marshaller:        &erdMocks.MarshalizerMock{},
		TxHasher:          keccak.NewKeccak(),
		SignatureVerifier: &testscommon.SignatureVerifierStub{},
		GuardedTxBuilder:  &testscommon.GuardedTxBuilderStub{},
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
	t.Run("nil SignatureVerifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SignatureVerifier = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilSignatureVerifier, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil GuardedTxBuilder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.GuardedTxBuilder = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilGuardedTxBuilder, err)
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
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("validate credentials fails - get account address error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
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
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for first private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
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
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
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
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
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
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
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
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
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
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{},
					&erdMocks.PrivateKeyStub{},
				}, nil
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkGetGuardianAddressResults(t, args, expectedErr, emptyAddress)
	})
	t.Run("second time registering, Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.StorerStub{}
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
		args.RegisteredUsersDB = &testscommon.StorerStub{}
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
				return &api.GuardianData{}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		checkGetGuardianAddressResults(t, args, nil, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, both missing(nil data from proxy) from chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.StorerStub{
			HasCalled: func(key []byte) error {
				return nil
			},
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
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
				return nil, nil
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian: &api.Guardian{
						Address: string(providedUserInfo.SecondGuardian.PublicKey),
					},
					PendingGuardian: &api.Guardian{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian: &api.Guardian{
						Address: string(providedUserInfo.FirstGuardian.PublicKey),
					},
					PendingGuardian: &api.Guardian{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian: &api.Guardian{
						Address: string(providedUserInfo.FirstGuardian.PublicKey),
					},
					PendingGuardian: &api.Guardian{},
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian: &api.Guardian{
						Address: string(providedUserInfo.SecondGuardian.PublicKey),
					},
					PendingGuardian: &api.Guardian{},
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian:  &api.Guardian{},
					PendingGuardian: &api.Guardian{},
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
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkRegisterUserResults(t, args, requests.RegistrationPayload{}, expectedErr, nil)
	})
	t.Run("validate guardian fails - getUserInfo error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.Provider = &testscommon.ProviderStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.Provider = &testscommon.ProviderStub{
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
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, expectedErr)
	})
	t.Run("verify code and update otp returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Provider = &testscommon.ProviderStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.Provider = &testscommon.ProviderStub{
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
		args.RegisteredUsersDB = &testscommon.StorerStub{
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
		args.Provider = &testscommon.ProviderStub{
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

func TestServiceResolver_SignTransaction(t *testing.T) {
	t.Parallel()

	providedSender := "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	providedRequest := requests.SignTransaction{
		Tx: data.Transaction{
			SndAddr:   providedSender,
			Signature: hex.EncodeToString([]byte("signature")),
		},
	}
	t.Run("validate credentials fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkSignTransactionResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("tx validation fails, sender is different than credentials one", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String("erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49")
			},
		}
		checkSignTransactionResults(t, args, providedRequest, nil, ErrInvalidSender)
	})
	t.Run("tx validation fails, marshal fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.Marshaller = &erdMocks.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkSignTransactionResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("tx validation fails, signature verification fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.SignatureVerifier = &testscommon.SignatureVerifierStub{
			VerifyCalled: func(pk []byte, msg []byte, skBytes []byte) error {
				return expectedErr
			},
		}
		checkSignTransactionResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("code validation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.Provider = &testscommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian, userCode string) error {
				return expectedErr
			},
		}
		checkSignTransactionResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("tx request validation fails, getUserInfo error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkSignTransactionResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("getGuardianForTx fails, unknown guardian", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: data.Transaction{
				SndAddr:      providedSender,
				GuardianAddr: "unknown guardian",
				Signature:    hex.EncodeToString([]byte("signature")),
			},
		}
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		checkSignTransactionResults(t, args, request, nil, ErrInvalidGuardian)
	})
	t.Run("getGuardianForTx fails, provided guardian is not usable yet", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsableYet
		request := requests.SignTransaction{
			Tx: data.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfoCopy.FirstGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		checkSignTransactionResults(t, args, request, nil, ErrGuardianNotYetUsable)
	})
	t.Run("apply guardian signature fails", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: data.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(skGuardianBytes []byte, tx *data.Transaction) error {
				return expectedErr
			},
		}
		checkSignTransactionResults(t, args, request, nil, expectedErr)
	})
	t.Run("marshal fails for final tx", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: data.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		counter := 0
		args.Marshaller = &erdMocks.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 1 {
					return nil, expectedErr
				}
				return erdMocks.MarshalizerMock{}.Marshal(obj)
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return erdMocks.MarshalizerMock{}.Unmarshal(obj, buff)
			},
		}
		checkSignTransactionResults(t, args, request, nil, expectedErr)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: data.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(skGuardianBytes []byte, tx *data.Transaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		args.Marshaller = &erdMocks.MarshalizerMock{}
		txCopy := request.Tx
		txCopy.GuardianSignature = providedGuardianSignature
		finalTxBuff, _ := args.Marshaller.Marshal(&txCopy)
		checkSignTransactionResults(t, args, request, finalTxBuff, nil)
	})
}

func TestServiceResolver_SignMultipleTransactions(t *testing.T) {
	t.Parallel()

	providedSender := "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	providedRequest := requests.SignMultipleTransactions{
		Txs: []data.Transaction{
			{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
			}, {
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
			},
		},
	}
	t.Run("validate credentials fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkSignMultipleTransactionsResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("tx validation fails, different guardians on txs", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Txs: []data.Transaction{
				{
					SndAddr:      providedSender,
					Signature:    hex.EncodeToString([]byte("signature")),
					GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
				}, {
					SndAddr:      providedSender,
					Signature:    hex.EncodeToString([]byte("signature")),
					GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
				},
			},
		}
		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		checkSignMultipleTransactionsResults(t, args, request, nil, ErrGuardianMismatch)
	})
	t.Run("tx validation fails, different senders on txs", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Txs: []data.Transaction{
				{
					SndAddr:   providedSender,
					Signature: hex.EncodeToString([]byte("signature")),
				}, {
					SndAddr:   "erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49",
					Signature: hex.EncodeToString([]byte("signature")),
				},
			},
		}
		args := createMockArgs()
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		checkSignMultipleTransactionsResults(t, args, request, nil, ErrInvalidSender)
	})
	t.Run("apply guardian signature fails for second tx", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		counter := 0
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(skGuardianBytes []byte, tx *data.Transaction) error {
				counter++
				if counter > 1 {
					return expectedErr
				}
				return nil
			},
		}
		checkSignMultipleTransactionsResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("marshal fails for second tx", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		counter := 0
		args.Marshaller = &erdMocks.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 3 {
					return nil, expectedErr
				}
				return erdMocks.MarshalizerMock{}.Marshal(obj)
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return erdMocks.MarshalizerMock{}.Unmarshal(obj, buff)
			},
		}
		checkSignMultipleTransactionsResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(skGuardianBytes []byte, tx *data.Transaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		args.Marshaller = &erdMocks.MarshalizerMock{}
		expectedResponse := make([][]byte, len(providedRequest.Txs))
		for idx := range providedRequest.Txs {
			txCopy := providedRequest.Txs[idx]
			txCopy.GuardianSignature = providedGuardianSignature
			expectedResponse[idx], _ = args.Marshaller.Marshal(txCopy)
		}
		checkSignMultipleTransactionsResults(t, args, providedRequest, expectedResponse, nil)
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

func checkSignTransactionResults(t *testing.T, args ArgServiceResolver, providedRequest requests.SignTransaction, expectedHash []byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	txHash, err := resolver.SignTransaction(providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedHash, txHash)
}

func checkSignMultipleTransactionsResults(t *testing.T, args ArgServiceResolver, providedRequest requests.SignMultipleTransactions, expectedHashes [][]byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	txHashes, err := resolver.SignMultipleTransactions(providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedHashes, txHashes)
}
