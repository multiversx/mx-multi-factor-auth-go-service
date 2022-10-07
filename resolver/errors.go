package resolver

import "errors"

// ErrProviderDoesNotExists signals that the given provider does exist for the given account
var ErrProviderDoesNotExists = errors.New("provider does not exist")

// ErrInvalidGuardian signals that the given guardian is not valid
var ErrInvalidGuardian = errors.New("invalid guardian")
