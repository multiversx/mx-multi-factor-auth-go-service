package resolver

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	sdkTestsCommon "github.com/multiversx/mx-sdk-go/testsCommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/secureOtp"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon"
)

const usrAddr = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"

var (
	expectedErr        = errors.New("expected err")
	providedMessageAge = int64(100)
	providedUserInfo   = &core.UserInfo{
		Index: 7,
		FirstGuardian: core.GuardianInfo{
			PublicKey:  []byte("first public"),
			PrivateKey: []byte("first private"),
			State:      core.Usable,
			OTPData: core.OTPInfo{
				OTP:                     []byte("otp1"),
				LastTOTPChangeTimestamp: time.Now().Unix() - providedMessageAge,
			},
		},
		SecondGuardian: core.GuardianInfo{
			PublicKey:  []byte("second public"),
			PrivateKey: []byte("second private"),
			State:      core.Usable,
			OTPData: core.OTPInfo{
				OTP:                     []byte("otp2"),
				LastTOTPChangeTimestamp: time.Now().Unix() - providedMessageAge,
			},
		},
	}
	testKeygen      = signing.NewKeyGenerator(ed25519.NewEd25519())
	testSk, _       = testKeygen.GeneratePair()
	providedOTPInfo = &requests.OTP{
		Scheme:              "otpauth",
		Host:                "totp",
		Issuer:              "MultiversX",
		Account:             "erd1",
		Algorithm:           "SHA1",
		Counter:             0,
		Digits:              6,
		Period:              30,
		Secret:              "secret",
		TimeSinceGeneration: providedMessageAge,
	}
	providedUrl       = "otpauth://totp/MultiversX:erd1?algorithm=SHA1&counter=0&digits=6&issuer=MultiversX&period=30&secret=secret"
	defaultFirstCode  = "123456"
	defaultSecondCode = "234567"
)

func createMockArgs() ArgServiceResolver {
	return ArgServiceResolver{
		UserEncryptor: &testscommon.UserEncryptorStub{
			EncryptUserInfoCalled: func(user *core.UserInfo) (*core.UserInfo, error) {
				userCopy := *user
				return &userCopy, nil
			},
			DecryptUserInfoCalled: func(encryptedUserInfo *core.UserInfo) (*core.UserInfo, error) {
				encryptedUserInfoCopy := *encryptedUserInfo
				return &encryptedUserInfoCopy, nil
			},
		},
		TOTPHandler: &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					UrlCalled: func() (string, error) {
						return providedUrl, nil
					},
				}, nil
			},
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					UrlCalled: func() (string, error) {
						return providedUrl, nil
					},
				}, nil
			},
		},
		SecureOtpHandler: &testscommon.SecureOtpHandlerStub{
			IsVerificationAllowedAndIncreaseTrialsCalled: func(account, ip string) (*requests.OTPCodeVerifyData, error) {
				return &requests.OTPCodeVerifyData{
					RemainingTrials:             0,
					ResetAfter:                  0,
					SecurityModeRemainingTrials: 0,
					SecurityModeResetAfter:      0,
				}, nil
			},
		},
		HttpClientWrapper: &testscommon.HttpClientWrapperStub{
			GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian:  &api.Guardian{},
					PendingGuardian: &api.Guardian{},
					Guarded:         false,
				}, nil
			},
			GetAccountCalled: func(ctx context.Context, address string) (*sdkData.Account, error) {
				return &sdkData.Account{Balance: "1"}, nil
			},
		},
		KeysGenerator: &testscommon.KeysGeneratorStub{
			GenerateManagedKeyCalled: func() (crypto.PrivateKey, error) {
				return testSk, nil
			},
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&sdkTestsCommon.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return providedUserInfo.FirstGuardian.PublicKey, nil
						},
						GeneratePublicCalled: func() crypto.PublicKey {
							return &sdkTestsCommon.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return providedUserInfo.FirstGuardian.PublicKey, nil
								},
							}
						},
					},
					&sdkTestsCommon.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return providedUserInfo.SecondGuardian.PrivateKey, nil
						},
						GeneratePublicCalled: func() crypto.PublicKey {
							return &sdkTestsCommon.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return providedUserInfo.SecondGuardian.PrivateKey, nil
								},
							}
						},
					},
				}, nil
			},
		},
		PubKeyConverter: &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return []byte(humanReadable), nil
			},
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		},
		RegisteredUsersDB: &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return errors.New("missing key")
			},
		},
		UserDataMarshaller:            &sdkTestsCommon.MarshalizerMock{},
		TxMarshaller:                  &sdkTestsCommon.MarshalizerMock{},
		TxHasher:                      keccak.NewKeccak(),
		SignatureVerifier:             &sdkTestsCommon.SignerStub{},
		GuardedTxBuilder:              &testscommon.GuardedTxBuilderStub{},
		KeyGen:                        testKeygen,
		CryptoComponentsHolderFactory: &testscommon.CryptoComponentsHolderFactoryStub{},
		Config: config.ServiceResolverConfig{
			RequestTimeInSeconds:             1,
			SkipTxUserSigVerify:              false,
			MaxTransactionsAllowedForSigning: 10,
			DelayBetweenOTPWritesInSec:       minDelayBetweenOTPUpdates,
		},
	}
}

func TestNewServiceResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil UserEncryptor should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.UserEncryptor = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilUserEncryptor, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil HttpClientWrapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.HttpClientWrapper = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilHTTPClientWrapper, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil KeysGenerator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilKeysGenerator, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil PubKeyConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilPubKeyConverter, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil RegisteredUsersDB should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = nil
		resolver, err := NewServiceResolver(args)
		assert.True(t, errors.Is(err, ErrNilDB))
		assert.Nil(t, resolver)
	})
	t.Run("nil totp should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TOTPHandler = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilTOTPHandler, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil secureOtpHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SecureOtpHandler = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilSecureOtpHandler, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil userDataMarshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.UserDataMarshaller = nil
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), ErrNilMarshaller.Error()))
		assert.True(t, strings.Contains(err.Error(), "userData marshaller"))
		assert.Nil(t, resolver)
	})
	t.Run("nil txMarshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TxMarshaller = nil
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), ErrNilMarshaller.Error()))
		assert.True(t, strings.Contains(err.Error(), "tx marshaller"))
		assert.Nil(t, resolver)
	})
	t.Run("nil TxHasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TxHasher = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilHasher, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil SignatureVerifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SignatureVerifier = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilSignatureVerifier, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil GuardedTxBuilder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.GuardedTxBuilder = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilGuardedTxBuilder, err)
		assert.Nil(t, resolver)
	})
	t.Run("invalid request time should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Config.RequestTimeInSeconds = 0
		resolver, err := NewServiceResolver(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "RequestTimeInSeconds"))
		assert.Nil(t, resolver)
	})
	t.Run("nil KeyGen should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeyGen = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilKeyGenerator, err)
		assert.Nil(t, resolver)
	})
	t.Run("nil CryptoComponentsHolderFactory should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.CryptoComponentsHolderFactory = nil
		resolver, err := NewServiceResolver(args)
		assert.Equal(t, ErrNilCryptoComponentsHolderFactory, err)
		assert.Nil(t, resolver)
	})
	t.Run("invalid delay between OTP updates should fail", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Config.DelayBetweenOTPWritesInSec = 0
		resolver, err := NewServiceResolver(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.Nil(t, resolver)
	})
	t.Run("invalid max txs allowed for signing should fail", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Config.MaxTransactionsAllowedForSigning = 0
		resolver, err := NewServiceResolver(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, check.IfNil(resolver))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		resolver, err := NewServiceResolver(createMockArgs())
		assert.Nil(t, err)
		assert.NotNil(t, resolver)
	})
}

