package middleware

import "errors"

// ErrMalformedToken signals that a malformed token has been provided
var ErrMalformedToken = errors.New("malformed token")
