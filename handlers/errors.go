package handlers

import "errors"

// ErrNoOtpForAddress signals that the given account has no otp register
var ErrNoOtpForAddress = errors.New("no otp created for account")

// ErrNoOtpForGuardian signals that the given guardian has no otp register
var ErrNoOtpForGuardian = errors.New("no otp created for guardian")

// ErrInvalidNumberOfGuardians signals that the number of guardians from file is invalid
var ErrInvalidNumberOfGuardians = errors.New("invalid number of guardians")

// ErrNilTOTPHandler signals that a nil totp handler was provided
var ErrNilTOTPHandler = errors.New("nil totp handler")

// ErrNilOTP signals that a nil otp was provided
var ErrNilOTP = errors.New("nil otp")

// ErrInvalidDBKey signals that an invalid db key was received
var ErrInvalidDBKey = errors.New("invalid db key")

// ErrNilDB signals that a nil database was provided
var ErrNilDB = errors.New("nil db")
