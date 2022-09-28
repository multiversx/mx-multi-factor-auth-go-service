package guardian

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/mock"
	erdgoTestscommon "github.com/ElrondNetwork/elrond-sdk-erdgo/testsCommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/config"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
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

	g, err := NewGuardian(createMockGuardianArgs())
	assert.False(t, check.IfNil(g))
	assert.Nil(t, err)

	providedUser := "provided user"
	callsMap := make(map[string]int)
	g.usersHandler = &testsCommon.UsersHandlerStub{
		AddUserCalled: func(address string) {
			assert.Equal(t, providedUser, address)
			callsMap["AddUser"]++
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

	g.AddUser(providedUser)
	assert.True(t, g.HasUser(providedUser))
	g.RemoveUser(providedUser)

	assert.Equal(t, 1, callsMap["AddUser"])
	assert.Equal(t, 1, callsMap["HasUser"])
	assert.Equal(t, 1, callsMap["RemoveUser"])
}
