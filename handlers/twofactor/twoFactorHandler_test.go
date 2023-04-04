package twofactor_test

import (
	"crypto"
	"fmt"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/twofactor"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestTwoFactorHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("nil otp provider should error", func(t *testing.T) {
		t.Parallel()

		handler, err := twofactor.NewTwoFactorHandler(nil)
		assert.Equal(t, handlers.ErrNilOTPProvider, err)
		assert.Nil(t, handler)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasGenerateTOTPCalled := false
		wasTOTPFromBytesCalled := false
		handler, err := twofactor.NewTwoFactorHandler(&testscommon.OTPProviderStub{
			GenerateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				wasGenerateTOTPCalled = true
				return &testscommon.TotpStub{}, nil
			},
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				wasTOTPFromBytesCalled = true
				return &testscommon.TotpStub{}, nil
			},
		})
		assert.Nil(t, err)
		assert.NotNil(t, handler)

		totp, err := handler.CreateTOTP("account", crypto.SHA1)
		assert.Nil(t, err)
		assert.Equal(t, "*testscommon.TotpStub", fmt.Sprintf("%T", totp))
		assert.True(t, wasGenerateTOTPCalled)

		totpFromBytes, err := handler.TOTPFromBytes([]byte(""))
		assert.Nil(t, err)
		assert.Equal(t, "*testscommon.TotpStub", fmt.Sprintf("%T", totpFromBytes))
		assert.True(t, wasTOTPFromBytesCalled)
	})
}
