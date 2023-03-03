package storage

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const (
	keySeparator = "_"
)

// ArgDBOTPHandler is the DTO used to create a new instance of dbOTPHandler
type ArgDBOTPHandler struct {
	RegisteredUsersDB core.ShardedStorageWithIndex
	TOTPHandler       handlers.TOTPHandler
}

type dbOTPHandler struct {
	registeredUsersDB core.ShardedStorageWithIndex
	totpHandler       handlers.TOTPHandler
}

// NewDBOTPHandler returns a new instance of dbOTPHandler
func NewDBOTPHandler(args ArgDBOTPHandler) (*dbOTPHandler, error) {
	err := checkArgDBOTPHandler(args)
	if err != nil {
		return nil, err
	}

	handler := &dbOTPHandler{
		registeredUsersDB: args.RegisteredUsersDB,
		totpHandler:       args.TOTPHandler,
	}

	return handler, nil
}

func checkArgDBOTPHandler(args ArgDBOTPHandler) error {
	if check.IfNil(args.RegisteredUsersDB) {
		return handlers.ErrNilDB
	}
	if check.IfNil(args.TOTPHandler) {
		return handlers.ErrNilTOTPHandler
	}

	return nil
}

// Save saves the one time password if possible, otherwise returns an error
func (handler *dbOTPHandler) Save(account, guardian string, otp handlers.OTP) error {
	if otp == nil {
		return handlers.ErrNilOTP
	}

	newEncodedOTP, err := otp.ToBytes()
	if err != nil {
		return err
	}

	key := computeKey(account, guardian)
	return handler.registeredUsersDB.Put(key, newEncodedOTP)
}

// Get returns the one time password
func (handler *dbOTPHandler) Get(account, guardian string) (handlers.OTP, error) {
	key := computeKey(account, guardian)
	oldEncodedOTP, err := handler.registeredUsersDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("%w, account %s and guardian %s", err, account, guardian)
	}

	return handler.totpHandler.TOTPFromBytes(oldEncodedOTP)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *dbOTPHandler) IsInterfaceNil() bool {
	return handler == nil
}

func computeKey(account, guardian string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", account, keySeparator, guardian))
}
