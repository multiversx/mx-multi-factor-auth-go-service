package handlers

import "errors"

// ErrInvalidConfig signals that an invalid configuration was provided
var ErrInvalidConfig = errors.New("invalid config")

// ErrRegistrationFailed signals that registration failed
var ErrRegistrationFailed = errors.New("registration failed")

// ErrNilOTPProvider signals that a nil otp provider was provided
var ErrNilOTPProvider = errors.New("nil otp provider")

// ErrNilRateLimiter signals that a nil rate limiter was provided
var ErrNilRateLimiter = errors.New("nil rate limiter")
