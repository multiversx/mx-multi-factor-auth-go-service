package facade

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO: modify and to tests for authFacade

func createMockArguments() ArgsAuthFacade {
	return ArgsAuthFacade{
		RarityCalculator: modules.NewRarityCalculator(),
		ApiInterface:     core.WebServerOffString,
		PprofEnabled:     true,
	}
}

func TestNewRelayerFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil rarityCalculator should error", func(t *testing.T) {
		args := createMockArguments()
		args.RarityCalculator = nil

		facade, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.True(t, errors.Is(err, ErrNilRarityCalculator))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArguments()

		facade, err := NewAuthFacade(args)
		assert.False(t, check.IfNil(facade))
		assert.Nil(t, err)
	})
}

func TestRelayerFacade_Getters(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	facade, _ := NewAuthFacade(args)

	assert.Equal(t, args.ApiInterface, facade.RestApiInterface())
	assert.Equal(t, args.PprofEnabled, facade.PprofEnabled())
}

func TestRelayerFacade_GetMetrics(t *testing.T) {
	t.Parallel()

	rarityCalculator := modules.NewRarityCalculator()
	require.NotNil(t, rarityCalculator)

	t.Run("should return rarity", func(t *testing.T) {
		args := createMockArguments()
		args.RarityCalculator = rarityCalculator
		facade, _ := NewAuthFacade(args)

		response, err := facade.GetRarity("OGS-123456")
		require.Nil(t, response)
		require.NotNil(t, err)
	})
}
