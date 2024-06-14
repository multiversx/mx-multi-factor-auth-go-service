package middleware

import "errors"

// ErrMalformedToken signals that a malformed token has been provided
var ErrMalformedToken = errors.New("malformed token")

// ErrUnknownContentLength signals that the content length is unknown
var ErrUnknownContentLength = errors.New("unknown content length")

// ErrContentLengthTooLarge signals the content length is too large
var ErrContentLengthTooLarge = errors.New("content length too large")

// ErrMaxSizeByteTooSmall signals that the maximum byte size provided in config is too small.
var ErrMaxSizeByteTooSmall = errors.New("maximum byte size provided is too small")
