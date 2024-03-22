package secureOtp_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/secureOtp"
	"github.com/multiversx/mx-multi-factor-auth-go-service/redis"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon"
)

func createMockArgsFrozenOtpHandler() secureOtp.ArgsSecureOtpHandler {
	return secureOtp.ArgsSecureOtpHandler{
		RateLimiter: &testscommon.RateLimiterStub{},
	}
}

const (
	account = "test_account"
	ip      = "127.0.0.1"
)

func TestNewFrozenOtpHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil rate limiter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		args.RateLimiter = nil

		totp, err := secureOtp.NewFrozenOtpHandler(args)
		require.True(t, errors.Is(err, handlers.ErrNilRateLimiter))
		require.Nil(t, totp)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, err := secureOtp.NewFrozenOtpHandler(args)
		require.Nil(t, err)
		require.NotNil(t, totp)
		require.False(t, totp.IsInterfaceNil())
	})
}

func TestFrozenOtpHandler_IsVerificationAllowed(t *testing.T) {
	t.Parallel()

	t.Run("on error should return err", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		expectedErr := errors.New("expected error")
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedAndIncreaseTrialsCalled: func(key string) (*redis.RateLimiterResult, error) {
				return &redis.RateLimiterResult{}, expectedErr
			},
		}
		totp, _ := secureOtp.NewFrozenOtpHandler(args)

		_, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Equal(t, expectedErr, err)
	})

	t.Run("num allowed equals zero, should return false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedAndIncreaseTrialsCalled: func(key string) (*redis.RateLimiterResult, error) {
				wasCalled = true
				return &redis.RateLimiterResult{Allowed: false}, nil
			},
		}
		totp, _ := secureOtp.NewFrozenOtpHandler(args)

		_, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)

		require.True(t, wasCalled)
	})

	t.Run("num allowed equals one, should return true", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedAndIncreaseTrialsCalled: func(key string) (*redis.RateLimiterResult, error) {
				wasCalled = true
				return &redis.RateLimiterResult{Allowed: true, Remaining: 1, ResetAfter: time.Duration(10) * time.Second}, nil
			},
		}
		totp, _ := secureOtp.NewFrozenOtpHandler(args)

		codeVerifyData, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)

		require.True(t, wasCalled)
		require.Equal(t, 1, codeVerifyData.RemainingTrials)
		require.Equal(t, 10, codeVerifyData.ResetAfter)
	})

	t.Run("should block after max verifications exceeded", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		args.RateLimiter = testscommon.NewRateLimiterMock(3, 10)
		totp, _ := secureOtp.NewFrozenOtpHandler(args)

		_, err := totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		_, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		_, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Nil(t, err)
		_, err = totp.IsVerificationAllowedAndIncreaseTrials(account, ip)
		require.Equal(t, core.ErrTooManyFailedAttempts, err)
	})
}

func TestFrozenOtpHandler_Reset(t *testing.T) {
	t.Parallel()

	args := createMockArgsFrozenOtpHandler()

	wasCalled := false
	args.RateLimiter = &testscommon.RateLimiterStub{
		ResetCalled: func(key string) error {
			wasCalled = true
			return nil
		},
	}
	totp, _ := secureOtp.NewFrozenOtpHandler(args)

	totp.Reset(account, ip)

	require.True(t, wasCalled)
}
