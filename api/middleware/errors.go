package middleware

import "errors"

// ErrMalformedToken signals that a malformed token has been provided
var ErrMalformedToken = errors.New("malformed token")

// ErrNilStatusMetricsExtractor signals that a nil status metrics extractor has been provided
var ErrNilStatusMetricsExtractor = errors.New("nil status metrics extractor")
