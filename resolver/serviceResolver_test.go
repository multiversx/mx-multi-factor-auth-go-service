package resolver

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/encryption/x25519"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/testsCommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgs() ArgServiceResolver {
	return ArgServiceResolver{
		Provider: &testscommon.ProviderStub{},
		Proxy:    &testsCommon.ProxyStub{},
		KeysGenerator: &testscommon.KeysGeneratorStub{
			GenerateManagedKeyCalled: func() (crypto.PrivateKey, error) {
				return testSk, nil
			},
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&testsCommon.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return providedUserInfo.FirstGuardian.PublicKey, nil
						},
						GeneratePublicCalled: func() crypto.PublicKey {
							return &testsCommon.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return providedUserInfo.FirstGuardian.PublicKey, nil
								},
							}
						},
					},
					&testsCommon.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return providedUserInfo.SecondGuardian.PrivateKey, nil
						},
						GeneratePublicCalled: func() crypto.PublicKey {
							return &testsCommon.PublicKeyStub{
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
		UserDataMarshaller:            &testsCommon.MarshalizerMock{},
		EncryptionMarshaller:          &testsCommon.MarshalizerMock{},
		TxMarshaller:                  &testsCommon.MarshalizerMock{},
		TxHasher:                      keccak.NewKeccak(),
		SignatureVerifier:             &testsCommon.SignerStub{},
		GuardedTxBuilder:              &testscommon.GuardedTxBuilderStub{},
		RequestTime:                   time.Second,
		KeyGen:                        testKeygen,
		CryptoComponentsHolderFactory: &testscommon.CryptoComponentsHolderFactoryStub{},
		SkipTxUserSigVerify:           false,
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
	testKeygen = signing.NewKeyGenerator(ed25519.NewEd25519())
	testSk, _  = testKeygen.GeneratePair()
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
	t.Run("nil userDataMarshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.UserDataMarshaller = nil
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), ErrNilMarshaller.Error()))
		assert.True(t, strings.Contains(err.Error(), "userData marshaller"))
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil encryptionMarshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.EncryptionMarshaller = nil
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), ErrNilMarshaller.Error()))
		assert.True(t, strings.Contains(err.Error(), "encryption marshaller"))
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil txMarshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TxMarshaller = nil
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), ErrNilMarshaller.Error()))
		assert.True(t, strings.Contains(err.Error(), "tx marshaller"))
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
	t.Run("nil KeyGen should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeyGen = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilKeyGenerator, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("nil CryptoComponentsHolderFactory should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CryptoComponentsHolderFactory = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilCryptoComponentsHolderFactory, err)
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("GenerateManagedKey fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
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
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
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
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for first private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&testsCommon.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return nil, expectedErr
						},
					},
					&testsCommon.PrivateKeyStub{},
				}, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for first public key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&testsCommon.PrivateKeyStub{
						GeneratePublicCalled: func() crypto.PublicKey {
							return &testsCommon.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return nil, expectedErr
								},
							}
						},
					},
					&testsCommon.PrivateKeyStub{},
				}, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for second private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&testsCommon.PrivateKeyStub{},
					&testsCommon.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return nil, expectedErr
						},
					},
				}, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for second public key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&testsCommon.PrivateKeyStub{},
					&testsCommon.PrivateKeyStub{
						GeneratePublicCalled: func() crypto.PublicKey {
							return &testsCommon.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return nil, expectedErr
								},
							}
						},
					},
				}, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but computeDataAndSave fails on Marshal", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.UserDataMarshaller = &testsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but computeDataAndSave fails while encrypting", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateManagedKeyCalled: func() (crypto.PrivateKey, error) {
				return &testsCommon.PrivateKeyStub{
					GeneratePublicCalled: func() crypto.PublicKey {
						return &testsCommon.PublicKeyStub{
							ToByteArrayCalled: func() ([]byte, error) {
								return nil, expectedErr
							},
						}
					},
				}, nil
			},
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&testsCommon.PrivateKeyStub{},
					&testsCommon.PrivateKeyStub{},
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
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but computeDataAndSave fails during second marshal", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return errors.New("missing key")
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		counter := 0
		args.UserDataMarshaller = &testsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 1 {
					return nil, expectedErr
				}
				return json.Marshal(obj)
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering, but computeDataAndSave fails while saving to db", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return errors.New("missing key")
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("first time registering should work", func(t *testing.T) {
		t.Parallel()

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, createMockArgs(), userAddress, nil, providedUserInfo.FirstGuardian.PublicKey)
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
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, first Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{}
		args.EncryptionMarshaller = &testsCommon.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, decrypt fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		marshallerMock := &testsCommon.MarshalizerMock{}
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateManagedKeyCalled: func() (crypto.PrivateKey, error) {
				return &testsCommon.PrivateKeyStub{
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
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, second Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		marshallerMock := &testsCommon.MarshalizerMock{}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, marshallerMock, *providedUserInfo), nil
			},
		}
		counter := 0
		args.EncryptionMarshaller = &testsCommon.MarshalizerStub{
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
		args.UserDataMarshaller = &testsCommon.MarshalizerStub{
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
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, first not usable yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				assert.Equal(t, providedUserInfoCopy.FirstGuardian.PublicKey, pkBytes)
				return string(pkBytes)
			},
		}

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfoCopy.FirstGuardian.PublicKey)
	})
	t.Run("second time registering, first usable but second not yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				assert.Equal(t, providedUserInfoCopy.SecondGuardian.PublicKey, pkBytes)
				return string(pkBytes)
			},
		}

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfoCopy.SecondGuardian.PublicKey)
	})
	t.Run("second time registering, both usable but proxy returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		args.Proxy = &testsCommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address sdkCore.AddressHandler) (*api.GuardianData, error) {
				return nil, expectedErr
			},
		}

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
	t.Run("second time registering, both missing from chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		args.Proxy = &testsCommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address sdkCore.AddressHandler) (*api.GuardianData, error) {
				return &api.GuardianData{}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.FirstGuardian.PublicKey)
	})
	t.Run("second time registering, both missing(nil data from proxy) from chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		args.Proxy = &testsCommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address sdkCore.AddressHandler) (*api.GuardianData, error) {
				return nil, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.FirstGuardian.PublicKey)
	})
	t.Run("second time registering, both on chain and first pending should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		args.Proxy = &testsCommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address sdkCore.AddressHandler) (*api.GuardianData, error) {
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

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.FirstGuardian.PublicKey)
	})
	t.Run("second time registering, both on chain and first active should return second", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		args.Proxy = &testsCommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address sdkCore.AddressHandler) (*api.GuardianData, error) {
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

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.SecondGuardian.PublicKey)
	})
	t.Run("second time registering, only first on chain should return second", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		args.Proxy = &testsCommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address sdkCore.AddressHandler) (*api.GuardianData, error) {
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

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.SecondGuardian.PublicKey)
	})
	t.Run("second time registering, only second on chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		args.Proxy = &testsCommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address sdkCore.AddressHandler) (*api.GuardianData, error) {
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

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.FirstGuardian.PublicKey)
	})
	t.Run("second time registering, final put returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		args.Proxy = &testsCommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address sdkCore.AddressHandler) (*api.GuardianData, error) {
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

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, emptyAddress)
	})
}

