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
		KeysGenerator: &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&erdMocks.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return providedUserInfo.FirstGuardian.PublicKey, nil
						},
						GeneratePublicCalled: func() crypto.PublicKey {
							return &erdMocks.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return providedUserInfo.FirstGuardian.PublicKey, nil
								},
							}
						},
					},
					&erdMocks.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return providedUserInfo.SecondGuardian.PrivateKey, nil
						},
						GeneratePublicCalled: func() crypto.PublicKey {
							return &erdMocks.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return providedUserInfo.SecondGuardian.PrivateKey, nil
								},
							}
						},
					},
				}, nil
			},
		},
		PubKeyConverter: &mock.PubkeyConverterStub{},
		RegisteredUsersDB: &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return errors.New("missing key")
			},
		},
		Marshaller:        &erdMocks.MarshalizerMock{},
		TxHasher:          keccak.NewKeccak(),
		SignatureVerifier: &erdMocks.SignerStub{},
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

const usrAddr = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"

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
	t.Run("nil RegisteredUsersDB should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = nil
		resolver, err := NewServiceResolver(args)
		assert.True(t, errors.Is(err, ErrNilDB))
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
	t.Run("nil TxHasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TxHasher = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilHasher, err)
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

	// First time registering
	t.Run("first time registering, but allocate index", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			AllocateIndexCalled: func(address []byte) (uint32, error) {
				return 0, expectedErr
			},
			HasCalled: func(key []byte) error {
				return expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but keys generator fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
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
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
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
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
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
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
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
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
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
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return errors.New("missing key")
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedAddress)
	})

	// Second time registering
	t.Run("second time registering, get from db returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{}
		args.Marshaller = &erdMocks.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{}
		args.Marshaller = &erdMocks.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, first not usable yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, first usable but second not yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, string(providedUserInfoCopy.SecondGuardian.PublicKey))
	})
	t.Run("second time registering, both usable but proxy returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.Proxy = &erdMocks.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*api.GuardianData, error) {
				return nil, expectedErr
			},
		}

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, both missing from chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, both missing(nil data from proxy) from chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, both on chain and first pending should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, both on chain and first active should return second", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, string(providedUserInfo.SecondGuardian.PublicKey))
	})
	t.Run("second time registering, only first on chain should return second", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, string(providedUserInfo.SecondGuardian.PublicKey))
	})
	t.Run("second time registering, only second on chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("second time registering, final put returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)

	})
}

func TestServiceResolver_RegisterUser(t *testing.T) {
	t.Parallel()

	addr, _ := data.NewAddressFromBech32String(usrAddr)
	t.Run("should return first guardian if none registered", func(t *testing.T) {
		t.Parallel()

		expectedQR := []byte("expected qr")
		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return expectedErr
			},
		}
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, tag, guardian string) ([]byte, error) {
				assert.Equal(t, addr.AddressAsBech32String(), account)
				assert.Equal(t, addr.Pretty(), tag)
				return expectedQR, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		req := requests.RegistrationPayload{}
		checkRegisterUserResults(t, args, addr, req, nil, expectedQR, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("should return first guardian if first is registered but not usable", func(t *testing.T) {
		t.Parallel()

		expectedQR := []byte("expected qr")
		args := createMockArgs()
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsableYet
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return expectedErr
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, tag, guardian string) ([]byte, error) {
				assert.Equal(t, addr.AddressAsBech32String(), account)
				assert.Equal(t, addr.Pretty(), tag)
				return expectedQR, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		req := requests.RegistrationPayload{}
		checkRegisterUserResults(t, args, addr, req, nil, expectedQR, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("should work for first guardian and real address", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		req := requests.RegistrationPayload{}
		expectedQR := []byte("expected qr")
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, tag, guardian string) ([]byte, error) {
				assert.Equal(t, addr.AddressAsBech32String(), account)
				assert.Equal(t, addr.Pretty(), tag)
				return expectedQR, nil
			},
		}
		checkRegisterUserResults(t, args, addr, req, nil, expectedQR, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("getGuardianAddress returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return expectedErr
			},
		}
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}

		req := requests.RegistrationPayload{}
		expectedQR := []byte("expected qr")
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, tag, guardian string) ([]byte, error) {
				assert.Equal(t, addr.AddressAsBech32String(), account)
				assert.Equal(t, addr.Pretty(), tag)
				return expectedQR, nil
			},
		}

		checkRegisterUserResults(t, args, addr, req, expectedErr, nil, emptyAddress)
	})
	t.Run("RegisterUser returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsableYet
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		req := requests.RegistrationPayload{}
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, tag, guardian string) ([]byte, error) {
				assert.Equal(t, addr.AddressAsBech32String(), account)
				assert.Equal(t, addr.Pretty(), tag)
				return nil, expectedErr
			},
		}

		checkRegisterUserResults(t, args, addr, req, expectedErr, nil, emptyAddress)
	})
	t.Run("should work for second guardian and tag provided", func(t *testing.T) {
		t.Parallel()

		providedTag := "tag"
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		req := requests.RegistrationPayload{
			Tag: providedTag,
		}
		expectedQR := []byte("expected qr")
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, tag, guardian string) ([]byte, error) {
				assert.Equal(t, providedTag, tag)
				assert.NotEqual(t, account, tag)
				return expectedQR, nil
			},
		}
		checkRegisterUserResults(t, args, addr, req, nil, expectedQR, string(providedUserInfoCopy.SecondGuardian.PublicKey))
	})
}

