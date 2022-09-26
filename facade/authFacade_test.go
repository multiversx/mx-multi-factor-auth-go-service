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

var (
	expectedErr      = errors.New("expected error")
	providedGuardian = "provided guardian"
)

func createMockArguments() ArgsAuthFacade {
	providersMap := make(map[string]core.Provider)
	providersMap["totp"] = &testsCommon.ProviderStub{}

	guardian := &testsCommon.GuardianStub{
		GetAddressCalled: func() string {
			return providedGuardian
		},
	}
	return ArgsAuthFacade{
		ProvidersMap: providersMap,
		Guardian:     guardian,
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

		facadeInstance, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facadeInstance))
		assert.True(t, errors.Is(err, ErrEmptyProvidersMap))
	})
	t.Run("empty providersMap should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ProvidersMap = make(map[string]core.Provider)

		facadeInstance, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facadeInstance))
		assert.True(t, errors.Is(err, ErrEmptyProvidersMap))
	})
	t.Run("providersMap contains nil provider", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ProvidersMap = make(map[string]core.Provider)
		args.ProvidersMap["totp"] = nil
		facadeInstance, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facadeInstance))
		assert.Equal(t, fmt.Errorf("%s:%s", ErrNilProvider, "totp"), err)
	})
	t.Run("nil guardian should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.Guardian = nil

		facadeInstance, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facadeInstance))
		assert.True(t, errors.Is(err, ErrNilGuardian))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()

		facadeInstance, err := NewAuthFacade(args)
		assert.False(t, check.IfNil(facadeInstance))
		assert.Nil(t, err)
	})
}

func TestAuthFacade_Getters(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	facadeInstance, _ := NewAuthFacade(args)

	assert.Equal(t, args.ApiInterface, facadeInstance.RestApiInterface())
	assert.Equal(t, args.PprofEnabled, facadeInstance.PprofEnabled())
}

func TestAuthFacade_Validate(t *testing.T) {
	t.Parallel()

	t.Run("nil codes array", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		facadeInstance, _ := NewAuthFacade(args)

		hash, err := facadeInstance.Validate(requests.SendTransaction{
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
		facadeInstance, _ := NewAuthFacade(args)

		hash, err := facadeInstance.Validate(requests.SendTransaction{
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
		facadeInstance, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: "invalid-provider",
				Code:     "123456",
			},
		)
		hash, err := facadeInstance.Validate(requests.SendTransaction{
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
		facadeInstance, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: provider,
				Code:     "123456",
			},
		)
		hash, err := facadeInstance.Validate(requests.SendTransaction{
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
		facadeInstance, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: provider,
				Code:     "123456",
			},
		)
		hash, err := facadeInstance.Validate(requests.SendTransaction{
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
		facadeInstance, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: provider,
				Code:     "123456",
			},
		)
		hash, err := facadeInstance.Validate(requests.SendTransaction{
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
		facadeInstance, _ := NewAuthFacade(args)

		codes := make([]requests.Code, 0)
		codes = append(codes,
			requests.Code{
				Provider: provider,
				Code:     "123456",
			},
		)
		hash, err := facadeInstance.Validate(requests.SendTransaction{
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
		facadeInstance, _ := NewAuthFacade(args)

		dataBytes, err := facadeInstance.RegisterUser(requests.Register{
			Account:  "accnt1",
			Provider: "provider",
		})
		require.Equal(t, 0, len(dataBytes))
		assert.True(t, strings.Contains(err.Error(), ErrProviderDoesNotExists.Error()))
	})
	t.Run("invalid guardian", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		facadeInstance, _ := NewAuthFacade(args)

		dataBytes, err := facadeInstance.RegisterUser(requests.Register{
			Account:  "accnt1",
			Provider: "provider",
			Guardian: "not a guardian",
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
		facadeInstance, _ := NewAuthFacade(args)

		dataBytes, err := facadeInstance.RegisterUser(requests.Register{
			Account:  "accnt1",
			Provider: provider,
			Guardian: providedGuardian,
		})
		require.Equal(t, 0, len(dataBytes))
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
}