func TestServiceResolver_RegisterUser(t *testing.T) {
	t.Parallel()

	addr, _ := sdkData.NewAddressFromBech32String(usrAddr)
	t.Run("invalid address should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Proxy = &testsCommon.ProxyStub{
			GetAccountCalled: func(address sdkCore.AddressHandler) (*sdkData.Account, error) {
				return nil, expectedErr
			},
		}
		req := requests.RegistrationPayload{}
		checkRegisterUserResults(t, args, addr, req, expectedErr, nil, "")
	})
	t.Run("should return first guardian if none registered", func(t *testing.T) {
		t.Parallel()

		expectedQR := []byte("expected qr")
		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return expectedErr
			},
		}
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, guardian []byte, tag string) ([]byte, error) {
				assert.Equal(t, addr.AddressBytes(), account)
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
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return expectedErr
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, guardian []byte, tag string) ([]byte, error) {
				assert.Equal(t, addr.AddressBytes(), account)
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
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
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
			RegisterUserCalled: func(account, guardian []byte, tag string) ([]byte, error) {
				assert.Equal(t, addr.AddressBytes(), account)
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
			RegisterUserCalled: func(account, guardian []byte, tag string) ([]byte, error) {
				assert.Equal(t, addr.AddressBytes(), account)
				assert.Equal(t, addr.Pretty(), tag)
				return expectedQR, nil
			},
		}

		checkRegisterUserResults(t, args, addr, req, expectedErr, nil, "")
	})
	t.Run("RegisterUser returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		req := requests.RegistrationPayload{}
		args.Provider = &testscommon.ProviderStub{
			RegisterUserCalled: func(account, guardian []byte, tag string) ([]byte, error) {
				assert.Equal(t, addr.AddressBytes(), account)
				assert.Equal(t, addr.Pretty(), tag)
				return nil, expectedErr
			},
		}

		checkRegisterUserResults(t, args, addr, req, expectedErr, nil, "")
	})
	t.Run("should work for second guardian and tag provided", func(t *testing.T) {
		t.Parallel()

		providedTag := "tag"
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
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
			RegisterUserCalled: func(account, guardian []byte, tag string) ([]byte, error) {
				assert.Equal(t, providedTag, tag)
				assert.NotEqual(t, account, tag)
				return expectedQR, nil
			},
		}

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkRegisterUserResults(t, args, userAddress, req, nil, expectedQR, string(providedUserInfoCopy.SecondGuardian.PublicKey))
	})
}

