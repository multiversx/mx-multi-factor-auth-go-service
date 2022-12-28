package resolver

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/mock"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/encryption/x25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	erdMocks "github.com/ElrondNetwork/elrond-sdk-erdgo/testsCommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgs(userAddress string) ArgServiceResolver {
	return ArgServiceResolver{
		Provider: &testscommon.ProviderStub{},
		Proxy:    &erdMocks.ProxyStub{},
		CredentialsHandler: &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(userAddress)
			},
		},
		KeysGenerator: &testscommon.KeysGeneratorStub{
			GenerateManagedKeyCalled: func() (crypto.PrivateKey, error) {
				return testSk, nil
			},
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
		SignatureVerifier: &testscommon.SignatureVerifierStub{},
		GuardedTxBuilder:  &testscommon.GuardedTxBuilderStub{},
		RequestTime:       time.Second,
		KeyGen:            testKeygen,
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
	testKeygen = crypto.NewKeyGenerator(ed25519.NewEd25519())
	testSk, _  = testKeygen.GeneratePair()
)

const usrAddr = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"

func TestNewServiceResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil Proxy should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.Proxy = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilProxy, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil CredentialsHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.CredentialsHandler = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilCredentialsHandler, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil KeysGenerator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.KeysGenerator = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilKeysGenerator, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil PubKeyConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.PubKeyConverter = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilPubKeyConverter, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil RegisteredUsersDB should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.RegisteredUsersDB = nil
		resolver, err := NewServiceResolver(args)
		assert.True(t, errors.Is(err, ErrNilDB))
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.Provider = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilProvider, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.Marshaller = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilMarshaller, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil TxHasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.TxHasher = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilHasher, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil SignatureVerifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.SignatureVerifier = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilSignatureVerifier, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil GuardedTxBuilder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.GuardedTxBuilder = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilGuardedTxBuilder, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("invalid request time should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.RequestTime = minRequestTime - time.Nanosecond
		resolver, err := NewServiceResolver(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "RequestTime"))
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil KeyGen should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.KeyGen = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilKeyGenerator, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("GenerateManagedKey fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateManagedKeyCalled: func() (crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, expectedErr, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		resolver, err := NewServiceResolver(createMockArgs(usrAddr))
		assert.Nil(t, err)
		assert.False(t, check.IfNil(resolver))
	})
}

