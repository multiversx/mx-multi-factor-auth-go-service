package frozenOtp

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func createMockArgsFrozenOtpHandler() ArgsFrozenOtpHandler {
	return ArgsFrozenOtpHandler{
		MaxFailures: 3,
		BackoffTime: time.Minute * 5,
	}
}

func TestNewFrozenOtpHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid value for MaxFailures should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		args.MaxFailures = minMaxFailures - 1
		totp, err := NewFrozenOtpHandler(args)
		assert.True(t, errors.Is(err, handlers.ErrInvalidConfig))
		assert.True(t, strings.Contains(err.Error(), "MaxFailures"))
		assert.True(t, check.IfNil(totp))
	})
	t.Run("invalid value for BackoffTimeInSeconds should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		args.BackoffTime = minBackoff - time.Millisecond
		totp, err := NewFrozenOtpHandler(args)
		assert.True(t, errors.Is(err, handlers.ErrInvalidConfig))
		assert.True(t, strings.Contains(err.Error(), "BackoffTime"))
		assert.True(t, check.IfNil(totp))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, err := NewFrozenOtpHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(totp))
	})
}

func TestTimeBasedOnetimePassword_incrementFailures(t *testing.T) {
	t.Parallel()

	account := []byte("test_account")
	ip := "127.0.0.1"

	t.Run("should increment failures", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		for i := uint64(0); i < args.MaxFailures-1; i++ {
			totp.IncrementFailures(account, ip)
		}

		// Verify the number of failures for the given account and ip
		key := string(account) + ":" + ip
		failures := totp.totalVerificationFailures[key]
		assert.Equal(t, args.MaxFailures-1, failures)
	})

	t.Run("should freeze user after max failures", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		for i := uint64(0); i < args.MaxFailures; i++ {
			totp.IncrementFailures(account, ip)
		}

		key := string(account) + ":" + ip
		_, isFrozen := totp.frozenUsers[key]
		assert.True(t, isFrozen)

		_, isPresent := totp.totalVerificationFailures[key]
		assert.False(t, isPresent)
	})
}

func TestTimeBasedOnetimePassword_checkFrozen(t *testing.T) {
	t.Parallel()

	account := []byte("test_account")
	ip := "127.0.0.1"
	key := string(account) + ":" + ip

	t.Run("should return true when user is not frozen", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		isAllowed := totp.IsVerificationAllowed(account, ip)
		assert.True(t, isAllowed)
	})
	t.Run("should return false when user is frozen and backoff time has not elapsed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		totp.frozenUsers[key] = time.Now().UTC().Add(-args.BackoffTime + time.Minute)

		isAllowed := totp.IsVerificationAllowed(account, ip)
		assert.False(t, isAllowed)
	})
	t.Run("should return true when user is frozen and backoff time has elapsed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsFrozenOtpHandler()
		totp, _ := NewFrozenOtpHandler(args)

		totp.frozenUsers[key] = time.Now().UTC().Add(-args.BackoffTime - time.Minute)

		isAllowed := totp.IsVerificationAllowed(account, ip)
		assert.True(t, isAllowed)
		_, found := totp.frozenUsers[key]
		assert.False(t, found)
	})
}

func TestTimeBasedOnetimePassword_validBackoffTime(t *testing.T) {
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