func TestServiceResolver_VerifyCode(t *testing.T) {
	t.Parallel()

	providedRequest := requests.VerificationPayload{
		Code: "secret code",
	}
	t.Run("verify code and update otp returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Provider = &testscommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian []byte, tag string) error {
				return expectedErr
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
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
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
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
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, expectedErr)
	})
	t.Run("update guardian state if needed fails - trying to update first guardian but already usable", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.Usable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.FirstGuardian.PublicKey, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, ErrInvalidGuardianState)
	})
	t.Run("update guardian state if needed fails - trying to update second guardian but already usable", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.Usable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.SecondGuardian.PublicKey, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, ErrInvalidGuardianState)
	})
	t.Run("should work for first guardian", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.FirstGuardian.PublicKey, nil
			},
		}
		wasCalled := false
		args.Provider = &testscommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian []byte, userCode string) error {
				assert.Equal(t, providedRequest.Code, userCode)
				wasCalled = true
				return nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, nil)
		assert.True(t, wasCalled)
	})
	t.Run("should work for second guardian", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.SecondGuardian.PublicKey, nil
			},
		}
		wasCalled := false
		args.Provider = &testscommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian []byte, userCode string) error {
				assert.Equal(t, providedRequest.Code, userCode)
				wasCalled = true
				return nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, nil)
		assert.True(t, wasCalled)
	})
}

