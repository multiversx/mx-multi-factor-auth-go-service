package guardian

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	mock2 "github.com/ElrondNetwork/elrond-go-core/core/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data/mock"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	erdgoTestscommon "github.com/ElrondNetwork/elrond-sdk-erdgo/testsCommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
)

const providedGuardianAddr = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"

func createMockGuardianConfig() config.GuardianConfig {
	return config.GuardianConfig{
		PrivateKeyFile:       "../../factory/testdata/alice.pem",
		RequestTimeInSeconds: 2,
	}
}

func createMockGuardianArgs() ArgGuardian {
	return ArgGuardian{
		Config:          createMockGuardianConfig(),
		Proxy:           &erdgoTestscommon.ProxyStub{},
		PubKeyConverter: &mock.PubkeyConverterStub{},
	}
}

func TestNewGuardian(t *testing.T) {
	t.Parallel()
	t.Run("nil proxy", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.Proxy = nil

		g, err := NewGuardian(args)
		assert.True(t, check.IfNil(g))
		assert.True(t, errors.Is(err, core.ErrNilProxy))
	})
	t.Run("invalid RequestTimeInSeconds", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.Config.RequestTimeInSeconds = 0

		g, err := NewGuardian(args)
		assert.True(t, check.IfNil(g))
		assert.True(t, errors.Is(err, core.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "RequestTimeInSeconds"))
	})
	t.Run("nil public key converter", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.PubKeyConverter = nil

		g, err := NewGuardian(args)
		assert.True(t, check.IfNil(g))
		assert.True(t, errors.Is(err, core.ErrNilPubkeyConverter))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		g, err := NewGuardian(createMockGuardianArgs())
		assert.False(t, check.IfNil(g))
		assert.Nil(t, err)
	})
}

func TestNewGuardian_usersHandler(t *testing.T) {
	t.Parallel()

	args := createMockGuardianArgs()
	args.Proxy = &erdgoTestscommon.ProxyStub{
		GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
			return &data.GuardianData{
				ActiveGuardian: &data.Guardian{
					Address: providedGuardianAddr,
				},
				PendingGuardian: &data.Guardian{},
			}, nil
		},
	}
	g, err := NewGuardian(args)
	assert.False(t, check.IfNil(g))
	assert.Nil(t, err)

	providedUser := "erd1p72ru5zcdsvgkkcm9swtvw2zy5epylwgv8vwquptkw7ga7pfvk7qz7snzw"
	callsMap := make(map[string]int)
	g.usersHandler = &testsCommon.UsersHandlerStub{
		AddUserCalled: func(address string) error {
			assert.Equal(t, providedUser, address)
			callsMap["AddUser"]++
			return nil
		},
		HasUserCalled: func(address string) bool {
			assert.Equal(t, providedUser, address)
			callsMap["HasUser"]++
			return address == providedUser
		},
		RemoveUserCalled: func(address string) {
			assert.Equal(t, providedUser, address)
			callsMap["RemoveUser"]++
		},
	}

	_ = g.AddUser(providedUser)
	assert.True(t, g.HasUser(providedUser))
	g.RemoveUser(providedUser)

	assert.Equal(t, 1, callsMap["AddUser"])
	assert.Equal(t, 1, callsMap["HasUser"])
	assert.Equal(t, 1, callsMap["RemoveUser"])
}

