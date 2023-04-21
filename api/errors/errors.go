package errors

import "errors"

// ErrNilHttpServer signals that a nil http server has been provided
var ErrNilHttpServer = errors.New("nil http server")

// ErrNilFacade signals that a nil facade has been provided
var ErrNilFacade = errors.New("nil facade")

// ErrNilNativeAuthServer signals that a nil native authentication server has been provided
var ErrNilNativeAuthServer = errors.New("nil native auth server")

// ErrNilNativeAuthWhitelistHandler signals that a nil native authentication whitelist handler has been provided
var ErrNilNativeAuthWhitelistHandler = errors.New("nil native auth whitelist handler")
