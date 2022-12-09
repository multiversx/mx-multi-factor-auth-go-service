package resolver

import "errors"

// ErrNilProvider signals that a nil provider was provided
var ErrNilProvider = errors.New("nil provider")

// ErrNilProxy signals that a nil proxy was provided
var ErrNilProxy = errors.New("nil proxy")

// ErrNilIndexHandler signals that a nil index handler was provided
var ErrNilIndexHandler = errors.New("nil index handler")

// ErrNilKeysGenerator signals that a nil keys generator was provided
var ErrNilKeysGenerator = errors.New("nil keys generator")

// ErrNilPubKeyConverter signals that a nil pub key converter was provided
var ErrNilPubKeyConverter = errors.New("nil pub key converter")

// ErrNilMarshaller signals that a nil marshaller was provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilHasher signals that a nil hasher was provided
var ErrNilHasher = errors.New("nil hasher")

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

// ErrNilSignatureVerifier signals that a nil signature verifier was provided
var ErrNilSignatureVerifier = errors.New("nil signature verifier")

// ErrNilGuardedTxBuilder signals that a nil guarded tx builder was provided
var ErrNilGuardedTxBuilder = errors.New("nil guarded tx builder")

// ErrNilDB signals that a nil db was provided
var ErrNilDB = errors.New("nil db")
