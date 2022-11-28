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

// InvalidNumberOfBuckets signals that an invalid number of buckets was received
var InvalidNumberOfBuckets = errors.New("invalid number of buckets")

// ErrInvalidBucketID signals that an invalid bucket id was provided
var ErrInvalidBucketID = errors.New("invalid bucket id")
