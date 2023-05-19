package facade

import (
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArguments() ArgsGuardianFacade {
	return ArgsGuardianFacade{
		ServiceResolver:      &testscommon.ServiceResolverStub{},
		StatusMetricsHandler: &testscommon.StatusMetricsStub{},
	}
}

func TestNewGuardianFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil service resolver should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ServiceResolver = nil

		facadeInstance, err := NewGuardianFacade(args)
		assert.Nil(t, facadeInstance)
		assert.True(t, errors.Is(err, ErrNilServiceResolver))
	})

	t.Run("nil metrics handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.StatusMetricsHandler = nil

		facadeInstance, err := NewGuardianFacade(args)
		assert.Nil(t, facadeInstance)
		assert.True(t, errors.Is(err, core.ErrNilMetricsHandler))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewGuardianFacade(createMockArguments())
		assert.NotNil(t, facadeInstance)
		assert.Nil(t, err)
	})
}

func TestGuardianFacade_Getters(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	args := createMockArguments()
	expectedGuardian := "expected guardian"
	wasVerifyCodeCalled := false
	providedVerifyCodeReq := requests.VerificationPayload{
		Code:     "VerifyCode code",
		Guardian: "VerifyCode guardian",
	}
	providedUserAddress, _ := data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
	expectedQR := []byte("expected qr")
	wasRegisterUserCalled := false
	otpDelay := uint64(50)
	backoffWrongCode := uint64(100)

	args.ServiceResolver = &testscommon.ServiceResolverStub{
		VerifyCodeCalled: func(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) error {
			assert.Equal(t, providedVerifyCodeReq, request)
			wasVerifyCodeCalled = true
			return nil
		},
		RegisterUserCalled: func(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
			assert.Equal(t, providedUserAddress, userAddress)
			wasRegisterUserCalled = true
			return expectedQR, expectedGuardian, nil
		},
		TcsConfigCalled: func() *core.TcsConfig {
			return &core.TcsConfig{
				OTPDelay:         otpDelay,
				BackoffWrongCode: backoffWrongCode,
			}
		},
	}
	facadeInstance, _ := NewGuardianFacade(args)

	assert.Nil(t, facadeInstance.VerifyCode(providedUserAddress, "userIp", providedVerifyCodeReq))
	assert.True(t, wasVerifyCodeCalled)

	qr, guardian, err := facadeInstance.RegisterUser(providedUserAddress, requests.RegistrationPayload{})
	assert.Nil(t, err)
	assert.Equal(t, expectedQR, qr)
	assert.Equal(t, expectedGuardian, guardian)
	assert.True(t, wasRegisterUserCalled)

	tcsConfig := facadeInstance.TcsConfig()
	require.NotNil(t, tcsConfig)
	require.Equal(t, otpDelay, tcsConfig.OTPDelay)
	require.Equal(t, backoffWrongCode, tcsConfig.BackoffWrongCode)
}
