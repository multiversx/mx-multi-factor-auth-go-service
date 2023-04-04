package providers

import "errors"

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilTOTPHandler signals that a nil totp handler was provided
var ErrNilTOTPHandler = errors.New("nil totp handler")

// ErrNilStorageHandler signals that a nil storage handler was provided
var ErrNilStorageHandler = errors.New("nil storage handler")

// ErrLockDown signals that the verification is locked down, too many attempts
var ErrLockDown = errors.New("the verification is locked down, too many attempts")
