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

// ErrNilMarshaller signals that a nil marshaller was provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrRegistrationFailed signals that registration failed
var ErrRegistrationFailed = errors.New("registration failed")

// ErrNilOTPProvider signals that a nil otp provider was provided
var ErrNilOTPProvider = errors.New("nil otp provider")

// ErrNilUserInfo signals that a nil user info was provided
var ErrNilUserInfo = errors.New("nil user info")

// ErrGuardianNotFound signals that a guardian was not found
var ErrGuardianNotFound = errors.New("guardian not found")
