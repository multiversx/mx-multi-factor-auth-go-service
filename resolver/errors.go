package resolver

import "errors"

// ErrNilProvider signals that a nil provider was provided
var ErrNilProvider = errors.New("nil provider")

// ErrNilProxy signals that a nil proxy was provided
var ErrNilProxy = errors.New("nil proxy")

// ErrNilCredentialsHandler signals that a nil credentials handler was provided
var ErrNilCredentialsHandler = errors.New("nil credentials handler")

// ErrNilIndexHandler signals that a nil index handler was provided
var ErrNilIndexHandler = errors.New("nil index handler")

// ErrNilKeysGenerator signals that a nil keys generator was provided
var ErrNilKeysGenerator = errors.New("nil keys generator")

// ErrNilPubKeyConverter signals that a nil pub key converter was provided
var ErrNilPubKeyConverter = errors.New("nil pub key converter")

// ErrNilStorer signals that a nil storer was provided
var ErrNilStorer = errors.New("nil storer")

// ErrNilMarshaller signals that a nil marshaller was provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrInvalidGuardian signals that the given guardian is not valid
var ErrInvalidGuardian = errors.New("invalid guardian")

// ErrInvalidGuardianState signals that a guardian's state is invalid
var ErrInvalidGuardianState = errors.New("invalid guardian state")

// ErrInvalidSender signals that the given sender is not valid
var ErrInvalidSender = errors.New("invalid sender")

// ErrGuardianNotYetUsable signals that the given guardian is not yet usable
var ErrGuardianNotYetUsable = errors.New("guardian not yet usable")

// ErrGuardianMismatch signals that a guardian mismatch was detected on transactions
var ErrGuardianMismatch = errors.New("guardian mismatch")
