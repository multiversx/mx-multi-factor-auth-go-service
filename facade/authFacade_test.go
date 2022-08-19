package facade

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO: modify and to tests for authFacade

func createMockArguments() ArgsAuthFacade {
	providersMap := make(map[string]core.Provider)
	providersMap["totp"] = &testsCommon.ProviderStub{}
	return ArgsAuthFacade{
		ProvidersMap: providersMap,
		Guardian:     &testsCommon.GuardianStub{},
		ApiInterface: core.WebServerOffString,
		PprofEnabled: true,
	}
}

func TestNewAuthFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil providersMap should error", func(t *testing.T) {
		args := createMockArguments()
		args.ProvidersMap = nil

		facade, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.True(t, errors.Is(err, ErrEmptyProvidersMap))
	})
	t.Run("empty providersMap should error", func(t *testing.T) {
		args := createMockArguments()
		args.ProvidersMap = make(map[string]core.Provider)

		facade, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.True(t, errors.Is(err, ErrEmptyProvidersMap))
	})
	t.Run("nil guardian should error", func(t *testing.T) {
		args := createMockArguments()
		args.Guardian = nil

		facade, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.True(t, errors.Is(err, ErrNilGuardian))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArguments()

		facade, err := NewAuthFacade(args)
		assert.False(t, check.IfNil(facade))
		assert.Nil(t, err)
	})
}

func TestAuthFacade_Getters(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	facade, _ := NewAuthFacade(args)

	assert.Equal(t, args.ApiInterface, facade.RestApiInterface())
	assert.Equal(t, args.PprofEnabled, facade.PprofEnabled())
}

func TestAuthFacade_Validate(t *testing.T) {
	t.Parallel()

	t.Run("empty codes array", func(t *testing.T) {
		args := createMockArguments()
		facade, _ := NewAuthFacade(args)

		response, err := facade.Validate()
		require.Nil(t, response)
		require.NotNil(t, err)
	})
}
