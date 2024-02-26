package tcs

import "errors"

// ErrNilConfigs signals that a nil config has been provided
var ErrNilConfigs = errors.New("nil configs provided")
