package frozenOtp_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/frozenOtp"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
)

func createMockArgsFrozenOtpHandler() frozenOtp.ArgsFrozenOtpHandler {
	return frozenOtp.ArgsFrozenOtpHandler{
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

		totp, err := frozenOtp.NewFrozenOtpHandler(args)
		require.True(t, errors.Is(err, handlers.ErrNilRateLimiter))
		require.Nil(t, totp)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, err := frozenOtp.NewFrozenOtpHandler(args)
		require.Nil(t, err)
		require.NotNil(t, totp)
		require.False(t, totp.IsInterfaceNil())
	})
}

func TestFrozenOtpHandler_IsVerificationAllowed(t *testing.T) {
	t.Parallel()

	t.Run("on error should return false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedCalled: func(key string) (*redis.RateLimiterResult, error) {
				return &redis.RateLimiterResult{}, errors.New("err")
			},
		}
		totp, _ := frozenOtp.NewFrozenOtpHandler(args)

		_, isAllowed := totp.IsVerificationAllowed(account, ip)
		require.False(t, isAllowed)
	})

	t.Run("num remaining equals zero, should return false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedCalled: func(key string) (*redis.RateLimiterResult, error) {
				wasCalled = true
				return &redis.RateLimiterResult{}, nil
			},
		}
		totp, _ := frozenOtp.NewFrozenOtpHandler(args)

		_, isAllowed := totp.IsVerificationAllowed(account, ip)
		require.False(t, isAllowed)

		require.True(t, wasCalled)
	})

	t.Run("num remaining less than max, should return true", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedCalled: func(key string) (*redis.RateLimiterResult, error) {
				wasCalled = true
				return &redis.RateLimiterResult{Remaining: 1, ResetAfter: time.Duration(10) * time.Second}, nil
			},
		}
		totp, _ := frozenOtp.NewFrozenOtpHandler(args)

		codeVerifyData, isAllowed := totp.IsVerificationAllowed(account, ip)
		require.True(t, isAllowed)

		require.True(t, wasCalled)
		require.Equal(t, 1, codeVerifyData.RemainingTrials)
		require.Equal(t, 10, codeVerifyData.ResetAfter)
	})

	t.Run("should block after max verifications exceeded", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		args.RateLimiter = testscommon.NewRateLimiterMock(3, 10)
		totp, _ := frozenOtp.NewFrozenOtpHandler(args)

		_, isAllowed := totp.IsVerificationAllowed(account, ip)
		require.True(t, isAllowed)
		_, isAllowed = totp.IsVerificationAllowed(account, ip)
		require.True(t, isAllowed)
		_, isAllowed = totp.IsVerificationAllowed(account, ip)
		require.True(t, isAllowed)
		_, isAllowed = totp.IsVerificationAllowed(account, ip)
		require.False(t, isAllowed)
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
	totp, _ := frozenOtp.NewFrozenOtpHandler(args)

	totp.Reset(account, ip)

	require.True(t, wasCalled)
}
