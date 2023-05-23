package frozenOtp_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/frozenOtp"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func createMockArgsFrozenOtpHandler() frozenOtp.ArgsFrozenOtpHandler {
	return frozenOtp.ArgsFrozenOtpHandler{
		MaxFailures: 3,
		BackoffTime: time.Minute * 5,
		RateLimiter: &testscommon.RateLimiterStub{},
	}
}

const (
	account = "test_account"
	ip      = "127.0.0.1"
	key     = account + ":" + ip
)

func TestNewFrozenOtpHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid max failures should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		args.MaxFailures = 0

		totp, err := frozenOtp.NewFrozenOtpHandler(args)
		require.True(t, errors.Is(err, handlers.ErrInvalidConfig))
		require.True(t, strings.Contains(err.Error(), "MaxFailures"))
		require.Nil(t, totp)
	})

	t.Run("invalid backoff time in seconds should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		args.BackoffTime = time.Second - time.Millisecond

		totp, err := frozenOtp.NewFrozenOtpHandler(args)
		require.True(t, errors.Is(err, handlers.ErrInvalidConfig))
		require.True(t, strings.Contains(err.Error(), "BackoffTime"))
		require.Nil(t, totp)
	})

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

		require.Equal(t, totp.BackOffTime(), uint64(args.BackoffTime.Seconds()))
	})
}

func TestFrozenOtpHandler_IsVerificationAllowed(t *testing.T) {
	t.Parallel()

	logger.SetLogLevel("*:DEBUG")

	t.Run("on error should return false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedCalled: func(key string, maxFailures int, maxDuration time.Duration) (int, error) {
				return 0, errors.New("err")
			},
		}
		totp, _ := frozenOtp.NewFrozenOtpHandler(args)

		isAllowed := totp.IsVerificationAllowed(account, ip)
		require.False(t, isAllowed)
	})

	t.Run("num remaining equals zero, should return false", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedCalled: func(key string, maxFailures int, maxDuration time.Duration) (int, error) {
				wasCalled = true
				return 0, nil
			},
		}
		totp, _ := frozenOtp.NewFrozenOtpHandler(args)

		isAllowed := totp.IsVerificationAllowed(account, ip)
		require.False(t, isAllowed)

		require.True(t, wasCalled)
	})

	t.Run("num remaining less than max, should return true", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		wasCalled := false
		args.RateLimiter = &testscommon.RateLimiterStub{
			CheckAllowedCalled: func(key string, maxFailures int, maxDuration time.Duration) (int, error) {
				wasCalled = true
				return 1, nil
			},
		}
		totp, _ := frozenOtp.NewFrozenOtpHandler(args)

		isAllowed := totp.IsVerificationAllowed(account, ip)
		require.True(t, isAllowed)

		require.True(t, wasCalled)
	})

	t.Run("should block after max verifications exceeded", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()

		args.RateLimiter = testscommon.NewRateLimiterMock()
		totp, _ := frozenOtp.NewFrozenOtpHandler(args)

		isAllowed := totp.IsVerificationAllowed(account, ip)
		require.True(t, isAllowed)
		isAllowed = totp.IsVerificationAllowed(account, ip)
		require.True(t, isAllowed)
		isAllowed = totp.IsVerificationAllowed(account, ip)
		require.True(t, isAllowed)
		isAllowed = totp.IsVerificationAllowed(account, ip)
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
