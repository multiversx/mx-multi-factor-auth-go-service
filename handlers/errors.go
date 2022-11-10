package handlers

import "errors"

// ErrNoOtpForAddress signals that the given account has no otp register
var ErrNoOtpForAddress = errors.New("no otp created for account")

// ErrNoOtpForGuardian signals that the given guardian has no otp register
var ErrNoOtpForGuardian = errors.New("no otp created for guardian")

// ErrInvalidNumberOfGuardians signals that the number of guardians from file is invalid
var ErrInvalidNumberOfGuardians = errors.New("invalid number of guardians")

// ErrGuardiansLimitReached signals that the guardians limit for account was reached
var ErrGuardiansLimitReached = errors.New("guardians limit reached")

// ErrNilTOTPHandler signals that a nil totp handler was provided
var ErrNilTOTPHandler = errors.New("nil totp handler")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrNilOTP signals that a nil otp was provided
var ErrNilOTP = errors.New("nil otp")
