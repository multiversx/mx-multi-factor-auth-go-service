package handlers

import "errors"

// ErrNilTOTPHandler signals that a nil totp handler was provided
var ErrNilTOTPHandler = errors.New("nil totp handler")

// ErrNilOTP signals that a nil otp was provided
var ErrNilOTP = errors.New("nil otp")

// ErrNilDB signals that a nil database was provided
var ErrNilDB = errors.New("nil db")

// ErrInvalidConfig signals that an invalid configuration was provided
var ErrInvalidConfig = errors.New("invalid config")