func TestServiceResolver_GetGuardianAddress(t *testing.T) {
	t.Parallel()

	// First time registering
	t.Run("first time registering, but allocate index", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but computeDataAndSave fails while encrypting", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateManagedKeyCalled: func() (crypto.PrivateKey, error) {
				return &erdMocks.PrivateKeyStub{
					GeneratePublicCalled: func() crypto.PublicKey {
						return &erdMocks.PublicKeyStub{
							ToByteArrayCalled: func() ([]byte, error) {
								return nil, expectedErr
							},
						}
					},
				}, nil
			},
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
	t.Run("first time registering, but computeDataAndSave fails during second marshal", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return errors.New("missing key")
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		counter := 0
		args.Marshaller = &erdMocks.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 1 {
					return nil, expectedErr
				}
				return json.Marshal(obj)
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but computeDataAndSave fails while saving to db", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
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
		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, first Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{}
		args.Marshaller = &erdMocks.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, decrypt fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		marshallerMock := &erdMocks.MarshalizerMock{}
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateManagedKeyCalled: func() (crypto.PrivateKey, error) {
				return &erdMocks.PrivateKeyStub{
					ToByteArrayCalled: func() ([]byte, error) {
						return nil, expectedErr
					},
				}, nil
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, marshallerMock, *providedUserInfo), nil
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, second Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		marshallerMock := &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, marshallerMock, *providedUserInfo), nil
			},
		}
		counter := 0
		args.Marshaller = &erdMocks.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				counter++
				if counter > 1 {
					return expectedErr
				}
				return marshallerMock.Unmarshal(obj, buff)
			},
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return marshallerMock.Marshal(obj)
			},
		}
		userAddress, _ := data.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, first not usable yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}

		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

	addr, err := data.NewAddressFromBech32String(usrAddr)
	assert.Nil(t, err)
	t.Run("validate credentials fails - verify error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkRegisterUserResults(t, args, requests.RegistrationPayload{}, expectedErr, nil, emptyAddress)
	})
	t.Run("validate credentials fails - get account address error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return nil, expectedErr
			},
		}
		checkRegisterUserResults(t, args, requests.RegistrationPayload{}, expectedErr, nil, emptyAddress)
	})
	t.Run("validate credentials fails - get account error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.Proxy = &erdMocks.ProxyStub{
			GetAccountCalled: func(address erdgoCore.AddressHandler) (*data.Account, error) {
				return nil, expectedErr
			},
		}
		checkRegisterUserResults(t, args, requests.RegistrationPayload{}, expectedErr, nil, emptyAddress)
	})
	t.Run("should return first guardian if none registered", func(t *testing.T) {
		t.Parallel()

		expectedQR := []byte("expected qr")
		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs(usrAddr)
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
		checkRegisterUserResults(t, args, req, nil, expectedQR, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("should return first guardian if first is registered but not usable", func(t *testing.T) {
		t.Parallel()

		expectedQR := []byte("expected qr")
		args := createMockArgs(usrAddr)
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return expectedErr
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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
		checkRegisterUserResults(t, args, req, nil, expectedQR, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("should work for first guardian and real address", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return addr, nil
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
		checkRegisterUserResults(t, args, req, nil, expectedQR, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("getGuardianAddress returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
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

		checkRegisterUserResults(t, args, req, expectedErr, nil, emptyAddress)
	})
	t.Run("RegisterUser returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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

		checkRegisterUserResults(t, args, req, expectedErr, nil, emptyAddress)
	})
	t.Run("should work for second guardian and tag provided", func(t *testing.T) {
		t.Parallel()

		providedTag := "tag"
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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
		checkRegisterUserResults(t, args, req, nil, expectedQR, string(providedUserInfoCopy.SecondGuardian.PublicKey))
	})
}

func TestServiceResolver_VerifyCode(t *testing.T) {
	t.Parallel()

	providedRequest := requests.VerificationPayload{
		Code: "secret code",
	}
	t.Run("validate credentials fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, expectedErr)
	})
	t.Run("verify code and update otp returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.Provider = &testscommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian, userCode string) error {
				return expectedErr
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, expectedErr)
	})
	t.Run("decode returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return nil, expectedErr
			},
		}
		checkVerifyCodeResults(t, args, providedRequest, expectedErr)
	})
	t.Run("update guardian state if needed fails - get user info error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
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

		args := createMockArgs(usrAddr)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			VerifyCalled: func(credentials string) error {
				return expectedErr
			},
		}
		checkSignTransactionResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("tx validation fails, sender is different than credentials one", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String("erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49")
			},
		}
		checkSignTransactionResults(t, args, providedRequest, nil, ErrInvalidSender)
	})
	t.Run("tx validation fails, marshal fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
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

		args := createMockArgs(usrAddr)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
			},
		}
		checkSignTransactionResults(t, args, request, nil, ErrInvalidGuardian)
	})
	t.Run("getGuardianForTx fails, provided guardian is not usable yet", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		request := requests.SignTransaction{
			Tx: data.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfoCopy.FirstGuardian.PublicKey),
			},
		}
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, providedUserInfoCopy), nil
			},
		}
		checkSignTransactionResults(t, args, request, nil, ErrGuardianNotUsable)
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
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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
		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

		args := createMockArgs(usrAddr)
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
		args := createMockArgs(usrAddr)
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
		args := createMockArgs(usrAddr)
		args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
			GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
				return data.NewAddressFromBech32String(providedSender)
			},
		}
		checkSignMultipleTransactionsResults(t, args, request, nil, ErrInvalidSender)
	})
	t.Run("apply guardian signature fails for second tx", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
			},
		}
		counter := 0
		args.Marshaller = &erdMocks.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 4 {
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

		args := createMockArgs(usrAddr)
		args.Marshaller = &erdMocks.MarshalizerMock{}
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
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.Marshaller, *providedUserInfo), nil
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

