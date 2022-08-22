package facade

import "errors"

// ErrNilGuardian signals that a nil guardian was provided
var ErrNilGuardian = errors.New("nil guardian")

// ErrNilProvider signals that a nil provider was provided
var ErrNilProvider = errors.New("nil provider")

// ErrEmptyProvidersMap signals that an empty providers map was provided
var ErrEmptyProvidersMap = errors.New("empty providers map")

// ErrEmptyCodesArray signals that an empty array for the codes has been provided
var ErrEmptyCodesArray = errors.New("empty codes array")

// ErrProviderDoesNotExists signals that the given provider does exist for the given account
var ErrProviderDoesNotExists = errors.New("provider does not exist")

// ErrRequestNotValid signals that the given request is not valid
var ErrRequestNotValid = errors.New("request not valid")
