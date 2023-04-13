package mongodb

import "errors"

// ErrNilMongoDBClient signals that a nil mongodb client has been provided
var ErrNilMongoDBClient = errors.New("nil mongodb client")

// ErrEmptyMongoDBName signals that an empty db name has been provided
var ErrEmptyMongoDBName = errors.New("empty db name")

// ErrCollectionNotFound signals that provided mongodb collection is not available
var ErrCollectionNotFound = errors.New("mongodb collection not found")