func TestServiceResolver_VerifyCode(t *testing.T) {
	t.Parallel()

	providedRequest := requests.VerificationPayload{
		Code: "secret code",
	}
	userAddress, _ := data.NewAddressFromBech32String(usrAddr)
	t.Run("verify code and update otp returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Provider = &testscommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian, userCode string) error {
				return expectedErr
			},
		}
		checkVerifyCodeResults(t, args, userAddress, providedRequest, expectedErr)
	})
	t.Run("decode returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkVerifyCodeResults(t, args, userAddress, providedRequest, expectedErr)
	})
	t.Run("update guardian state if needed fails - get user info error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkVerifyCodeResults(t, args, userAddress, providedRequest, expectedErr)
	})
	t.Run("update guardian state if needed fails - trying to update first guardian but already usable", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.Usable
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.FirstGuardian.PublicKey, nil
			},
		}
		checkVerifyCodeResults(t, args, userAddress, providedRequest, ErrInvalidGuardianState)
	})
	t.Run("update guardian state if needed fails - trying to update second guardian but already usable", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.Usable
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.SecondGuardian.PublicKey, nil
			},
		}
		checkVerifyCodeResults(t, args, userAddress, providedRequest, ErrInvalidGuardianState)
	})
	t.Run("should work for first guardian", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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
		checkVerifyCodeResults(t, args, userAddress, providedRequest, nil)
		assert.True(t, wasCalled)
	})
	t.Run("should work for second guardian", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsableYet
		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfoCopy)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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
		checkVerifyCodeResults(t, args, userAddress, providedRequest, nil)
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
	userAddress, _ := data.NewAddressFromBech32String(providedSender)
	t.Run("tx validation fails, sender is different than credentials one", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		anotherSender := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
		anotherAddress, _ := data.NewAddressFromBech32String(anotherSender)
		checkSignTransactionResults(t, args, anotherAddress, providedRequest, nil, ErrInvalidSender)
	})
	t.Run("tx validation fails, marshal fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkSignTransactionResults(t, args, userAddress, providedRequest, nil, expectedErr)
	})
	t.Run("tx validation fails, signature verification fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SignatureVerifier = &erdMocks.SignerStub{
			VerifyByteSliceCalled: func(msg []byte, publicKey crypto.PublicKey, sig []byte) error {
				return expectedErr
			},
		}
		checkSignTransactionResults(t, args, userAddress, providedRequest, nil, expectedErr)
	})
	t.Run("code validation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Provider = &testscommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian, userCode string) error {
				return expectedErr
			},
		}
		checkSignTransactionResults(t, args, userAddress, providedRequest, nil, expectedErr)
	})
	t.Run("tx request validation fails, getUserInfo error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkSignTransactionResults(t, args, userAddress, providedRequest, nil, expectedErr)
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		checkSignTransactionResults(t, args, userAddress, request, nil, ErrInvalidGuardian)
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
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		checkSignTransactionResults(t, args, userAddress, request, nil, ErrGuardianNotYetUsable)
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
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian erdgoCore.CryptoComponentsHolder, tx *data.Transaction) error {
				return expectedErr
			},
		}

		resolver, _ := NewServiceResolver(args)
		resolver.newCryptoComponentsHolderHandler = func(keyGen crypto.KeyGenerator, skBytes []byte) (erdgoCore.CryptoComponentsHolder, error) {
			return nil, nil
		}

		assert.False(t, check.IfNil(resolver))
		txHash, err := resolver.SignTransaction(userAddress, request)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHash)
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
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		resolver, _ := NewServiceResolver(args)
		resolver.newCryptoComponentsHolderHandler = func(keyGen crypto.KeyGenerator, skBytes []byte) (erdgoCore.CryptoComponentsHolder, error) {
			return nil, nil
		}

		assert.False(t, check.IfNil(resolver))
		txHash, err := resolver.SignTransaction(userAddress, request)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHash)
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
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian erdgoCore.CryptoComponentsHolder, tx *data.Transaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		args.Marshaller = &erdMocks.MarshalizerMock{}
		txCopy := request.Tx
		txCopy.GuardianSignature = providedGuardianSignature
		finalTxBuff, _ := args.Marshaller.Marshal(&txCopy)

		resolver, _ := NewServiceResolver(args)
		resolver.newCryptoComponentsHolderHandler = func(keyGen crypto.KeyGenerator, skBytes []byte) (erdgoCore.CryptoComponentsHolder, error) {
			return nil, nil
		}

		assert.False(t, check.IfNil(resolver))
		txHash, err := resolver.SignTransaction(userAddress, request)
		assert.Nil(t, err)
		assert.Equal(t, finalTxBuff, txHash)
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
	userAddress, _ := data.NewAddressFromBech32String(usrAddr)
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
		checkSignMultipleTransactionsResults(t, args, userAddress, request, nil, ErrGuardianMismatch)
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
		checkSignMultipleTransactionsResults(t, args, userAddress, request, nil, ErrInvalidSender)
	})
	t.Run("apply guardian signature fails for second tx", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		counter := 0
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian erdgoCore.CryptoComponentsHolder, tx *data.Transaction) error {
				counter++
				if counter > 1 {
					return expectedErr
				}
				return nil
			},
		}
		resolver, _ := NewServiceResolver(args)
		resolver.newCryptoComponentsHolderHandler = func(keyGen crypto.KeyGenerator, skBytes []byte) (erdgoCore.CryptoComponentsHolder, error) {
			return nil, nil
		}

		assert.False(t, check.IfNil(resolver))
		txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHashes)
	})
	t.Run("marshal fails for second tx", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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
		resolver, _ := NewServiceResolver(args)
		resolver.newCryptoComponentsHolderHandler = func(keyGen crypto.KeyGenerator, skBytes []byte) (erdgoCore.CryptoComponentsHolder, error) {
			return nil, nil
		}

		assert.False(t, check.IfNil(resolver))
		txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHashes)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &erdMocks.MarshalizerMock{}
		providedUserInfoBuff, _ := args.Marshaller.Marshal(providedUserInfo)
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return providedUserInfoBuff, nil
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian erdgoCore.CryptoComponentsHolder, tx *data.Transaction) error {
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
		resolver, _ := NewServiceResolver(args)
		resolver.newCryptoComponentsHolderHandler = func(keyGen crypto.KeyGenerator, skBytes []byte) (erdgoCore.CryptoComponentsHolder, error) {
			return nil, nil
		}

		assert.False(t, check.IfNil(resolver))
		txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
		assert.Equal(t, expectedResponse, txHashes)
		assert.Nil(t, err)
	})
}