func TestNewGuardian_ValidateAndSend(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected err")
	sk, _ := hex.DecodeString("28654d9264f55f18d810bb88617e22c117df94fa684dfe341a511a72dfbf2b68")

	argsCreateTx := data.ArgCreateTransaction{
		Nonce:        1,
		Value:        "11500313000000000000",
		RcvAddr:      "erd1p72ru5zcdsvgkkcm9swtvw2zy5epylwgv8vwquptkw7ga7pfvk7qz7snzw",
		Data:         []byte("dummy"),
		ChainID:      "T",
		GuardianAddr: providedGuardianAddr,
		Options:      transaction.MaskGuardedTransaction,
	}
	tb, _ := builders.NewTxBuilder(blockchain.NewTxSigner())
	providedInitialTx, _ := tb.ApplyUserSignatureAndGenerateTx(sk, argsCreateTx)

	t.Run("different guardian address on tx should error", func(t *testing.T) {
		t.Parallel()

		providedTxCopy := *providedInitialTx
		providedTxCopy.GuardianAddr = "new guardian addr"
		g, _ := NewGuardian(createMockGuardianArgs())
		assert.False(t, check.IfNil(g))

		hash, err := g.ValidateAndSend(providedTxCopy)
		assert.Equal(t, "", hash)
		assert.Equal(t, core.ErrInvalidGuardianAddress, err)
	})
	t.Run("user not registered to guardian should error", func(t *testing.T) {
		t.Parallel()

		g, _ := NewGuardian(createMockGuardianArgs())
		assert.False(t, check.IfNil(g))

		hash, err := g.ValidateAndSend(*providedInitialTx)
		assert.Equal(t, "", hash)
		assert.Equal(t, core.ErrInvalidSenderAddress, err)
	})
	t.Run("pub key converter fails to decode should error", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.Proxy = &erdgoTestscommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: providedGuardianAddr,
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter = &mock.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return nil, expectedErr
			},
		}
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		_ = g.AddUser(providedInitialTx.SndAddr)
		hash, err := g.ValidateAndSend(*providedInitialTx)
		assert.Equal(t, "", hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("signature decode fails should error", func(t *testing.T) {
		t.Parallel()

		providedTxCopy := *providedInitialTx
		providedTxCopy.Signature = "non hex signature"
		args := createMockGuardianArgs()
		args.Proxy = &erdgoTestscommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: providedGuardianAddr,
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, &mock2.LoggerMock{})
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		_ = g.AddUser(providedTxCopy.SndAddr)
		hash, err := g.ValidateAndSend(providedTxCopy)
		assert.Equal(t, "", hash)
		assert.NotNil(t, err)
	})
	t.Run("signature verification fails should error", func(t *testing.T) {
		t.Parallel()

		providedTxCopy := *providedInitialTx
		providedTxCopy.SndAddr = providedTxCopy.RcvAddr
		args := createMockGuardianArgs()
		args.Proxy = &erdgoTestscommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: providedGuardianAddr,
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, &mock2.LoggerMock{})
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		_ = g.AddUser(providedTxCopy.SndAddr)
		hash, err := g.ValidateAndSend(providedTxCopy)
		assert.Equal(t, "", hash)
		assert.NotNil(t, err)
	})
	t.Run("ApplyGuardianSignature fails should error", func(t *testing.T) {
		t.Parallel()

		argsCopy := argsCreateTx
		argsCopy.Options = 0 // no MaskGuardedTransaction
		providedTx, _ := tb.ApplyUserSignatureAndGenerateTx(sk, argsCopy)
		args := createMockGuardianArgs()
		args.Proxy = &erdgoTestscommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: providedGuardianAddr,
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, &mock2.LoggerMock{})
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		_ = g.AddUser(providedTx.SndAddr)
		hash, err := g.ValidateAndSend(*providedTx)
		assert.Equal(t, "", hash)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "guardian flag"))
	})
	t.Run("send transaction returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.Proxy = &erdgoTestscommon.ProxyStub{
			SendTransactionCalled: func(tx *data.Transaction) (string, error) {
				return "", expectedErr
			},
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: providedGuardianAddr,
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, &mock2.LoggerMock{})
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		_ = g.AddUser(providedInitialTx.SndAddr)
		hash, err := g.ValidateAndSend(*providedInitialTx)
		assert.Equal(t, "", hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		wasSendTxCalled := false
		providedHash := "provided hash"
		args.Proxy = &erdgoTestscommon.ProxyStub{
			SendTransactionCalled: func(tx *data.Transaction) (string, error) {
				assert.True(t, len(tx.GuardianSignature) > 0)
				tx.GuardianSignature = ""
				assert.Equal(t, providedInitialTx, tx)
				wasSendTxCalled = true
				return providedHash, nil
			},
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: providedGuardianAddr,
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		args.PubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, &mock2.LoggerMock{})
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		_ = g.AddUser(providedInitialTx.SndAddr)
		hash, err := g.ValidateAndSend(*providedInitialTx)
		assert.Nil(t, err)
		assert.Equal(t, providedHash, hash)
		assert.True(t, wasSendTxCalled)
	})
}
func TestNewGuardian_AddUser(t *testing.T) {
	t.Parallel()

	t.Run("invalid user address should error", func(t *testing.T) {
		t.Parallel()

		g, _ := NewGuardian(createMockGuardianArgs())
		assert.False(t, check.IfNil(g))

		err := g.AddUser("invalid sender addr")
		assert.NotNil(t, err)
	})
	t.Run("GetGuardianData returns error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockGuardianArgs()
		args.Proxy = &erdgoTestscommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return nil, expectedErr
			},
		}
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		err := g.AddUser("erd1p72ru5zcdsvgkkcm9swtvw2zy5epylwgv8vwquptkw7ga7pfvk7qz7snzw")
		assert.Equal(t, expectedErr, err)
	})
	t.Run("GetGuardianData returns different active guardian should error", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.Proxy = &erdgoTestscommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: "other active guardian",
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		err := g.AddUser("erd1p72ru5zcdsvgkkcm9swtvw2zy5epylwgv8vwquptkw7ga7pfvk7qz7snzw")
		assert.Equal(t, core.ErrInactiveGuardian, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.Proxy = &erdgoTestscommon.ProxyStub{
			GetGuardianDataCalled: func(ctx context.Context, address erdgoCore.AddressHandler) (*data.GuardianData, error) {
				return &data.GuardianData{
					ActiveGuardian: &data.Guardian{
						Address: providedGuardianAddr,
					},
					PendingGuardian: &data.Guardian{},
				}, nil
			},
		}
		g, _ := NewGuardian(args)
		assert.False(t, check.IfNil(g))

		err := g.AddUser("erd1p72ru5zcdsvgkkcm9swtvw2zy5epylwgv8vwquptkw7ga7pfvk7qz7snzw")
		assert.Nil(t, err)
	})
}