func TestServiceResolver_GetGuardianAddress(t *testing.T) {
	t.Parallel()

	// First time registering
	t.Run("first time registering (ErrKeyNotFound), but allocate index fails", func(t *testing.T) {
		t.Parallel()
		expectedDBGetErr := storage.ErrKeyNotFound
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			AllocateIndexCalled: func(address []byte) (uint32, error) {
				return 0, expectedErr
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDBGetErr
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering (ErrKeyNotFound), but keys generator fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}
		expectedDbGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDbGetErr
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering (ErrKeyNotFound), but getGuardianInfoForKey fails on ToByteArray for first private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&sdkTestsCommon.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return nil, expectedErr
						},
					},
					&sdkTestsCommon.PrivateKeyStub{},
				}, nil
			},
		}
		expectedDbGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDbGetErr
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for first public key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&sdkTestsCommon.PrivateKeyStub{
						GeneratePublicCalled: func() crypto.PublicKey {
							return &sdkTestsCommon.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return nil, expectedErr
								},
							}
						},
					},
					&sdkTestsCommon.PrivateKeyStub{},
				}, nil
			},
		}
		expectedDbGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDbGetErr
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for second private key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&sdkTestsCommon.PrivateKeyStub{},
					&sdkTestsCommon.PrivateKeyStub{
						ToByteArrayCalled: func() ([]byte, error) {
							return nil, expectedErr
						},
					},
				}, nil
			},
		}
		expectedDbGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDbGetErr
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering, but getGuardianInfoForKey fails on ToByteArray for second public key", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return []crypto.PrivateKey{
					&sdkTestsCommon.PrivateKeyStub{},
					&sdkTestsCommon.PrivateKeyStub{
						GeneratePublicCalled: func() crypto.PublicKey {
							return &sdkTestsCommon.PublicKeyStub{
								ToByteArrayCalled: func() ([]byte, error) {
									return nil, expectedErr
								},
							}
						},
					},
				}, nil
			},
		}
		expectedDbGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDbGetErr
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering, but computeNewUserDataAndSave fails on Marshal", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.UserDataMarshaller = &sdkTestsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		expectedDbGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDbGetErr
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering, but computeNewUserDataAndSave fails while encrypting", func(t *testing.T) {
		t.Parallel()

		expectedDbGetErr := storage.ErrKeyNotFound
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},

			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDbGetErr
			},
		}

		cnt := 0
		args.UserEncryptor = &testscommon.UserEncryptorStub{
			EncryptUserInfoCalled: func(userInfo *core.UserInfo) (*core.UserInfo, error) {
				if cnt == 0 {
					cnt++
					return userInfo, nil
				}

				return nil, expectedErr
			},
		}
		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering, but computeNewUserDataAndSave fails while saving to db", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		expectedDbGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDbGetErr
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("first time registering should work", func(t *testing.T) {
		t.Parallel()

		expectedDBGetErr := storage.ErrKeyNotFound
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDBGetErr
			},
		}
		otp := &testscommon.TotpStub{
			QRCalled: func() ([]byte, error) {
				return []byte("qrCode"), nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.FirstGuardian.PublicKey, otp, providedMessageAge)
	})
	// Second time registering
	t.Run("second time registering, decrypt fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian = core.GuardianInfo{
			PublicKey:  providedUserInfo.SecondGuardian.PublicKey,
			PrivateKey: providedUserInfo.SecondGuardian.PrivateKey,
			State:      core.NotUsable,
			OTPData: core.OTPInfo{
				OTP:                     nil,
				LastTOTPChangeTimestamp: 0,
			},
		}
		args.UserEncryptor = &testscommon.UserEncryptorStub{
			DecryptUserInfoCalled: func(encryptedUserInfo *core.UserInfo) (*core.UserInfo, error) {
				return nil, expectedErr
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("second time registering, first not usable yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				assert.Equal(t, providedUserInfoCopy.FirstGuardian.PublicKey, pkBytes)
				return string(pkBytes)
			},
		}
		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfoCopy.FirstGuardian.PublicKey, otp, providedMessageAge)
	})
	t.Run("second time registering, first usable but second not yet should work", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				assert.Equal(t, providedUserInfoCopy.SecondGuardian.PublicKey, pkBytes)
				return string(pkBytes)
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfoCopy.SecondGuardian.PublicKey, otp, providedMessageAge)
	})
	t.Run("second time registering, both usable but GetGuardianData returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
				return nil, expectedErr
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
	t.Run("second time registering, both missing from chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
				return &api.GuardianData{}, nil
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.FirstGuardian.PublicKey, otp, providedMessageAge)
	})
	t.Run("second time registering, both on chain and first pending should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
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

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.FirstGuardian.PublicKey, otp, providedMessageAge)
	})
	t.Run("second time registering, both on chain and first active should return second", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
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

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.SecondGuardian.PublicKey, otp, providedMessageAge)
	})
	t.Run("second time registering, only first on chain should return second", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian: &api.Guardian{
						Address: string(providedUserInfo.FirstGuardian.PublicKey),
					},
					PendingGuardian: &api.Guardian{},
				}, nil
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.SecondGuardian.PublicKey, otp, providedMessageAge)
	})
	t.Run("second time registering, only second on chain should return first", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian: &api.Guardian{
						Address: string(providedUserInfo.SecondGuardian.PublicKey),
					},
					PendingGuardian: &api.Guardian{},
				}, nil
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, nil, providedUserInfo.FirstGuardian.PublicKey, otp, providedMessageAge)
	})
	t.Run("second time registering, final put returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetGuardianDataCalled: func(ctx context.Context, address string) (*api.GuardianData, error) {
				return &api.GuardianData{
					ActiveGuardian:  &api.Guardian{},
					PendingGuardian: &api.Guardian{},
				}, nil
			},
		}

		otp := &testscommon.TotpStub{}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkGetGuardianAddressResults(t, args, userAddress, expectedErr, nil, otp, providedMessageAge)
	})
}

