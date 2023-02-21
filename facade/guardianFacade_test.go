package facade

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgsGuardianFacade {
	return ArgsGuardianFacade{
		ServiceResolver: &testscommon.ServiceResolverStub{},
		ApiInterface:    core.WebServerOffString,
		PprofEnabled:    true,
	}
}

func TestNewGuardianFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil service resolver should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ServiceResolver = nil

		facadeInstance, err := NewGuardianFacade(args)
		assert.True(t, check.IfNil(facadeInstance))
		assert.True(t, errors.Is(err, ErrNilServiceResolver))
	})
	t.Run("invalid api interface should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ApiInterface = ""

		facadeInstance, err := NewGuardianFacade(args)
		assert.True(t, check.IfNil(facadeInstance))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "ApiInterface"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewGuardianFacade(createMockArguments())
		assert.False(t, check.IfNil(facadeInstance))
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
	args.ServiceResolver = &testscommon.ServiceResolverStub{
		VerifyCodeCalled: func(userAddress sdkCore.AddressHandler, request requests.VerificationPayload) error {
			assert.Equal(t, providedVerifyCodeReq, request)
			wasVerifyCodeCalled = true
			return nil
		},
		RegisterUserCalled: func(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
			assert.Equal(t, providedUserAddress, userAddress)
			wasRegisterUserCalled = true
			return expectedQR, expectedGuardian, nil
		},
	}
	facadeInstance, _ := NewGuardianFacade(args)

	assert.Equal(t, args.ApiInterface, facadeInstance.RestApiInterface())
	assert.Equal(t, args.PprofEnabled, facadeInstance.PprofEnabled())
	assert.Nil(t, facadeInstance.VerifyCode(providedUserAddress, providedVerifyCodeReq))
	assert.True(t, wasVerifyCodeCalled)

	qr, guardian, err := facadeInstance.RegisterUser(providedUserAddress, requests.RegistrationPayload{})
	assert.Nil(t, err)
	assert.Equal(t, expectedQR, qr)
	assert.Equal(t, expectedGuardian, guardian)
	assert.True(t, wasRegisterUserCalled)
}