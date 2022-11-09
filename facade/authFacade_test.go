package facade

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgsAuthFacade {
	return ArgsAuthFacade{
		ServiceResolver: &testsCommon.ServiceResolverStub{},
		ApiInterface:    core.WebServerOffString,
		PprofEnabled:    true,
	}
}

func TestNewAuthFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil service resolver should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ServiceResolver = nil

		facadeInstance, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facadeInstance))
		assert.True(t, errors.Is(err, ErrNilServiceResolver))
	})
	t.Run("invalid api interface should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ApiInterface = ""

		facadeInstance, err := NewAuthFacade(args)
		assert.True(t, check.IfNil(facadeInstance))
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "ApiInterface"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facadeInstance, err := NewAuthFacade(createMockArguments())
		assert.False(t, check.IfNil(facadeInstance))
		assert.Nil(t, err)
	})
}

func TestAuthFacade_Getters(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	args := createMockArguments()
	expectedGuardian := "expected guardian"
	wasGetGuardianAddressCalled := false
	providedGetGuardianAddressReq := requests.GetGuardianAddress{
		Credentials: "GetGuardianAddress credentials",
	}
	wasVerifyCodeCalled := false
	providedVerifyCodeReq := requests.VerificationPayload{
		Credentials: "VerifyCode credentials",
		Code:        "VerifyCode code",
		Guardian:    "VerifyCode guardian",
	}
	expectedQR := []byte("expected qr")
	wasRegisterUserCalled := false
	providedRegisterReq := requests.RegistrationPayload{
		Credentials: "Register credentials",
		Guardian:    "Register guardian",
	}
	args.ServiceResolver = &testsCommon.ServiceResolverStub{
		GetGuardianAddressCalled: func(request requests.GetGuardianAddress) (string, error) {
			assert.Equal(t, providedGetGuardianAddressReq, request)
			wasGetGuardianAddressCalled = true
			return expectedGuardian, nil
		},
		VerifyCodeCalled: func(request requests.VerificationPayload) error {
			assert.Equal(t, providedVerifyCodeReq, request)
			wasVerifyCodeCalled = true
			return nil
		},
		RegisterUserCalled: func(request requests.RegistrationPayload) ([]byte, error) {
			assert.Equal(t, providedRegisterReq, request)
			wasRegisterUserCalled = true
			return expectedQR, nil
		},
	}
	facadeInstance, _ := NewAuthFacade(args)

	assert.Equal(t, args.ApiInterface, facadeInstance.RestApiInterface())
	assert.Equal(t, args.PprofEnabled, facadeInstance.PprofEnabled())
	address, err := facadeInstance.GetGuardianAddress(providedGetGuardianAddressReq)
	assert.Nil(t, err)
	assert.Equal(t, expectedGuardian, address)
	assert.True(t, wasGetGuardianAddressCalled)

	assert.Nil(t, facadeInstance.VerifyCode(providedVerifyCodeReq))
	assert.True(t, wasVerifyCodeCalled)

	qr, err := facadeInstance.RegisterUser(providedRegisterReq)
	assert.Nil(t, err)
	assert.Equal(t, expectedQR, qr)
	assert.True(t, wasRegisterUserCalled)
}
