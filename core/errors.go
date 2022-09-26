package core

import "errors"

// ErrNilProxy signals that a nil proxy was provided
var ErrNilProxy = errors.New("nil proxy")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilPubkeyConverter signals that nil address converter was provided
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrInvalidGuardianAddress signals that the guardian address is invalid
var ErrInvalidGuardianAddress = errors.New("invalid guardian address")

// ErrInvalidSenderAddress signals that the sender address is invalid
var ErrInvalidSenderAddress = errors.New("invalid sender address")