func TestServiceResolver_RegisterUser(t *testing.T) {
	t.Parallel()

	addr, _ := sdkData.NewAddressFromBech32String(usrAddr)
	t.Run("GetAccount (on register new account) returns error should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		expectedDBGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDBGetErr
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetAccountCalled: func(ctx context.Context, address string) (*sdkData.Account, error) {
				return nil, expectedErr
			},
		}

		req := requests.RegistrationPayload{}
		checkRegisterUserResults(t, args, addr, req, expectedErr, &requests.OTP{}, "")
	})
	t.Run("createTOTP error should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		req := requests.RegistrationPayload{}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string) (handlers.OTP, error) {
				return nil, expectedErr
			},
		}
		checkRegisterUserResults(t, args, addr, req, expectedErr, &requests.OTP{}, "")
	})
	t.Run("GetAccount returns empty balance should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		expectedDBGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDBGetErr
			},
		}
		args.HttpClientWrapper = &testscommon.HttpClientWrapperStub{
			GetAccountCalled: func(ctx context.Context, address string) (*sdkData.Account, error) {
				return &sdkData.Account{}, nil
			},
		}

		req := requests.RegistrationPayload{}
		checkRegisterUserResults(t, args, addr, req, ErrNoBalance, &requests.OTP{}, "")
	})
	t.Run("should return first guardian if none registered", func(t *testing.T) {
		t.Parallel()

		tag := "tag"
		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs()
		expectedDBGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDBGetErr
			},
		}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string) (handlers.OTP, error) {
				assert.Equal(t, tag, account)
				return &testscommon.TotpStub{
					UrlCalled: func() (string, error) {
						return providedUrl, nil
					},
				}, nil
			},
		}

		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		req := requests.RegistrationPayload{
			Tag: tag,
		}
		checkRegisterUserResults(t, args, addr, req, nil, providedOTPInfo, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("should propagate error if userData get error different than key not found", func(t *testing.T) {
		t.Parallel()

		expectedDBGetErr := errors.New("expected error")
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDBGetErr
			},
		}
		req := requests.RegistrationPayload{}
		checkRegisterUserResults(t, args, addr, req, expectedDBGetErr, &requests.OTP{}, "")
	})
	t.Run("should propagate error if userData unmarshall error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.UserDataMarshaller = &sdkTestsCommon.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return []byte("invalid data"), nil
			},
		}
		req := requests.RegistrationPayload{}
		checkRegisterUserResults(t, args, addr, req, expectedErr, &requests.OTP{}, "")
	})
	t.Run("should propagate error if userData decrypt error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.UserEncryptor = &testscommon.UserEncryptorStub{
			DecryptUserInfoCalled: func(encryptedUserInfo *core.UserInfo) (*core.UserInfo, error) {
				return nil, expectedErr
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		req := requests.RegistrationPayload{}
		checkRegisterUserResults(t, args, addr, req, expectedErr, &requests.OTP{}, "")
	})
	t.Run("should return first guardian if first is registered but not usable", func(t *testing.T) {
		t.Parallel()

		tag := ""
		args := createMockArgs()
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string) (handlers.OTP, error) {
				assert.Equal(t, addr.Pretty(), account)
				return &testscommon.TotpStub{
					UrlCalled: func() (string, error) {
						return providedUrl, nil
					},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		req := requests.RegistrationPayload{
			Tag: tag,
		}
		checkRegisterUserResults(t, args, addr, req, nil, providedOTPInfo, string(providedUserInfoCopy.FirstGuardian.PublicKey))
	})
	t.Run("should work for first guardian and real address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = testscommon.NewShardedStorageWithIndexMock()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		req := requests.RegistrationPayload{}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string) (handlers.OTP, error) {
				assert.Equal(t, addr.Pretty(), account)
				return &testscommon.TotpStub{
					UrlCalled: func() (string, error) {
						return providedUrl, nil
					},
				}, nil
			},
		}
		args.Config.DelayBetweenOTPWritesInSec = 2
		expectedOTPInfo := *providedOTPInfo
		expectedOTPInfo.TimeSinceGeneration = 0
		checkRegisterUserResults(t, args, addr, req, nil, &expectedOTPInfo, string(providedUserInfo.FirstGuardian.PublicKey))

		// register again should fail, too early
		time.Sleep(time.Second)
		expectedFailureOTPInfo := &requests.OTP{
			TimeSinceGeneration: 1,
		}
		checkRegisterUserResults(t, args, addr, req, handlers.ErrRegistrationFailed, expectedFailureOTPInfo, "")

		// wait until next otp generation allowed
		time.Sleep(time.Duration(args.Config.DelayBetweenOTPWritesInSec) * time.Second)
		checkRegisterUserResults(t, args, addr, req, nil, &expectedOTPInfo, string(providedUserInfo.FirstGuardian.PublicKey))
	})
	t.Run("registerUser returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		expectedDBGetErr := storage.ErrKeyNotFound
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedDBGetErr
			},
		}
		args.KeysGenerator = &testscommon.KeysGeneratorStub{
			GenerateKeysCalled: func(index uint32) ([]crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}

		req := requests.RegistrationPayload{}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string) (handlers.OTP, error) {
				assert.Equal(t, addr.Pretty(), account)
				return &testscommon.TotpStub{
					UrlCalled: func() (string, error) {
						return providedUrl, nil
					},
				}, nil
			},
		}

		checkRegisterUserResults(t, args, addr, req, expectedErr, &requests.OTP{}, "")
	})
	t.Run("RegisterUser returns error on Url call", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		req := requests.RegistrationPayload{}

		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string) (handlers.OTP, error) {
				assert.Equal(t, addr.Pretty(), account)
				return &testscommon.TotpStub{
					UrlCalled: func() (string, error) {
						return "", expectedErr
					},
				}, nil
			},
		}

		checkRegisterUserResults(t, args, addr, req, expectedErr, &requests.OTP{}, "")
	})
	t.Run("should work for second guardian and tag provided", func(t *testing.T) {
		t.Parallel()

		providedTag := "tag"
		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
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
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string) (handlers.OTP, error) {
				assert.Equal(t, providedTag, account)
				return &testscommon.TotpStub{
					UrlCalled: func() (string, error) {
						return providedUrl, nil
					},
				}, nil
			},
		}

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		expectedOTPInfo := *providedOTPInfo
		expectedOTPInfo.TimeSinceGeneration = 0
		checkRegisterUserResults(t, args, userAddress, req, nil, &expectedOTPInfo, string(providedUserInfoCopy.SecondGuardian.PublicKey))
	})
}

