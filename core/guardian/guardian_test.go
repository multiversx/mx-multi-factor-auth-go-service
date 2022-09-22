package guardian

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/testsCommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/stretchr/testify/assert"
)

func createMockGuardianConfig() config.GuardianConfig {
	return config.GuardianConfig{
		PrivateKeyFile:       "../../factory/testdata/alice.pem",
		RequestTimeInSeconds: 2,
	}
}

func TestNewGuardian(t *testing.T) {
	t.Parallel()
	t.Run("nil proxy", func(t *testing.T) {
		t.Parallel()

		conf := createMockGuardianConfig()

		g, err := NewGuardian(conf, nil)
		assert.True(t, check.IfNil(g))
		assert.True(t, errors.Is(err, ErrNilProxy))
	})
	t.Run("invalid RequestTimeInSeconds", func(t *testing.T) {
		t.Parallel()

		conf := createMockGuardianConfig()
		conf.RequestTimeInSeconds = 0
		g, err := NewGuardian(conf, &testsCommon.ProxyStub{})
		assert.True(t, check.IfNil(g))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value RequestTimeInSeconds"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		conf := createMockGuardianConfig()

		g, err := NewGuardian(conf, &testsCommon.ProxyStub{})
		assert.False(t, check.IfNil(g))
		assert.Nil(t, err)
	})
}
