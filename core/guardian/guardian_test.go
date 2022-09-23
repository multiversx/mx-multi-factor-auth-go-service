package guardian

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	erdgoTestsCommon "github.com/ElrondNetwork/elrond-sdk-erdgo/testsCommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
)

func createMockGuardianConfig() config.GuardianConfig {
	return config.GuardianConfig{
		PrivateKeyFile:       "../../factory/testdata/alice.pem",
		RequestTimeInSeconds: 2,
	}
}

func createMockGuardianArgs() ArgGuardian {
	return ArgGuardian{
		Config:          createMockGuardianConfig(),
		Proxy:           &erdgoTestsCommon.ProxyStub{},
		PubKeyConverter: &testsCommon.PubkeyConverterStub{},
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
		assert.True(t, errors.Is(err, ErrNilProxy))
	})
	t.Run("invalid RequestTimeInSeconds", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.Config.RequestTimeInSeconds = 0

		g, err := NewGuardian(args)
		assert.True(t, check.IfNil(g))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value RequestTimeInSeconds"))
	})
	t.Run("nil public key converter", func(t *testing.T) {
		t.Parallel()

		args := createMockGuardianArgs()
		args.PubKeyConverter = nil

		g, err := NewGuardian(args)
		assert.True(t, check.IfNil(g))
		assert.True(t, errors.Is(err, ErrNilPubkeyConverter))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		g, err := NewGuardian(createMockGuardianArgs())
		assert.False(t, check.IfNil(g))
		assert.Nil(t, err)
	})
}
