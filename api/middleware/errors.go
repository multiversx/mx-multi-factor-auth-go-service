package middleware

import "errors"

// ErrMalformedToken signals that a maltformed token has been provided
var ErrMalformedToken = errors.New("malformed token")
