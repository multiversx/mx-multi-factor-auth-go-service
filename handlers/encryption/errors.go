package encryption

import "errors"

// ErrNilMarshaller is returned when a nil marshaller is provided
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrNilKeyGenerator is returned when a nil key generator is provided
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrNilPrivateKey is returned when a nil private key is provided
var ErrNilPrivateKey = errors.New("nil private key")
