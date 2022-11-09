package providers

import (
	"crypto"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockArgTimeBasedOneTimePassword() ArgTimeBasedOneTimePassword {
	return ArgTimeBasedOneTimePassword{
		TOTPHandler:       &testsCommon.TOTPHandlerStub{},
		OTPStorageHandler: &testsCommon.OTPStorageHandlerStub{},
	}
}

func TestTimebasedOnetimePassword(t *testing.T) {
	t.Parallel()

	t.Run("nil totp handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = nil
		totp, err := NewTimebasedOnetimePassword(args)
		assert.Equal(t, ErrNilTOTPHandler, err)
		assert.True(t, check.IfNil(totp))
	})
	t.Run("nil storage handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.OTPStorageHandler = nil
		totp, err := NewTimebasedOnetimePassword(args)
		assert.Equal(t, ErrNilStorageHandler, err)
		assert.True(t, check.IfNil(totp))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		totp, err := NewTimebasedOnetimePassword(createMockArgTimeBasedOneTimePassword())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(totp))
	})
}
func TestTimebasedOnetimePassword_ValidateCode(t *testing.T) {
	t.Parallel()

	t.Run("storage handler returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.OTPStorageHandler = &testsCommon.OTPStorageHandlerStub{
			GetCalled: func(account, guardian string) (handlers.OTP, error) {
				return nil, expectedErr
			},
		}
		totp, _ := NewTimebasedOnetimePassword(args)

		err := totp.ValidateCode("addr1", "guardian", "1234")
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgTimeBasedOneTimePassword()
		wasCalled := false
		providedCode := "1234"
		args.OTPStorageHandler = &testsCommon.OTPStorageHandlerStub{
			GetCalled: func(account, guardian string) (handlers.OTP, error) {
				return &testsCommon.TotpStub{
					ValidateCalled: func(userCode string) error {
						assert.Equal(t, providedCode, userCode)
						wasCalled = true
						return nil
					},
				}, nil
			},
		}
		totp, _ := NewTimebasedOnetimePassword(args)

		err := totp.ValidateCode("addr1", "guardian", providedCode)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestTimebasedOnetimePassword_RegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("create totp returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = &testsCommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				return nil, expectedErr
			},
		}
		totp, _ := NewTimebasedOnetimePassword(args)

		qr, err := totp.RegisterUser("addr1", "guardian")
		assert.Nil(t, qr)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("otp.QR returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = &testsCommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				return &testsCommon.TotpStub{
					QRCalled: func() ([]byte, error) {
						return nil, expectedErr
					},
				}, nil
			},
		}
		totp, _ := NewTimebasedOnetimePassword(args)

		qr, err := totp.RegisterUser("addr1", "guardian")
		assert.Nil(t, qr)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("storage handler returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = &testsCommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				return &testsCommon.TotpStub{}, nil
			},
		}
		args.OTPStorageHandler = &testsCommon.OTPStorageHandlerStub{
			SaveCalled: func(account, guardian string, otp handlers.OTP) error {
				return expectedErr
			},
		}
		totp, _ := NewTimebasedOnetimePassword(args)

		qr, err := totp.RegisterUser("addr1", "guardian")
		assert.Nil(t, qr)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedQR := []byte("expected qr")
		args := createMockArgTimeBasedOneTimePassword()
		args.TOTPHandler = &testsCommon.TOTPHandlerStub{
			CreateTOTPCalled: func(account string, hash crypto.Hash) (handlers.OTP, error) {
				return &testsCommon.TotpStub{
					QRCalled: func() ([]byte, error) {
						return expectedQR, nil
					},
				}, nil
			},
		}
		totp, _ := NewTimebasedOnetimePassword(args)

		qr, err := totp.RegisterUser("addr1", "guardian")
		assert.Nil(t, err)
		assert.Equal(t, expectedQR, qr)
	})
}
