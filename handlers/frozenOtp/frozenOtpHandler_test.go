package frozenOtp

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/stretchr/testify/assert"
)

func createMockArgsFrozenOtpHandler() ArgsFrozenOtpHandler {
	return ArgsFrozenOtpHandler{
		MaxFailures: 3,
		BackoffTime: time.Minute * 5,
	}
}

const (
	account = "test_account"
	ip      = "127.0.0.1"
	key     = account + ":" + ip
)

func TestNewFrozenOtpHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid value for MaxFailures should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		args.MaxFailures = minMaxFailures - 1
		totp, err := NewFrozenOtpHandler(args)
		assert.True(t, errors.Is(err, handlers.ErrInvalidConfig))
		assert.True(t, strings.Contains(err.Error(), "MaxFailures"))
		assert.Nil(t, totp)
	})
	t.Run("invalid value for BackoffTimeInSeconds should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		args.BackoffTime = minBackoff - time.Millisecond
		totp, err := NewFrozenOtpHandler(args)
		assert.True(t, errors.Is(err, handlers.ErrInvalidConfig))
		assert.True(t, strings.Contains(err.Error(), "BackoffTime"))
		assert.Nil(t, totp)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, err := NewFrozenOtpHandler(args)
		assert.Nil(t, err)
		assert.NotNil(t, totp)
		assert.False(t, totp.IsInterfaceNil())
	})
}

func TestFrozenOtpHandler_IncrementFailures(t *testing.T) {
	t.Parallel()

	t.Run("should increment failures", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		for i := uint8(0); i < args.MaxFailures-1; i++ {
			totp.IncrementFailures(account, ip)
		}

		// Verify the number of failures for the given account and ip
		info := totp.verificationFailures[key]
		assert.Equal(t, args.MaxFailures-1, info.failures)
	})

	t.Run("should reset failures after too much time since last failure", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		totp.verificationFailures[key] = &failuresInfo{
			failures:     args.MaxFailures - 1,
			lastFailTime: time.Now().UTC().Add(-args.BackoffTime - time.Minute),
		}

		totp.IncrementFailures(account, ip)

		// Verify the number of failures for the given account and ip
		info := totp.verificationFailures[key]
		assert.Equal(t, uint8(1), info.failures)
	})

	t.Run("should freeze user after max failures", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		for i := uint8(0); i < args.MaxFailures+1; i++ {
			totp.IncrementFailures(account, ip)
		}

		info, isPresent := totp.verificationFailures[key]
		assert.True(t, isPresent)
		assert.Equal(t, args.MaxFailures, info.failures)
	})
}

func TestFrozenOtpHandler_IsVerificationAllowed(t *testing.T) {
	t.Parallel()

	t.Run("should return true when user is new", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		isAllowed := totp.IsVerificationAllowed(account, ip)
		assert.True(t, isAllowed)
	})
	t.Run("should return true when user still has failures left", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		totp.verificationFailures[key] = &failuresInfo{
			failures: args.MaxFailures - 1,
		}

		isAllowed := totp.IsVerificationAllowed(account, ip)
		assert.True(t, isAllowed)
	})
	t.Run("should return false when user is frozen and backoff time has not elapsed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		totp.verificationFailures[key] = &failuresInfo{
			failures:     args.MaxFailures,
			lastFailTime: time.Now().UTC().Add(-args.BackoffTime + time.Minute),
		}

		isAllowed := totp.IsVerificationAllowed(account, ip)
		assert.False(t, isAllowed)
	})
	t.Run("should return true when user is frozen and backoff time has elapsed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		totp.verificationFailures[key] = &failuresInfo{
			failures:     args.MaxFailures,
			lastFailTime: time.Now().UTC().Add(-args.BackoffTime - time.Minute),
		}

		isAllowed := totp.IsVerificationAllowed(account, ip)
		assert.True(t, isAllowed)
		_, isPresent := totp.verificationFailures[key]
		assert.False(t, isPresent)
	})
}

func TestFrozenOtpHandler_isBackoffExpired(t *testing.T) {
	t.Parallel()

	t.Run("should return true when backoff time has elapsed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		lastVerification := time.Now().UTC().Add(-args.BackoffTime - time.Minute)
		isValid := totp.isBackoffExpired(lastVerification)
		assert.True(t, isValid)
	})

	t.Run("should return false when backoff time has not elapsed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		lastVerification := time.Now().UTC().Add(-args.BackoffTime + time.Minute)
		isValid := totp.isBackoffExpired(lastVerification)
		assert.False(t, isValid)
	})
}
