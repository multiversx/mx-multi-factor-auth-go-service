package groups

import "errors"

// ErrRegister signals that an error occurred while registering new user
var ErrRegister = errors.New("error register new user")

// ErrValidation signals that an error occurred while validating
var ErrValidation = errors.New("error validating")

// ErrGetGuardianAddress signals that an error occurred while getting the guardian address
var ErrGetGuardianAddress = errors.New("error get guardian address")
