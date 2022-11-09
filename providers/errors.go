package providers

import "errors"

// ErrNilTOTPHandler signals that a nil totp handler was provided
var ErrNilTOTPHandler = errors.New("nil totp handler")

// ErrNilStorageHandler signals that a nil storage handler was provided
var ErrNilStorageHandler = errors.New("nil storage handler")