func TestPutGet(t *testing.T) {
	t.Parallel()

	addr1, _ := data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
	addr2, _ := data.NewAddressFromBech32String("erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49")
	args := createMockArgs(usrAddr)
	localCacher := make(map[string][]byte)
	args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
		PutCalled: func(key, data []byte) error {
			localCacher[string(key)] = data
			return nil
		},
		GetCalled: func(key []byte) ([]byte, error) {
			return localCacher[string(key)], nil
		},
	}
	args.Marshaller = &erdMocks.MarshalizerMock{}
	counter := 0
	args.CredentialsHandler = &testscommon.CredentialsHandlerStub{
		GetAccountAddressCalled: func(credentials string) (erdgoCore.AddressHandler, error) {
			counter++
			if counter == 1 {
				return addr1, nil
			}
			return addr2, nil
		},
	}

	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))

	firstGuardian1 := core.GuardianInfo{
		PublicKey:  []byte("public key first 1"),
		PrivateKey: []byte("private key first 1"),
		State:      core.Usable,
	}
	secondGuardian1 := core.GuardianInfo{
		PublicKey:  []byte("public key second 1"),
		PrivateKey: []byte("private key second 1"),
		State:      core.Usable,
	}
	providedUserInfo1 := &core.UserInfo{
		Index:          1,
		FirstGuardian:  firstGuardian1,
		SecondGuardian: secondGuardian1,
	}
	err := resolver.marshalAndSave(addr1.AddressBytes(), providedUserInfo1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(localCacher))

	firstGuardian2 := core.GuardianInfo{
		PublicKey:  []byte("public key first 2"),
		PrivateKey: []byte("private key first 2"),
		State:      core.Usable,
	}
	secondGuardian2 := core.GuardianInfo{
		PublicKey:  []byte("public key second 2"),
		PrivateKey: []byte("private key second 2"),
		State:      core.Usable,
	}
	providedUserInfo2 := &core.UserInfo{
		Index:          2,
		FirstGuardian:  firstGuardian2,
		SecondGuardian: secondGuardian2,
	}

	err = resolver.marshalAndSave(addr2.AddressBytes(), providedUserInfo2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(localCacher))

	userInfo, err := resolver.getUserInfo(addr1.AddressBytes())
	assert.Nil(t, err)
	assert.Equal(t, providedUserInfo1.Index, userInfo.Index)
	assert.Equal(t, providedUserInfo1.FirstGuardian, userInfo.FirstGuardian)
	assert.Equal(t, providedUserInfo1.SecondGuardian, userInfo.SecondGuardian)

	userInfo, err = resolver.getUserInfo(addr2.AddressBytes())
	assert.Nil(t, err)
	assert.Equal(t, providedUserInfo2.Index, userInfo.Index)
	assert.Equal(t, providedUserInfo2.FirstGuardian, userInfo.FirstGuardian)
	assert.Equal(t, providedUserInfo2.SecondGuardian, userInfo.SecondGuardian)
}

func checkGetGuardianAddressResults(t *testing.T, args ArgServiceResolver, userAddress erdgoCore.AddressHandler, expectedErr error, expectedAddress string) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	addr, err := resolver.getGuardianAddress(userAddress)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, expectedAddress, addr)
}

func checkRegisterUserResults(t *testing.T, args ArgServiceResolver, providedRequest requests.RegistrationPayload, expectedErr error, expectedCode []byte, expectedGuardian string) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	qrCode, guardian, err := resolver.RegisterUser(providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedCode, qrCode)
	assert.Equal(t, expectedGuardian, guardian)
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

func getEncryptedDataBuff(t *testing.T, marshaller marshal.Marshalizer, providedUserInfo core.UserInfo) []byte {
	providedUserInfoBuff, err := marshaller.Marshal(providedUserInfo)
	assert.Nil(t, err)
	providedEncryptedData := &x25519.EncryptedData{}
	err = providedEncryptedData.Encrypt(providedUserInfoBuff, testSk.GeneratePublic(), testSk)
	assert.Nil(t, err)
	providedEncryptedDataBuff, err := marshaller.Marshal(providedEncryptedData)
	assert.Nil(t, err)

	return providedEncryptedDataBuff
}
