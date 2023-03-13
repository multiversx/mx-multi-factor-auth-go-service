package sec51_test

import (
	"crypto"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/handlers/twofactor/sec51"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestSec51Wrapper_ShouldWork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if !check.IfNilReflect(r) {
			assert.Fail(t, "should not panic")
		}
	}()

	s := sec51.NewSec51Wrapper(6, "MultiversX")
	assert.False(t, check.IfNil(s))

	totp, err := s.GenerateTOTP("account", crypto.SHA1)
	assert.Nil(t, err)
	assert.NotNil(t, totp)

	bytes, err := totp.ToBytes()
	assert.Nil(t, err)

	totpFromBytes, err := s.TOTPFromBytes(bytes)
	assert.Nil(t, err)
	assert.NotNil(t, totpFromBytes)
}
