package handlers

import "errors"

// ErrNilTOTPHandler signals that a nil totp handler was provided
var ErrNilTOTPHandler = errors.New("nil totp handler")

// ErrNilOTP signals that a nil otp was provided
var ErrNilOTP = errors.New("nil otp")

// ErrNilDB signals that a nil database was provided
var ErrNilDB = errors.New("nil db")

// ErrNilBucketIDProvider signals that a nil bucket id provider was provided
var ErrNilBucketIDProvider = errors.New("nil bucket id provider")

// ErrNilBucketIndexHolder signals that a nil bucket index holder was provided
var ErrNilBucketIndexHolder = errors.New("nil bucket index holder")
