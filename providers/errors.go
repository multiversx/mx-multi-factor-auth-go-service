package providers

import "errors"

var (
	initializationFailedError = errors.New("totp has not been initialized correctly")
	LockDownError             = errors.New("the verification is locked down, because of too many trials")
)
