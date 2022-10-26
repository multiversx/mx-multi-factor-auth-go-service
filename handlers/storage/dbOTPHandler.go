package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
)

const (
	maxGuardiansPerAccount = 2
	keySeparator           = "_"
)

type guardiansEncodedInfo = map[string][]byte
type guardiansOTP = map[string]handlers.OTP

// ArgDBOTPHandler is the DTO used to create a new instance of dbOTPHandler
type ArgDBOTPHandler struct {
	DB          core.Persister
	TOTPHandler handlers.TOTPHandler
}

type dbOTPHandler struct {
	db                   core.Persister
	totpHandler          handlers.TOTPHandler
	guardiansOTPs        map[string]guardiansOTP
	encodedGuardiansOTPs map[string]guardiansEncodedInfo
	mut                  sync.RWMutex
}

// NewDBOTPHandler returns a new instance of dbOTPHandler
func NewDBOTPHandler(args ArgDBOTPHandler) (*dbOTPHandler, error) {
	err := checkArgDBOTPHandler(args)
	if err != nil {
		return nil, err
	}

	handler := &dbOTPHandler{
		db:                   args.DB,
		totpHandler:          args.TOTPHandler,
		guardiansOTPs:        make(map[string]guardiansOTP),
		encodedGuardiansOTPs: make(map[string]guardiansEncodedInfo),
	}

	err = handler.loadSavedAccounts()
	if err != nil {
		return nil, err
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

	oldGuardiansOTPsForAccount, exists := handler.encodedGuardiansOTPs[account]
	if !exists {
		handler.encodedGuardiansOTPs[account] = make(guardiansEncodedInfo)
		handler.encodedGuardiansOTPs[account][guardian] = newEncodedOTP
		err = handler.db.Put(computeKey(account, guardian), newEncodedOTP)
		if err != nil {
			handler.encodedGuardiansOTPs[account] = oldGuardiansOTPsForAccount
			return err
		}
		handler.guardiansOTPs[account] = make(guardiansOTP)
		handler.guardiansOTPs[account][guardian] = otp

		return nil
	}

	oldEncodedOTP, exists := oldGuardiansOTPsForAccount[guardian]
	isSameOTP := string(oldEncodedOTP) == string(newEncodedOTP)
	if exists && isSameOTP {
		return nil
	}

	handler.encodedGuardiansOTPs[account][guardian] = newEncodedOTP
	err = handler.db.Put(computeKey(account, guardian), newEncodedOTP)
	if err != nil {
		handler.encodedGuardiansOTPs[account][guardian] = oldEncodedOTP
		return err
	}
	handler.guardiansOTPs[account][guardian] = otp

	return nil
}

// Get returns the one time password
func (handler *dbOTPHandler) Get(account, guardian string) (handlers.OTP, error) {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	guardiansOTPs, exists := handler.guardiansOTPs[account]
	if !exists {
		return nil, fmt.Errorf("%w, account %s and guardian %s", handlers.ErrNoOtpForAddress, account, guardian)
	}

	otp, exists := guardiansOTPs[guardian]
	if !exists {
		return nil, fmt.Errorf("%w, account %s and guardian %s", handlers.ErrNoOtpForGuardian, account, guardian)
	}

	return otp, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *dbOTPHandler) IsInterfaceNil() bool {
	return handler == nil
}

func (handler *dbOTPHandler) loadSavedAccounts() error {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	err := handler.readOTPsFromDB()
	if err != nil {
		return err
	}
	if len(handler.encodedGuardiansOTPs) == 0 {
		return nil
	}

	for address, guardians := range handler.encodedGuardiansOTPs {
		if len(guardians) > maxGuardiansPerAccount {
			return fmt.Errorf("%w, max expected %d", handlers.ErrInvalidNumberOfGuardians, maxGuardiansPerAccount)
		}

		handler.guardiansOTPs[address] = make(guardiansOTP)
		for guardianAddr, guardianOTP := range guardians {
			handler.guardiansOTPs[address][guardianAddr], err = handler.totpHandler.TOTPFromBytes(guardianOTP)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func computeKey(account, guardian string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", account, keySeparator, guardian))
}

func parseKey(key []byte) (string, string) {
	keyStr := string(key)
	keyComponents := strings.Split(keyStr, keySeparator)
	if len(keyComponents) != 2 {
		return "", ""
	}

	return keyComponents[0], keyComponents[1]
}

func (handler *dbOTPHandler) readOTPsFromDB() error {
	var err error
	handler.db.RangeKeys(func(key []byte, val []byte) bool {
		account, guardian := parseKey(key)
		if len(account) == 0 || len(guardian) == 0 {
			err = handlers.ErrInvalidDBKey
			return false
		}

		if len(handler.encodedGuardiansOTPs[account]) == 0 {
			handler.encodedGuardiansOTPs[account] = make(guardiansEncodedInfo)
		}
		handler.encodedGuardiansOTPs[account][guardian] = val

		return true
	})

	return err
}