func TestServiceResolver_verifySecurityModeCode(t *testing.T) {
	t.Parallel()

	providedRequest := requests.VerificationPayload{
		Code:       "secret code",
		SecondCode: "secret code 2",
		Guardian:   string(providedUserInfo.FirstGuardian.PublicKey),
	}

	wrongCode := "wrong code"
	expectedWrongCodeErr := errors.New("wrong code expected error")
	totpHandler := &testscommon.TOTPHandlerStub{
		TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
			return &testscommon.TotpStub{
				ValidateCalled: func(userCode string) error {
					if userCode == wrongCode {
						return expectedWrongCodeErr
					}
					return nil
				},
			}, nil
		},
	}

	t.Run("positive remaining security mode trials, will not verify second code", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						require.Fail(t, "should not be called")
						return nil
					},
				}, nil
			},
		}
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, resolver)
		require.Nil(t, err)

		firstCode := "firstCode"
		secondCode := "second code"
		guardianAddr := []byte(providedRequest.Guardian)
		remainingTrials := 3
		err = resolver.verifySecurityModeCode(providedUserInfo, usrAddr, firstCode, secondCode, guardianAddr, remainingTrials)
		require.Nil(t, err)
	})
	t.Run("zero remaining security mode trials, with invalid code should return err", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TOTPHandler = totpHandler
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, resolver)
		require.Nil(t, err)

		firstCode := "123456"
		guardianAddr := []byte(providedRequest.Guardian)
		remainingTrials := 0
		err = resolver.verifySecurityModeCode(providedUserInfo, usrAddr, firstCode, wrongCode, guardianAddr, remainingTrials)
		require.ErrorIs(t, err, ErrSecondCodeInvalidInSecurityMode)
	})
	t.Run("zero remaining security mode trials, with valid code should not return error, if decrement gives error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TOTPHandler = totpHandler
		args.SecureOtpHandler = &testscommon.SecureOtpHandlerStub{
			DecrementSecurityModeFailedTrialsCalled: func(account string) error {
				return expectedErr
			},
		}
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, resolver)
		require.Nil(t, err)

		guardianAddr := []byte(providedRequest.Guardian)
		remainingTrials := 0
		err = resolver.verifySecurityModeCode(providedUserInfo, usrAddr, "code1", "code2", guardianAddr, remainingTrials)
		require.Nil(t, err)
	})
	t.Run("zero remaining security mode trials, with valid code ok", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TOTPHandler = totpHandler
		decrementCalled := 0
		args.SecureOtpHandler = &testscommon.SecureOtpHandlerStub{
			DecrementSecurityModeFailedTrialsCalled: func(account string) error {
				decrementCalled++
				return nil
			},
		}
		resolver, err := NewServiceResolver(args)

		require.NotNil(t, resolver)
		require.Nil(t, err)

		guardianAddr := []byte(providedRequest.Guardian)
		remainingTrials := 0
		err = resolver.verifySecurityModeCode(providedUserInfo, usrAddr, "code1", "code2", guardianAddr, remainingTrials)
		require.Nil(t, err)
		require.Equal(t, 1, decrementCalled)
	})

	t.Run("zero remaining security mode trials, with same second code, will fail", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						require.Fail(t, "should not be called")
						return nil
					},
				}, nil
			},
		}
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, resolver)
		require.Nil(t, err)

		firstCode := "firstCode"
		secondCode := firstCode
		guardianAddr := []byte(providedRequest.Guardian)
		remainingTrials := 0

		err = resolver.verifySecurityModeCode(providedUserInfo, usrAddr, firstCode, secondCode, guardianAddr, remainingTrials)
		require.True(t, errors.Is(err, ErrSecondCodeInvalidInSecurityMode))
	})
}

func TestServiceResolver_checkAllowanceAndVerifyCode(t *testing.T) {
	t.Parallel()

	providedRequest := requests.VerificationPayload{
		Code:       "secret code",
		SecondCode: "secret code 2",
		Guardian:   string(providedUserInfo.FirstGuardian.PublicKey),
	}

	wrongCodeExpectedErr := errors.New("wrong code expected error")
	wrongCode := "wrong code"
	totp := &testscommon.TOTPHandlerStub{
		TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
			return &testscommon.TotpStub{
				ValidateCalled: func(userCode string) error {
					if userCode == wrongCode {
						return wrongCodeExpectedErr
					}
					return nil
				},
			}, nil
		},
	}
	isVerificationAllowedOtpData := requests.OTPCodeVerifyData{
		RemainingTrials:             3,
		ResetAfter:                  10,
		SecurityModeRemainingTrials: 10,
		SecurityModeResetAfter:      100,
	}

	maxNormalModeFailures := uint64(4)
	secureOtpHandler := &testscommon.SecureOtpHandlerStub{
		IsVerificationAllowedAndIncreaseTrialsCalled: func(account string, ip string) (*requests.OTPCodeVerifyData, error) {
			return &isVerificationAllowedOtpData, nil
		},
		ResetCalled: func(account string, ip string) {},
		DecrementSecurityModeFailedTrialsCalled: func(account string) error {
			return nil
		},
		FreezeMaxFailuresCalled: func() uint64 {
			return maxNormalModeFailures
		},
	}

	t.Run("secureOTP handler returns error on verification allowed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SecureOtpHandler = &testscommon.SecureOtpHandlerStub{
			IsVerificationAllowedAndIncreaseTrialsCalled: func(account string, ip string) (*requests.OTPCodeVerifyData, error) {
				return nil, expectedErr
			},
			ResetCalled: func(account string, ip string) {},
			DecrementSecurityModeFailedTrialsCalled: func(account string) error {
				return nil
			},
			FreezeMaxFailuresCalled: func() uint64 {
				return 4
			},
		}
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, resolver)
		require.Nil(t, err)

		providedUserInfoCopy := *providedUserInfo
		otpVerifyData, err := resolver.checkAllowanceAndVerifyCode(
			&providedUserInfoCopy,
			usrAddr,
			"userIP",
			providedRequest.Code,
			providedRequest.SecondCode,
			[]byte(providedRequest.Guardian))

		require.Equal(t, expectedErr, err)
		require.Nil(t, otpVerifyData)
	})
	t.Run("first code validation failed with remaining trials for normal and security mode, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SecureOtpHandler = secureOtpHandler
		args.TOTPHandler = totp
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, resolver)
		require.Nil(t, err)

		providedUserInfoCopy := *providedUserInfo
		otpVerifyData, err := resolver.checkAllowanceAndVerifyCode(
			&providedUserInfoCopy,
			usrAddr,
			"userIP",
			wrongCode,
			providedRequest.SecondCode,
			[]byte(providedRequest.Guardian))

		require.Equal(t, wrongCodeExpectedErr, err)
		require.Equal(t, isVerificationAllowedOtpData, *otpVerifyData)
	})
	t.Run("first code ok, wrong second code but with remaining trials, should not error (second code will not be verified) ", func(t *testing.T) {
		t.Parallel()

		validateCalled := 0
		args := createMockArgs()
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						validateCalled++
						if userCode == wrongCode {
							return wrongCodeExpectedErr
						}
						return nil
					},
				}, nil
			},
		}

		secureOtpHandlerCopy := *secureOtpHandler
		resetCalled := 0
		secureOtpHandlerCopy.ResetCalled = func(account string, ip string) {
			resetCalled++
		}
		decrementCalled := 0
		secureOtpHandlerCopy.DecrementSecurityModeFailedTrialsCalled = func(account string) error {
			decrementCalled++
			return nil
		}

		args.SecureOtpHandler = &secureOtpHandlerCopy
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, resolver)
		require.Nil(t, err)

		providedUserInfoCopy := *providedUserInfo
		otpVerifyData, err := resolver.checkAllowanceAndVerifyCode(
			&providedUserInfoCopy,
			usrAddr,
			"userIP",
			providedRequest.Code,
			wrongCode,
			[]byte(providedRequest.Guardian))

		expectedData := requests.OTPCodeVerifyData{
			RemainingTrials:             int(maxNormalModeFailures),
			ResetAfter:                  0,
			SecurityModeRemainingTrials: isVerificationAllowedOtpData.SecurityModeRemainingTrials,
			SecurityModeResetAfter:      isVerificationAllowedOtpData.SecurityModeResetAfter}
		require.Nil(t, err)
		require.Equal(t, 1, resetCalled)
		require.Equal(t, 1, decrementCalled)
		// called only for first code
		require.Equal(t, 1, validateCalled)
		require.Equal(t, expectedData, *otpVerifyData)
	})
	t.Run("first code ok, wrong second code with no remaining trials for security mode should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		isVerificationAllowedOtpDataCopy := isVerificationAllowedOtpData
		// to enforce verification of the second code
		isVerificationAllowedOtpDataCopy.SecurityModeRemainingTrials = 0

		secureOtpHandlerCopy := *secureOtpHandler
		resetCalled := false
		secureOtpHandlerCopy.IsVerificationAllowedAndIncreaseTrialsCalled = func(account string, ip string) (*requests.OTPCodeVerifyData, error) {
			return &isVerificationAllowedOtpDataCopy, nil
		}
		secureOtpHandlerCopy.ResetCalled = func(account string, ip string) {
			resetCalled = true
		}
		secureOtpHandlerCopy.DecrementSecurityModeFailedTrialsCalled = func(account string) error {
			require.Fail(t, "should not have been called")
			return nil
		}

		args.SecureOtpHandler = &secureOtpHandlerCopy
		args.TOTPHandler = totp
		resolver, err := NewServiceResolver(args)
		require.NotNil(t, resolver)
		require.Nil(t, err)

		providedUserInfoCopy := *providedUserInfo
		otpVerifyData, err := resolver.checkAllowanceAndVerifyCode(
			&providedUserInfoCopy,
			usrAddr,
			"userIP",
			providedRequest.Code,
			wrongCode,
			[]byte(providedRequest.Guardian))

		expectedData := requests.OTPCodeVerifyData{
			RemainingTrials:             int(maxNormalModeFailures),
			ResetAfter:                  0,
			SecurityModeRemainingTrials: 0,
			SecurityModeResetAfter:      isVerificationAllowedOtpDataCopy.SecurityModeResetAfter}
		require.ErrorIs(t, err, ErrSecondCodeInvalidInSecurityMode, err)
		require.True(t, resetCalled)
		require.Equal(t, expectedData, *otpVerifyData)
	})
}

