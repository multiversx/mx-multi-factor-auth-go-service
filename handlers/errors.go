package handlers

import "errors"

// ErrNoOtpForAddress signals that the given account has no otp register
var ErrNoOtpForAddress = errors.New("no otp created for account")

// ErrNoOtpForGuardian signals that the given guardian has no otp register
var ErrNoOtpForGuardian = errors.New("no otp created for guardian")

// ErrInvalidCode signals that the given code is invalid
var ErrInvalidCode = errors.New("invalid code")

// ErrCannotUpdateInformation signals that the address information cannot be updated
var ErrCannotUpdateInformation = errors.New("cannot update information")

// ErrInvalidNumberOfGuardians signals that the number of guardians from file is invalid
var ErrInvalidNumberOfGuardians = errors.New("invalid number of guardians")

// ErrGuardiansLimitReached signals that the guardians limit for account was reached
var ErrGuardiansLimitReached = errors.New("guardians limit reached")
