package providers

import "errors"

// ErrNoOtpForAddress signals that the given account has no otp register
var ErrNoOtpForAddress = errors.New("no otp created for account")

// ErrInvalidCode signals that the given code is invalid
var ErrInvalidCode = errors.New("invalid code")

// ErrCannotUpdateInformation signals that the address information cannot be updated
var ErrCannotUpdateInformation = errors.New("cannot update information")

// ErrCannotUpdateInformation signals that the address information cannot be updated
var ErrCannotGenerateQR = errors.New("cannot update information")
