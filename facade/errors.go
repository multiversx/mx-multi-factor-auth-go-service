package facade

import "errors"

// ErrNilServiceResolver signals that a nil service resolver was provided
var ErrNilServiceResolver = errors.New("nil service resolver")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")
