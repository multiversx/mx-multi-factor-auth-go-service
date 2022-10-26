package handlers

import (
	"crypto"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
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

	handler := NewTwoFactorHandler(6, "Elrond")
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
