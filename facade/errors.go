package facade

import "errors"

// ErrNilGuardian signals that a nil guardian was provided
var ErrNilGuardian = errors.New("nil guardian")

// ErrEmptyProvidersMap signals that an empty providers map was provided
var ErrEmptyProvidersMap = errors.New("empty providers map")
