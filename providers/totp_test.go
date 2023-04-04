package providers

import (
	"crypto"
	"errors"
	"strings"
	"testing"
	"time"

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
		BackoffTime:       time.Minute * 5,
		MaxFailures:       3,
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
	t.Run("invalid value for MaxFailures should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.MaxFailures = minMaxFailures - 1
		totp, err := NewTimeBasedOnetimePassword(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "MaxFailures"))
		assert.True(t, check.IfNil(totp))
	})
	t.Run("invalid value for BackoffTimeInSeconds should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		args.BackoffTime = minBackoff - time.Millisecond
		totp, err := NewTimeBasedOnetimePassword(args)
		assert.True(t, errors.Is(err, ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "BackoffTime"))
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
			GetCalled: func(account, guardian []byte) (handlers.OTP, error) {
				return nil, expectedErr
			},
		}
		totp, _ := NewTimeBasedOnetimePassword(args)

		err := totp.ValidateCode([]byte("addr1"), []byte("guardian"), "userIp", "1234")
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgTimeBasedOneTimePassword()
		wasCalled := false
		providedCode := "1234"
		args.OTPStorageHandler = &testscommon.OTPStorageHandlerStub{
			GetCalled: func(account, guardian []byte) (handlers.OTP, error) {
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

		err := totp.ValidateCode([]byte("addr1"), []byte("guardian"), "userIp", providedCode)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("should increment verification attempts on error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		otpStub := &testscommon.TotpStub{
			ValidateCalled: func(userCode string) error {
				return expectedErr
			}}

		args.OTPStorageHandler = &testscommon.OTPStorageHandlerStub{
			GetCalled: func(account, guardian []byte) (handlers.OTP, error) {
				return otpStub, nil
			},
		}

		totp, _ := NewTimeBasedOnetimePassword(args)

		err := totp.ValidateCode([]byte("addr1"), []byte("guardian"), "userIp", "1234")

		// Verify that the expected error is returned and the failure count is incremented
		assert.Equal(t, expectedErr, err)
		key := "addr1:userIp"
		failures, found := totp.totalVerificationFailures[key]
		assert.True(t, found)
		assert.Equal(t, uint64(1), failures)
	})
	t.Run("frozen account should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgTimeBasedOneTimePassword()
		otpStub := &testscommon.TotpStub{
			ValidateCalled: func(userCode string) error {
				assert.Fail(t, "should not have been called")
				return nil
			}}

		args.OTPStorageHandler = &testscommon.OTPStorageHandlerStub{
			GetCalled: func(account, guardian []byte) (handlers.OTP, error) {
				return otpStub, nil
			},
		}

		totp, _ := NewTimeBasedOnetimePassword(args)
		key := "addr1:userIp"
		totp.frozenUsers[key] = time.Now()

		err := totp.ValidateCode([]byte("addr1"), []byte("guardian"), "userIp", "1234")

		// Verify that the expected error is returned and the failure count is incremented
		assert.Equal(t, ErrLockDown, err)

		_, found := totp.totalVerificationFailures[key]
		assert.False(t, found)
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

		qr, err := totp.RegisterUser([]byte("addr1"), []byte("guardian"), "addr1")
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

		qr, err := totp.RegisterUser([]byte("addr1"), []byte("guardian"), "addr1")
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
			SaveCalled: func(account, guardian []byte, otp handlers.OTP) error {
				return expectedErr
			},
		}
		totp, _ := NewTimeBasedOnetimePassword(args)

		qr, err := totp.RegisterUser([]byte("addr1"), []byte("guardian"), "addr1")
		assert.Nil(t, qr)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedAddr := []byte("addr1")
		providedTag := "tag"
		expectedQR := []byte("expected qr")
		args := createMockArgTimeBasedOneTimePassword()
		args.OTPStorageHandler = &testscommon.OTPStorageHandlerStub{
			SaveCalled: func(account, guardian []byte, otp handlers.OTP) error {
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

		qr, err := totp.RegisterUser(providedAddr, []byte("guardian"), providedTag)
		assert.Nil(t, err)
		assert.Equal(t, expectedQR, qr)
	})
}