func TestServiceResolver_VerifyCode(t *testing.T) {
	t.Parallel()

	providedRequest := requests.VerificationPayload{
		Code:       "secret code",
		SecondCode: "secret code 2",
		Guardian:   string(providedUserInfo.FirstGuardian.PublicKey),
	}
	t.Run("verify code and update otp returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		providedUserInfoCopy := *providedUserInfo
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}

		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						return expectedErr
					},
				}, nil
			},
		}

		isVerificationAllowedCalled := false
		args.SecureOtpHandler = &testscommon.SecureOtpHandlerStub{
			IsVerificationAllowedAndIncreaseTrialsCalled: func(account, ip string) (*requests.OTPCodeVerifyData, error) {
				isVerificationAllowedCalled = true
				return nil, nil
			},
			ResetCalled: func(account string, ip string) {
				require.Fail(t, "should have not been called")
			},
		}

		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, expectedErr)
		require.True(t, isVerificationAllowedCalled)
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

	t.Run("secureOtpHandler verification failed, should error", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs()
		args.SecureOtpHandler = &testscommon.SecureOtpHandlerStub{
			IsVerificationAllowedAndIncreaseTrialsCalled: func(account string, ip string) (*requests.OTPCodeVerifyData, error) {
				return nil, expectedErr
			},
		}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				assert.Fail(t, "should not have called this")
				return nil, nil
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
			PutCalled: func(key, data []byte) error {
				require.Error(t, errors.New("should not have been called"))
				return nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)

		resolver, _ := NewServiceResolver(args)
		otpVerifyCodeData, err := resolver.VerifyCode(userAddress, "userIp", providedRequest)
		assert.Equal(t, expectedErr, err)
		require.Nil(t, otpVerifyCodeData)
	})

	t.Run("secureOtpHandler verification is not allowed should error", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		args := createMockArgs()
		args.SecureOtpHandler = &testscommon.SecureOtpHandlerStub{
			IsVerificationAllowedAndIncreaseTrialsCalled: func(account string, ip string) (*requests.OTPCodeVerifyData, error) {
				return &requests.OTPCodeVerifyData{
					RemainingTrials: 2,
					ResetAfter:      10,
				}, core.ErrTooManyFailedAttempts
			},
		}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				assert.Fail(t, "should not have called this")
				return nil, nil
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
			PutCalled: func(key, data []byte) error {
				require.Error(t, errors.New("should not have been called"))
				return nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)

		resolver, _ := NewServiceResolver(args)
		otpVerifyCodeData, err := resolver.VerifyCode(userAddress, "userIp", providedRequest)
		assert.True(t, errors.Is(err, core.ErrTooManyFailedAttempts))
		assert.Equal(t, 2, otpVerifyCodeData.RemainingTrials)
		assert.Equal(t, 10, otpVerifyCodeData.ResetAfter)
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
	t.Run("verify code successful but first guardian already usable - save not called", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.Usable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
			PutCalled: func(key, data []byte) error {
				require.Error(t, errors.New("should not have been called"))
				return nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.FirstGuardian.PublicKey, nil
			},
		}
		wasResetCalled := false
		args.SecureOtpHandler = &testscommon.SecureOtpHandlerStub{
			ResetCalled: func(account string, ip string) {
				wasResetCalled = true
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, nil)
		require.True(t, wasResetCalled)
	})
	t.Run("verify code successful but second guardian already usable - save not called", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.Usable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
			PutCalled: func(key, data []byte) error {
				require.Error(t, errors.New("should not have been called"))
				return nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.SecondGuardian.PublicKey, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, nil)
	})
	t.Run("save fails should error", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
			PutCalled: func(key, data []byte) error {
				require.Error(t, errors.New("should not have been called"))
				return expectedErr
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, expectedErr)
	})
	t.Run("should work for first guardian", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		args := createMockArgs()
		putCalled := false
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
			PutCalled: func(key, data []byte) error {
				putCalled = true
				return nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.FirstGuardian.PublicKey, nil
			},
		}
		numCalled := 0
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						switch numCalled {
						case 0:
							assert.Equal(t, providedRequest.Code, userCode)
						case 1:
							assert.Equal(t, providedRequest.SecondCode, userCode)
						}
						numCalled++
						return nil
					},
				}, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, nil)
		require.Equal(t, 2, numCalled)
		require.True(t, putCalled)
	})
	t.Run("should work for second guardian", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.SecondGuardian.State = core.NotUsable
		args := createMockArgs()
		putCalled := false
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
			PutCalled: func(key, data []byte) error {
				putCalled = true
				return nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return providedUserInfo.SecondGuardian.PublicKey, nil
			},
		}
		numCalls := 0
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						switch numCalls {
						case 0:
							assert.Equal(t, providedRequest.Code, userCode)
						case 1:
							assert.Equal(t, providedRequest.SecondCode, userCode)
						}
						numCalls++
						return nil
					},
				}, nil
			},
		}
		userAddress, _ := sdkData.NewAddressFromBech32String(usrAddr)
		checkVerifyCodeResults(t, args, userAddress, providedRequest, nil)
		require.Equal(t, 2, numCalls)
		require.True(t, putCalled)
	})
}

