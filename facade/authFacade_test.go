package facade

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO: modify and to tests for authFacade

var expectedErr = errors.New("expected error")

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
		t.Parallel()

		args := createMockArguments()
		args.ProvidersMap = nil

		facade, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.True(t, errors.Is(err, ErrEmptyProvidersMap))
	})
	t.Run("empty providersMap should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ProvidersMap = make(map[string]core.Provider)

		facade, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.True(t, errors.Is(err, ErrEmptyProvidersMap))
	})
	t.Run("providersMap contains nil provider", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ProvidersMap = make(map[string]core.Provider)
		args.ProvidersMap["totp"] = nil
		facade, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.Equal(t, fmt.Errorf("%s:%s", ErrNilProvider, "totp"), err)
	})
	t.Run("nil guardian should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.Guardian = nil

		facade, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facade))
		assert.True(t, errors.Is(err, ErrNilGuardian))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

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

	t.Run("nil codes array", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		facade, _ := NewAuthFacade(args)

		hash, err := facade.Validate(requests.SendTransaction{
			Account: "accnt1",
			Codes:   nil,
			Tx:      data.Transaction{},
		})
		require.Equal(t, "", hash)
		require.Equal(t, ErrEmptyCodesArray, err)
	})
	t.Run("empty codes array", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		facade, _ := NewAuthFacade(args)

		hash, err := facade.Validate(requests.SendTransaction{
			Account: "accnt1",
			Codes:   make([]requests.Code, 0),
			Tx:      data.Transaction{},
		})
		require.Equal(t, "", hash)
		require.Equal(t, ErrEmptyCodesArray, err)
	})
	t.Run("provider does not exist", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		facade, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: "invalid-provider",
				Code:     "123456",
			},
		)
		hash, err := facade.Validate(requests.SendTransaction{
			Account: "accnt1",
			Codes:   codes,
			Tx:      data.Transaction{},
		})
		require.Equal(t, "", hash)
		assert.True(t, strings.Contains(err.Error(), ErrProviderDoesNotExists.Error()))
	})
	t.Run("provider return error", func(t *testing.T) {
		t.Parallel()

		providersMap := make(map[string]core.Provider)
		account := "accnt1"
		provider := "totp"
		providersMap[provider] = &testsCommon.ProviderStub{
			ValidateCalled: func(account, userCode string) (bool, error) {
				return false, expectedErr
			},
		}
		args := createMockArguments()
		args.ProvidersMap = providersMap
		facade, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: provider,
				Code:     "123456",
			},
		)
		hash, err := facade.Validate(requests.SendTransaction{
			Account: account,
			Codes:   codes,
			Tx:      data.Transaction{},
		})
		require.Equal(t, "", hash)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})

	t.Run("provider return invalid request", func(t *testing.T) {
		t.Parallel()

		providersMap := make(map[string]core.Provider)
		account := "accnt1"
		provider := "totp"
		providersMap[provider] = &testsCommon.ProviderStub{
			ValidateCalled: func(account, userCode string) (bool, error) {
				return false, nil
			},
		}
		args := createMockArguments()
		args.ProvidersMap = providersMap
		facade, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: provider,
				Code:     "123456",
			},
		)
		hash, err := facade.Validate(requests.SendTransaction{
			Account: account,
			Codes:   codes,
			Tx:      data.Transaction{},
		})
		require.Equal(t, "", hash)
		assert.True(t, strings.Contains(err.Error(), ErrRequestNotValid.Error()))
	})
	t.Run("provider return valid request but guardian ValidateAndSend returns error", func(t *testing.T) {
		t.Parallel()

		providersMap := make(map[string]core.Provider)
		account := "accnt1"
		provider := "totp"
		providersMap[provider] = &testsCommon.ProviderStub{
			ValidateCalled: func(account, userCode string) (bool, error) {
				return true, nil
			},
		}
		args := createMockArguments()
		args.ProvidersMap = providersMap
		args.Guardian = &testsCommon.GuardianStub{
			ValidateAndSendCalled: func(transaction data.Transaction) (string, error) {
				return "", expectedErr
			},
		}
		facade, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: provider,
				Code:     "123456",
			},
		)
		hash, err := facade.Validate(requests.SendTransaction{
			Account: account,
			Codes:   codes,
			Tx:      data.Transaction{},
		})
		require.Equal(t, "", hash)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providersMap := make(map[string]core.Provider)
		account := "accnt1"
		provider := "totp"
		expectedHash := "expectedHash"
		providersMap[provider] = &testsCommon.ProviderStub{
			ValidateCalled: func(account, userCode string) (bool, error) {
				return true, nil
			},
		}
		args := createMockArguments()
		args.ProvidersMap = providersMap
		args.Guardian = &testsCommon.GuardianStub{
			ValidateAndSendCalled: func(transaction data.Transaction) (string, error) {
				return expectedHash, nil
			},
		}
		facade, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: provider,
				Code:     "123456",
			},
		)
		hash, err := facade.Validate(requests.SendTransaction{
			Account: account,
			Codes:   codes,
			Tx:      data.Transaction{},
		})
		require.Equal(t, expectedHash, hash)
		assert.Nil(t, err)
	})
}

func TestAuthFacade_RegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("provider does not exist", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		facade, _ := NewAuthFacade(args)

		dataBytes, err := facade.RegisterUser(requests.Register{
			Account:  "accnt1",
			Provider: "provider",
		})
		require.Equal(t, 0, len(dataBytes))
		assert.True(t, strings.Contains(err.Error(), ErrProviderDoesNotExists.Error()))
	})
	t.Run("provider return error", func(t *testing.T) {
		t.Parallel()

		providersMap := make(map[string]core.Provider)
		provider := "totp"
		providersMap[provider] = &testsCommon.ProviderStub{
			RegisterUserCalled: func(account string) ([]byte, error) {
				return make([]byte, 0), expectedErr
			},
		}
		args := createMockArguments()
		args.ProvidersMap = providersMap
		facade, _ := NewAuthFacade(args)

		dataBytes, err := facade.RegisterUser(requests.Register{
			Account:  "accnt1",
			Provider: provider,
		})
		require.Equal(t, 0, len(dataBytes))
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
}