func checkGetGuardianAddressResults(t *testing.T, args ArgServiceResolver, userAddress erdgoCore.AddressHandler, expectedErr error, expectedAddress string) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	addr, err := resolver.getGuardianAddress(userAddress)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, expectedAddress, addr)
}

func checkRegisterUserResults(t *testing.T, args ArgServiceResolver, userAddress erdgoCore.AddressHandler, request requests.RegistrationPayload, expectedErr error, expectedCode []byte, expectedGuardian string) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	qrCode, guardian, err := resolver.RegisterUser(userAddress, request)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedCode, qrCode)
	assert.Equal(t, expectedGuardian, guardian)
}

func checkVerifyCodeResults(t *testing.T, args ArgServiceResolver, userAddress erdgoCore.AddressHandler, providedRequest requests.VerificationPayload, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	err := resolver.VerifyCode(userAddress, providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
}

func checkSignTransactionResults(t *testing.T, args ArgServiceResolver, userAddress erdgoCore.AddressHandler, providedRequest requests.SignTransaction, expectedHash []byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	txHash, err := resolver.SignTransaction(userAddress, providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedHash, txHash)
}

func checkSignMultipleTransactionsResults(t *testing.T, args ArgServiceResolver, userAddress erdgoCore.AddressHandler, providedRequest requests.SignMultipleTransactions, expectedHashes [][]byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedHashes, txHashes)
}
