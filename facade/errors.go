package facade

import "github.com/go-errors/errors"

// ErrNilServiceResolver signals that a nil service resolver was provided
var ErrNilServiceResolver = errors.New("nil service resolver")