func TestServiceResolver_SignTransaction(t *testing.T) {
	t.Parallel()

	providedSender := "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	providedRequest := requests.SignTransaction{
		Code:       defaultFirstCode,
		SecondCode: defaultSecondCode,
		Tx: transaction.FrontendTransaction{
			Sender:       providedSender,
			Signature:    hex.EncodeToString([]byte("signature")),
			GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
		},
	}
	t.Run("tx validation fails, marshal fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TxMarshaller = &sdkTestsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		signTransactionAndCheckResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("tx validation fails, signature verification fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SignatureVerifier = &sdkTestsCommon.SignerStub{
			VerifyByteSliceCalled: func(msg []byte, publicKey crypto.PublicKey, sig []byte) error {
				return expectedErr
			},
		}
		signTransactionAndCheckResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("code validation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return []byte(humanReadable), nil
			},
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		providedRequestCopy := providedRequest
		providedRequestCopy.Tx.GuardianAddr = string(providedUserInfo.FirstGuardian.PublicKey)
		providedUserInfoCopy := *providedUserInfo
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						return expectedErr
					},
				}, nil
			},
		}

		signTransactionAndCheckResults(t, args, providedRequestCopy, nil, expectedErr)
	})
	t.Run("tx request validation fails, getUserInfo error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		signTransactionAndCheckResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("getGuardianForTx fails, unknown guardian", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Tx: transaction.FrontendTransaction{
				Sender:       providedSender,
				GuardianAddr: "unknown guardian",
				Signature:    hex.EncodeToString([]byte("signature")),
			},
		}
		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return []byte(humanReadable), nil
			},
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		signTransactionAndCheckResults(t, args, request, nil, ErrInvalidGuardian)
	})
	t.Run("getGuardianForTx fails, provided guardian is not usable yet", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		request := requests.SignTransaction{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Tx: transaction.FrontendTransaction{
				Sender:       providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfoCopy.FirstGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return []byte(humanReadable), nil
			},
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		signTransactionAndCheckResults(t, args, request, nil, ErrGuardianNotUsable)
	})
	t.Run("cryptoComponentsHolderFactory creation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.CryptoComponentsHolderFactory = &testscommon.CryptoComponentsHolderFactoryStub{
			CreateCalled: func(privateKeyBytes []byte) (sdkCore.CryptoComponentsHolder, error) {
				return nil, expectedErr
			},
		}
		signTransactionAndCheckResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("apply guardian signature fails", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Tx: transaction.FrontendTransaction{
				Sender:       providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *transaction.FrontendTransaction) error {
				return expectedErr
			},
		}

		resolver, _ := NewServiceResolver(args)

		assert.NotNil(t, resolver)
		txHash, _, err := resolver.SignTransaction("userIp", request)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHash)
	})
	t.Run("marshal fails for final tx", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Tx: transaction.FrontendTransaction{
				Sender:       providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		counter := 0
		args.TxMarshaller = &sdkTestsCommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				counter++
				if counter > 1 {
					return nil, expectedErr
				}
				return sdkTestsCommon.MarshalizerMock{}.Marshal(obj)
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return sdkTestsCommon.MarshalizerMock{}.Unmarshal(obj, buff)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}

		resolver, _ := NewServiceResolver(args)
		assert.NotNil(t, resolver)
		txHash, _, err := resolver.SignTransaction("userIp", request)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHash)
	})
	t.Run("should work with sig verification", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Tx: transaction.FrontendTransaction{
				Sender:       providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		const providedGuardianSignature = "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *transaction.FrontendTransaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		txCopy := request.Tx
		txCopy.GuardianSignature = providedGuardianSignature
		finalTxBuff, _ := args.TxMarshaller.Marshal(&txCopy)

		resolver, _ := NewServiceResolver(args)

		assert.NotNil(t, resolver)
		txHash, _, err := resolver.SignTransaction("userIp", request)
		assert.Nil(t, err)
		assert.Equal(t, finalTxBuff, txHash)
	})
	t.Run("should work without sig verification", func(t *testing.T) {
		t.Parallel()

		request := requests.SignTransaction{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Tx: transaction.FrontendTransaction{
				Sender:       providedSender,
				Signature:    "",
				GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
			},
		}
		args := createMockArgs()
		args.Config.SkipTxUserSigVerify = true
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *transaction.FrontendTransaction) error {
				tx.GuardianSignature = providedGuardianSignature
				return nil
			},
		}
		txCopy := request.Tx
		txCopy.GuardianSignature = providedGuardianSignature
		finalTxBuff, _ := args.TxMarshaller.Marshal(&txCopy)

		resolver, _ := NewServiceResolver(args)

		assert.NotNil(t, resolver)
		txHash, _, err := resolver.SignTransaction("userIp", request)
		assert.Nil(t, err)
		assert.Equal(t, finalTxBuff, txHash)
	})
}

