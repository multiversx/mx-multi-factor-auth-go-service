package providers

import (
	"crypto"
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockArgTimeBasedOneTimePassword() ArgTimeBasedOneTimePassword {
	return ArgTimeBasedOneTimePassword{
		TOTPHandler:       &testscommon.TOTPHandlerStub{},
		OTPStorageHandler: &testscommon.OTPStorageHandlerStub{},
	}
}

func TestTimeBasedOnetimePassword(t *testing.T) {
	t.Parallel()

	t.Run("nil totp handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = nil
		totp, err := NewTimeBasedOnetimePassword(args)
		assert.Equal(t, ErrNilTOTPHandler, err)
		assert.True(t, check.IfNil(totp))
	})
	t.Run("nil storage handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.OTPStorageHandler = nil
		totp, err := NewTimeBasedOnetimePassword(args)
		assert.Equal(t, ErrNilStorageHandler, err)
		assert.True(t, check.IfNil(totp))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		totp, err := NewTimeBasedOnetimePassword(createMockArgTimeBasedOneTimePassword())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(totp))
	})
}
func TestTimeBasedOnetimePassword_ValidateCode(t *testing.T) {
	t.Parallel()

	t.Run("storage handler returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.OTPStorageHandler = &testscommon.OTPStorageHandlerStub{
			GetCalled: func(account, guardian string) (handlers.OTP, error) {
				return nil, expectedErr
			},
		}
		totp, _ := NewTimeBasedOnetimePassword(args)

		err := totp.ValidateCode("addr1", "guardian", "1234")
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgTimeBasedOneTimePassword()
		wasCalled := false
		providedCode := "1234"
		args.OTPStorageHandler = &testscommon.OTPStorageHandlerStub{
			GetCalled: func(account, guardian string) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						assert.Equal(t, providedCode, userCode)
						wasCalled = true
						return nil
					},
				}, nil
			},
		}
		totp, _ := NewTimeBasedOnetimePassword(args)

		err := totp.ValidateCode("addr1", "guardian", providedCode)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestTimeBasedOnetimePassword_RegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("create totp returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				return nil, expectedErr
			},
		}
		totp, _ := NewTimeBasedOnetimePassword(args)

		qr, err := totp.RegisterUser("addr1", "addr1", "guardian")
		assert.Nil(t, qr)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("otp.QR returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				return &testscommon.TotpStub{
					QRCalled: func() ([]byte, error) {
						return nil, expectedErr
					},
				}, nil
			},
		}
		totp, _ := NewTimeBasedOnetimePassword(args)

		qr, err := totp.RegisterUser("addr1", "addr1", "guardian")
		assert.Nil(t, qr)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("storage handler returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				return &testscommon.TotpStub{}, nil
			},
		}
		args.OTPStorageHandler = &testscommon.OTPStorageHandlerStub{
			SaveCalled: func(account, guardian string, otp handlers.OTP) error {
				return expectedErr
			},
		}
		totp, _ := NewTimeBasedOnetimePassword(args)

		qr, err := totp.RegisterUser("addr1", "addr1", "guardian")
		assert.Nil(t, qr)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedAddr := "addr1"
		providedTag := "tag"
		expectedQR := []byte("expected qr")
		args := createMockArgTimeBasedOneTimePassword()
		args.OTPStorageHandler = &testscommon.OTPStorageHandlerStub{
			SaveCalled: func(account, guardian string, otp handlers.OTP) error {
				assert.Equal(t, providedAddr, account)
				return nil
			},
		}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				assert.Equal(t, providedTag, account)
				return &testscommon.TotpStub{
					QRCalled: func() ([]byte, error) {
						return expectedQR, nil
					},
				}, nil
			},
		}
		totp, _ := NewTimeBasedOnetimePassword(args)

		qr, err := totp.RegisterUser(providedAddr, providedTag, "guardian")
		assert.Nil(t, err)
		assert.Equal(t, expectedQR, qr)
	})
}
