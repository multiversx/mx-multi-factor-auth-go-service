package storage

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
)

const (
	keySeparator = "_"
)

// ArgDBOTPHandler is the DTO used to create a new instance of dbOTPHandler
type ArgDBOTPHandler struct {
	DB          core.Storer
	TOTPHandler handlers.TOTPHandler
}

type dbOTPHandler struct {
	db          core.Storer
	totpHandler handlers.TOTPHandler
	mut         sync.RWMutex
}

// NewDBOTPHandler returns a new instance of dbOTPHandler
func NewDBOTPHandler(args ArgDBOTPHandler) (*dbOTPHandler, error) {
	err := checkArgDBOTPHandler(args)
	if err != nil {
		return nil, err
	}

	handler := &dbOTPHandler{
		db:          args.DB,
		totpHandler: args.TOTPHandler,
	}

	return handler, nil
}

func checkArgDBOTPHandler(args ArgDBOTPHandler) error {
	if check.IfNil(args.DB) {
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

	handler.mut.Lock()
	defer handler.mut.Unlock()

	key := computeKey(account, guardian)
	oldEncodedOTP, err := handler.db.Get(key)
	if err != nil {
		return handler.db.Put(computeKey(account, guardian), newEncodedOTP)
	}

	if string(oldEncodedOTP) == string(newEncodedOTP) {
		return nil
	}

	return handler.db.Put(computeKey(account, guardian), newEncodedOTP)
}

// Get returns the one time password
func (handler *dbOTPHandler) Get(account, guardian string) (handlers.OTP, error) {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	key := computeKey(account, guardian)
	oldEncodedOTP, err := handler.db.Get(key)
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