func TestServiceResolver_SignMessage(t *testing.T) {
	t.Parallel()

	providedSender := "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	providedRequest := requests.SignMessage{
		Code:         defaultFirstCode,
		SecondCode:   defaultSecondCode,
		Message:      "message",
		UserAddr:     providedSender,
		GuardianAddr: providedSender,
	}
	t.Run("code validation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return []byte(humanReadable), nil
			},
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		providedRequestCopy := providedRequest
		providedRequestCopy.GuardianAddr = string(providedUserInfo.FirstGuardian.PublicKey)
		providedUserInfoCopy := *providedUserInfo
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						return expectedErr
					},
				}, nil
			},
		}

		signMessageAndCheckResults(t, args, providedRequestCopy, nil, expectedErr)
	})
	t.Run("sign message validation fails, getUserInfo error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		signMessageAndCheckResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("getGuardianForTx fails, unknown guardian", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMessage{
			Code:         defaultFirstCode,
			SecondCode:   defaultSecondCode,
			Message:      "message",
			UserAddr:     providedSender,
			GuardianAddr: "unknown addr",
		}
		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return []byte(humanReadable), nil
			},
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}

		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		signMessageAndCheckResults(t, args, request, nil, ErrInvalidGuardian)
	})
	t.Run("getGuardian fails, provided guardian is not usable yet", func(t *testing.T) {
		t.Parallel()

		providedUserInfoCopy := *providedUserInfo
		providedUserInfoCopy.FirstGuardian.State = core.NotUsable
		request := requests.SignMessage{
			Code:         defaultFirstCode,
			SecondCode:   defaultSecondCode,
			Message:      "message",
			UserAddr:     providedSender,
			GuardianAddr: string(providedUserInfoCopy.FirstGuardian.PublicKey),
		}
		args := createMockArgs()
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return []byte(humanReadable), nil
			},
			EncodeCalled: func(pkBytes []byte) string {
				return string(pkBytes)
			},
		}
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(&providedUserInfoCopy)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		signMessageAndCheckResults(t, args, request, nil, ErrGuardianNotUsable)
	})
	t.Run("cryptoComponentsHolderFactory creation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		providedRequest.GuardianAddr = string(providedUserInfo.FirstGuardian.PublicKey)
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.CryptoComponentsHolderFactory = &testscommon.CryptoComponentsHolderFactoryStub{
			CreateCalled: func(privateKeyBytes []byte) (sdkCore.CryptoComponentsHolder, error) {
				return nil, expectedErr
			},
		}
		signMessageAndCheckResults(t, args, providedRequest, nil, expectedErr)
	})
	t.Run("apply guardian signature fails", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMessage{
			Code:         defaultFirstCode,
			SecondCode:   defaultSecondCode,
			Message:      "message",
			UserAddr:     providedSender,
			GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
		}

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *transaction.FrontendTransaction) error {
				return expectedErr
			},
		}

		resolver, _ := NewServiceResolver(args)

		assert.NotNil(t, resolver)
		txHash, _, err := resolver.SignMessage("userIp", request)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHash)
	})
	t.Run("should work with sig verification", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMessage{
			Code:         defaultFirstCode,
			SecondCode:   defaultSecondCode,
			Message:      "message",
			UserAddr:     providedSender,
			GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
		}

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.CryptoComponentsHolderFactory = &testscommon.CryptoComponentsHolderFactoryStub{
			CreateCalled: func(privateKeyBytes []byte) (sdkCore.CryptoComponentsHolder, error) {
				c := &testscommon.CryptoComponentsHolderStub{
					GetPrivateKeyCalled: func() crypto.PrivateKey {
						return nil
					},
				}
				return c, nil
			},
		}
		args.SignatureVerifier = &testscommon.SignerStub{
			SignMessageCalled: func(msg []byte, _ crypto.PrivateKey) ([]byte, error) {
				return []byte("message"), nil
			},
		}

		resolver, _ := NewServiceResolver(args)

		assert.NotNil(t, resolver)
		message, _, err := resolver.SignMessage("userIp", request)
		assert.Nil(t, err)
		assert.Equal(t, []byte("message"), message)
	})
}

