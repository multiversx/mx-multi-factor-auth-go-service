package mongodb

import "errors"

// ErrNilMongoDBClientWrapper signals that a nil mongodb client wrapper has been provided
var ErrNilMongoDBClientWrapper = errors.New("nil mongodb client wrapper")

// ErrEmptyMongoDBName signals that an empty db name has been provided
var ErrEmptyMongoDBName = errors.New("empty db name")

// ErrCollectionNotFound signals that provided mongodb collection is not available
var ErrCollectionNotFound = errors.New("mongodb collection not found")
