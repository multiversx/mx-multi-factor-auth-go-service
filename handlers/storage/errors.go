package storage

import (
	"errors"

	"github.com/multiversx/mx-chain-storage-go/common"
)

// ErrKeyNotFound is returned when a key is not found in the storage
var ErrKeyNotFound = common.ErrKeyNotFound

// ErrVersionMismatch signals that the stored version differs from the expected one
var ErrVersionMismatch = errors.New("version mismatch")

// ErrKeyAlreadyExists signals that the key already exists in the storage
var ErrKeyAlreadyExists = errors.New("key already exists")