func TestServiceResolver_SignMultipleTransactions(t *testing.T) {
	t.Parallel()

	providedSender := "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
	providedRequest := requests.SignMultipleTransactions{
		Code:       defaultFirstCode,
		SecondCode: defaultSecondCode,
		Txs: []transaction.FrontendTransaction{
			{
				Sender:       providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
			},
			{
				Sender:       providedSender,
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
			},
		},
	}
	t.Run("tx validation fails, too many txs", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Txs: []transaction.FrontendTransaction{
				{
					Sender:       providedSender,
					Signature:    hex.EncodeToString([]byte("signature")),
					GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
				},
				{
					Sender:       providedSender,
					Signature:    hex.EncodeToString([]byte("signature")),
					GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
				},
				{
					Sender:       providedSender,
					Signature:    hex.EncodeToString([]byte("signature")),
					GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
				},
			},
		}
		args := createMockArgs()
		args.Config.MaxTransactionsAllowedForSigning = 2
		signMultipleTransactionsAndCheckResults(t, args, request, nil, ErrTooManyTransactionsToSign)
	})
	t.Run("tx validation fails, no tx", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Txs: []transaction.FrontendTransaction{},
		}
		args := createMockArgs()
		signMultipleTransactionsAndCheckResults(t, args, request, nil, ErrNoTransactionToSign)
	})
	t.Run("tx validation fails, tx has invalid sender", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Txs: []transaction.FrontendTransaction{
				{
					Sender: "invalid sender",
				},
			},
		}
		args := createMockArgs()
		signMultipleTransactionsAndCheckResults(t, args, request, nil, bech32.ErrInvalidCharacter(32))
	})
	t.Run("tx validation fails, different guardians on txs", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Txs: []transaction.FrontendTransaction{
				{
					Sender:       providedSender,
					Signature:    hex.EncodeToString([]byte("signature")),
					GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
				},
				{
					Sender:       providedSender,
					Signature:    hex.EncodeToString([]byte("signature")),
					GuardianAddr: string(providedUserInfo.SecondGuardian.PublicKey),
				},
			},
		}
		args := createMockArgs()
		signMultipleTransactionsAndCheckResults(t, args, request, nil, ErrGuardianMismatch)
	})
	t.Run("tx validation fails, different senders on txs", func(t *testing.T) {
		t.Parallel()

		request := requests.SignMultipleTransactions{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Txs: []transaction.FrontendTransaction{
				{
					Sender:    providedSender,
					Signature: hex.EncodeToString([]byte("signature")),
				},
				{
					Sender:    "erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49",
					Signature: hex.EncodeToString([]byte("signature")),
				},
			},
		}
		args := createMockArgs()
		signMultipleTransactionsAndCheckResults(t, args, request, nil, ErrInvalidSender)
	})
	t.Run("apply guardian signature fails for second tx", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		counter := 0
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *transaction.FrontendTransaction) error {
				counter++
				if counter > 1 {
					return expectedErr
				}
				return nil
			},
		}
		resolver, _ := NewServiceResolver(args)

		assert.NotNil(t, resolver)
		txHashes, _, err := resolver.SignMultipleTransactions("userIp", providedRequest)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHashes)
	})
	t.Run("crypto component holder creation fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		args.CryptoComponentsHolderFactory = &testscommon.CryptoComponentsHolderFactoryStub{
			CreateCalled: func(privateKeyBytes []byte) (sdkCore.CryptoComponentsHolder, error) {
				return nil, expectedErr
			},
		}
		resolver, _ := NewServiceResolver(args)

		assert.False(t, check.IfNil(resolver))
		txHashes, _, err := resolver.SignMultipleTransactions("userIp", providedRequest)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHashes)
	})
	t.Run("marshalling error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		cnt := 0
		args.TxMarshaller = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				// first marshal is for returning guardian
				// one per tx
				if cnt < 2 {
					cnt++
					return nil, nil
				}
				return nil, expectedErr
			},
		}
		resolver, _ := NewServiceResolver(args)

		assert.NotNil(t, resolver)
		txHashes, _, err := resolver.SignMultipleTransactions("userIp", providedRequest)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, txHashes)
	})
	t.Run("should work with sig verification", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *transaction.FrontendTransaction) error {
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

		assert.NotNil(t, resolver)
		txHashes, _, err := resolver.SignMultipleTransactions("userIp", providedRequest)
		assert.Equal(t, expectedResponse, txHashes)
		assert.Nil(t, err)
	})
	t.Run("should work without sig verification", func(t *testing.T) {
		t.Parallel()

		providedRequest := requests.SignMultipleTransactions{
			Code:       defaultFirstCode,
			SecondCode: defaultSecondCode,
			Txs: []transaction.FrontendTransaction{
				{
					Sender:       providedSender,
					Signature:    "",
					GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
				}, {
					Sender:       providedSender,
					Signature:    "",
					GuardianAddr: string(providedUserInfo.FirstGuardian.PublicKey),
				},
			},
		}

		args := createMockArgs()
		args.Config.SkipTxUserSigVerify = true
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			GetCalled: func(key []byte) ([]byte, error) {
				encryptedUser, err := args.UserEncryptor.EncryptUserInfo(providedUserInfo)
				require.Nil(t, err)
				return args.UserDataMarshaller.Marshal(encryptedUser)
			},
		}
		providedGuardianSignature := "provided signature"
		args.GuardedTxBuilder = &testscommon.GuardedTxBuilderStub{
			ApplyGuardianSignatureCalled: func(cryptoHolderGuardian sdkCore.CryptoComponentsHolder, tx *transaction.FrontendTransaction) error {
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

		assert.NotNil(t, resolver)
		txHashes, _, err := resolver.SignMultipleTransactions("userIp", providedRequest)
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

	assert.NotNil(t, resolver)
	count, err := resolver.RegisteredUsers()
	assert.Nil(t, err)
	assert.Equal(t, providedCount, count)
}

func TestServiceResolver_TcsConfig(t *testing.T) {
	t.Parallel()

	delayBetweenOTPWritesInSec := 60
	backoffTimeInSeconds := 600

	args := createMockArgs()
	args.Config.DelayBetweenOTPWritesInSec = uint64(delayBetweenOTPWritesInSec)

	secureOtpArgs := secureOtp.ArgsSecureOtpHandler{
		RateLimiter: testscommon.NewRateLimiterMock(3, backoffTimeInSeconds),
	}
	args.SecureOtpHandler, _ = secureOtp.NewSecureOtpHandler(secureOtpArgs)
	resolver, _ := NewServiceResolver(args)

	cfg := resolver.TcsConfig()
	require.Equal(t, delayBetweenOTPWritesInSec, int(cfg.OTPDelay))
	require.Equal(t, backoffTimeInSeconds, int(cfg.BackoffWrongCode))
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
	assert.NotNil(t, resolver)

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
	err := resolver.marshalAndSaveEncrypted(addr1.AddressBytes(), providedUserInfo1)
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

	err = resolver.marshalAndSaveEncrypted(addr2.AddressBytes(), providedUserInfo2)
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

func TestServiceResolver_parseUrl(t *testing.T) {
	t.Parallel()

	otpInfo, err := parseUrl("")
	assert.Equal(t, ErrEmptyUrl, err)
	assert.Equal(t, &requests.OTP{}, otpInfo)

	otpInfo, err = parseUrl("invalid path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "while parsing path")
	assert.True(t, errors.Is(err, ErrInvalidValue))
	assert.Equal(t, &requests.OTP{}, otpInfo)

	expectedOtpInfo := providedOTPInfo
	expectedOtpInfo.TimeSinceGeneration = 0
	otpInfo, err = parseUrl(providedUrl)
	assert.NoError(t, err)
	assert.Equal(t, expectedOtpInfo, otpInfo)
}

func checkGetGuardianAddressResults(t *testing.T, args ArgServiceResolver, userAddress sdkCore.AddressHandler, expectedErr error, expectedAddress []byte, otp handlers.OTP, expectedAge int64) {
	resolver, _ := NewServiceResolver(args)
	assert.NotNil(t, resolver)
	addr, otpAge, err := resolver.registerUser(userAddress, otp)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedAddress, addr)
	assert.LessOrEqual(t, otpAge, expectedAge)
}

func checkRegisterUserResults(t *testing.T, args ArgServiceResolver, userAddress sdkCore.AddressHandler, request requests.RegistrationPayload, expectedErr error, expectedOTPInfo *requests.OTP, expectedGuardian string) {
	resolver, _ := NewServiceResolver(args)
	assert.NotNil(t, resolver)
	otpInfo, guardian, err := resolver.RegisterUser(userAddress, request)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedOTPInfo, otpInfo)
	assert.Equal(t, expectedGuardian, guardian)
}

func checkVerifyCodeResults(t *testing.T, args ArgServiceResolver, userAddress sdkCore.AddressHandler, providedRequest requests.VerificationPayload, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.NotNil(t, resolver)
	_, err := resolver.VerifyCode(userAddress, "userIp", providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
}

func signMessageAndCheckResults(t *testing.T, args ArgServiceResolver, providedRequest requests.SignMessage, expectedMsg []byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.NotNil(t, resolver)
	txHash, _, err := resolver.SignMessage("userIp", providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedMsg, txHash)
}

func signTransactionAndCheckResults(t *testing.T, args ArgServiceResolver, providedRequest requests.SignTransaction, expectedHash []byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.NotNil(t, resolver)
	txHash, _, err := resolver.SignTransaction("userIp", providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedHash, txHash)
}

func signMultipleTransactionsAndCheckResults(t *testing.T, args ArgServiceResolver, providedRequest requests.SignMultipleTransactions, expectedHashes [][]byte, expectedErr error) {
	resolver, _ := NewServiceResolver(args)
	assert.NotNil(t, resolver)
	txHashes, _, err := resolver.SignMultipleTransactions("userIp", providedRequest)
	assert.True(t, errors.Is(err, expectedErr))
	assert.Equal(t, expectedHashes, txHashes)
}
