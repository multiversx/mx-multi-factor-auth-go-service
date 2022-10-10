package resolver

import "errors"

// ErrProviderDoesNotExists signals that the given provider does exist for the given account
var ErrProviderDoesNotExists = errors.New("provider does not exist")

// ErrInvalidGuardian signals that the given guardian is not valid
var ErrInvalidGuardian = errors.New("invalid guardian")

// ErrInvalidSender signals that the given sender is not valid
var ErrInvalidSender = errors.New("invalid sender")

// ErrGuardianNotYetUsable signals that the given guardian is not yet usable
var ErrGuardianNotYetUsable = errors.New("guardian not yet usable")

// ErrEmptyCodesSlice signals that the given slice of codes is empty
var ErrEmptyCodesSlice = errors.New("empty codes slice")
