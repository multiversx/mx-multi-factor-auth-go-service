package providers

import "errors"

//TODO: remove this file

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrLockDown signals that the verification is locked down, too many attempts
var ErrLockDown = errors.New("the verification is locked down, too many attempts")