func TestServiceResolver_SignTransaction(t *testing.T) {
	t.Parallel()

	providedSender := "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	providedRequest := requests.SignTransaction{
		Tx: sdkData.Transaction{
			SndAddr:   providedSender,
			Signature: hex.EncodeToString([]byte("signature")),
		},
	}
	userAddress, _ := sdkData.NewAddressFromBech32String(providedSender)
	t.Run("tx validation fails, sender is different than credentials one", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		anotherSender := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
		anotherAddress, _ := sdkData.NewAddressFromBech32String(anotherSender)
		signTransactionAndCheckResults(t, args, anotherAddress, providedRequest, nil, ErrInvalidSender)
	})
	t.Run("tx validation fails, marshal fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TxMarshaller = &testsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		signTransactionAndCheckResults(t, args, userAddress, providedRequest, nil, expectedErr)
	})
	t.Run("tx validation fails, signature verification fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SignatureVerifier = &testsCommon.SignerStub{
			VerifyByteSliceCalled: func(msg []byte, publicKey crypto.PublicKey, sig []byte) error {
				return expectedErr
			},
		}
		signTransactionAndCheckResults(t, args, userAddress, providedRequest, nil, expectedErr)
	})
	t.Run("code validation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Provider = &testscommon.ProviderStub{
			ValidateCodeCalled: func(account, guardian []byte, userCode string) error {
				return expectedErr
			},
		}
		signTransactionAndCheckResults(t, args, userAddress, providedRequest, nil, expectedErr)
	})
	t.Run("tx request validation fails, getUserInfo error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		signTransactionAndCheckResults(t, args, userAddress, providedRequest, nil, expectedErr)
	})
	t.Run("getGuardianForTx fails, unknown guardian", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: sdkData.Transaction{
				SndAddr:      providedSender,
				GuardianAddr: "unknown guardian",
				Signature:    hex.EncodeToString([]byte("signature")),
			},
		}
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		signTransactionAndCheckResults(t, args, userAddress, request, nil, ErrInvalidGuardian)
	})
	t.Run("getGuardianForTx fails, provided guardian is not usable yet", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		request := requests.SignTransaction{
			Tx: sdkData.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfoCopy.FirstGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, providedUserInfoCopy), nil
			},
		}
		signTransactionAndCheckResults(t, args, userAddress, request, nil, ErrGuardianNotUsable)
	})
	t.Run("apply guardian signature fails", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: sdkData.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *sdkData.Transaction) error {
				return expectedErr
			},
		}

		resolver, _ := NewServiceResolver(args)

		assert.False(t, check.IfNil(resolver))
		txHash, err := resolver.SignTransaction(userAddress, request)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHash)
	})
	t.Run("marshal fails for final tx", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: sdkData.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		counter := 0
		args.TxMarshaller = &testsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 3 {
					return nil, expectedErr
				}
				return testsCommon.MarshalizerMock{}.Marshal(obj)
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return testsCommon.MarshalizerMock{}.Unmarshal(obj, buff)
			},
		}
		args.EncryptionMarshaller = &testsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 3 {
					return nil, expectedErr
				}
				return testsCommon.MarshalizerMock{}.Marshal(obj)
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return testsCommon.MarshalizerMock{}.Unmarshal(obj, buff)
			},
		}

		resolver, _ := NewServiceResolver(args)

		assert.False(t, check.IfNil(resolver))
		txHash, err := resolver.SignTransaction(userAddress, request)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHash)
	})
	t.Run("should work with sig verification", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: sdkData.Transaction{
				SndAddr:      providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *sdkData.Transaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		txCopy := request.Tx
		txCopy.GuardianSignature = providedGuardianSignature
		finalTxBuff, _ := args.TxMarshaller.Marshal(&txCopy)

		resolver, _ := NewServiceResolver(args)

		assert.False(t, check.IfNil(resolver))
		txHash, err := resolver.SignTransaction(userAddress, request)
		assert.Nil(t, err)
		assert.Equal(t, finalTxBuff, txHash)
	})
	t.Run("should work without sig verification", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Tx: sdkData.Transaction{
				SndAddr:      providedSender,
				Signature:    "",
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.SkipTxUserSigVerify = true
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *sdkData.Transaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		txCopy := request.Tx
		txCopy.GuardianSignature = providedGuardianSignature
		finalTxBuff, _ := args.TxMarshaller.Marshal(&txCopy)

		resolver, _ := NewServiceResolver(args)

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
		Txs: []sdkData.Transaction{
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
	userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
	t.Run("tx validation fails, different guardians on txs", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Txs: []sdkData.Transaction{
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
		signMultipleTransactionsAndCheckResults(t, args, userAddress, request, nil, ErrGuardianMismatch)
	})
	t.Run("tx validation fails, different senders on txs", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Txs: []sdkData.Transaction{
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
		signMultipleTransactionsAndCheckResults(t, args, userAddress, request, nil, ErrInvalidSender)
	})
	t.Run("apply guardian signature fails for second tx", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		counter := 0
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *sdkData.Transaction) error {
				counter++
				if counter > 1 {
					return expectedErr
				}
				return nil
			},
		}
		resolver, _ := NewServiceResolver(args)

		assert.False(t, check.IfNil(resolver))
		txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHashes)
	})
	t.Run("marshal fails for second tx", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		counter := 0
		args.EncryptionMarshaller = &testsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 4 {
					return nil, expectedErr
				}
				return testsCommon.MarshalizerMock{}.Marshal(obj)
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return testsCommon.MarshalizerMock{}.Unmarshal(obj, buff)
			},
		}
		args.TxMarshaller = &testsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 4 {
					return nil, expectedErr
				}
				return testsCommon.MarshalizerMock{}.Marshal(obj)
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return testsCommon.MarshalizerMock{}.Unmarshal(obj, buff)
			},
		}
		resolver, _ := NewServiceResolver(args)

		assert.False(t, check.IfNil(resolver))
		txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHashes)
	})
	t.Run("should work with sig verification", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *sdkData.Transaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		expectedResponse := make([][]byte, len(providedRequest.Txs))
		for idx := range providedRequest.Txs {
			txCopy := providedRequest.Txs[idx]
			txCopy.GuardianSignature = providedGuardianSignature
			expectedResponse[idx], _ = args.TxMarshaller.Marshal(txCopy)
		}
		resolver, _ := NewServiceResolver(args)

		assert.False(t, check.IfNil(resolver))
		txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
		assert.Equal(t, expectedResponse, txHashes)
		assert.Nil(t, err)
	})
	t.Run("should work without sig verification", func(t *testing.T) {
		t.Parallel()

		providedRequest := requests.SignMultipleTransactions{
			Txs: []sdkData.Transaction{
				{
					SndAddr:      providedSender,
					Signature:    "",
					GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
				}, {
					SndAddr:      providedSender,
					Signature:    "",
					GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
				},
			},
		}

		args := createMockArgs()
		args.SkipTxUserSigVerify = true
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return getEncryptedDataBuff(t, args.EncryptionMarshaller, *providedUserInfo), nil
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *sdkData.Transaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		expectedResponse := make([][]byte, len(providedRequest.Txs))
		for idx := range providedRequest.Txs {
			txCopy := providedRequest.Txs[idx]
			txCopy.GuardianSignature = providedGuardianSignature
			expectedResponse[idx], _ = args.TxMarshaller.Marshal(txCopy)
		}
		resolver, _ := NewServiceResolver(args)

		assert.False(t, check.IfNil(resolver))
		txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
		assert.Equal(t, expectedResponse, txHashes)
		assert.Nil(t, err)
	})
}

