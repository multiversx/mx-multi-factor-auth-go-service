package nativeAuth

import "errors"

// ErrNilAuthServer signals that a nil auth server has been provided
var ErrNilAuthServer = errors.New("nil auth server")

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")
