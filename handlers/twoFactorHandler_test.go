package handlers_test

import (
	"crypto"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestTwoFactorHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	handler := handlers.NewTwoFactorHandler(6, "MultiversX")
	assert.False(t, check.IfNil(handler))

	totp, err := handler.CreateTOTP("account", crypto.SHA1)
	assert.Nil(t, err)
	assert.NotNil(t, totp)

	bytes, err := totp.ToBytes()
	assert.Nil(t, err)

	totpFromBytes, err := handler.TOTPFromBytes(bytes)
	assert.Nil(t, err)
	assert.NotNil(t, totpFromBytes)
}