func TestServiceResolver_RegisteredUsers(t *testing.T) {
	t.Parallel()

	providedCount := uint32(150)
	args := createMockArgs()
	args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
		CountCalled: func() (uint32, error) {
			return providedCount, nil
		},
	}
	resolver, _ := NewServiceResolver(args)

	assert.False(t, check.IfNil(resolver))
	count, err := resolver.RegisteredUsers()
	assert.Nil(t, err)
	assert.Equal(t, providedCount, count)
}

func TestPutGet(t *testing.T) {
	t.Parallel()

	addr1, _ := sdkData.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
	addr2, _ := sdkData.NewAddressFromBech32String("erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49")
	args := createMockArgs()
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

func checkGetGuardianAddressResults(t *testing.T, args ArgServiceResolver, userAddress sdkCore.AddressHandler, expectedErr error, expectedAddress []byte) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	addr, err := resolver.getGuardianAddress(userAddress)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, expectedAddress, addr)
}

func checkRegisterUserResults(t *testing.T, args ArgServiceResolver, userAddress sdkCore.AddressHandler, request requests.RegistrationPayload, expectedErr error, expectedCode []byte, expectedGuardian string) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	qrCode, guardian, err := resolver.RegisterUser(userAddress, request)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedCode, qrCode)
	assert.Equal(t, expectedGuardian, guardian)
}

func checkVerifyCodeResults(t *testing.T, args ArgServiceResolver, userAddress sdkCore.AddressHandler, providedRequest requests.VerificationPayload, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	err := resolver.VerifyCode(userAddress, providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
}

func signTransactionAndCheckResults(t *testing.T, args ArgServiceResolver, userAddress sdkCore.AddressHandler, providedRequest requests.SignTransaction, expectedHash []byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	txHash, err := resolver.SignTransaction(userAddress, providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedHash, txHash)
}

func signMultipleTransactionsAndCheckResults(t *testing.T, args ArgServiceResolver, userAddress sdkCore.AddressHandler, providedRequest requests.SignMultipleTransactions, expectedHashes [][]byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.False(t, check.IfNil(resolver))
	txHashes, err := resolver.SignMultipleTransactions(userAddress, providedRequest)
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
