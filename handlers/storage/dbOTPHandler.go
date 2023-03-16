package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	keySeparator              = "_"
	minDelayBetweenOTPUpdates = 1
)

// ArgDBOTPHandler is the DTO used to create a new instance of dbOTPHandler
type ArgDBOTPHandler struct {
	DB                          core.ShardedStorageWithIndex
	TOTPHandler                 handlers.TOTPHandler
	Marshaller                  core.Marshaller
	DelayBetweenOTPUpdatesInSec int64
}

type dbOTPHandler struct {
	db                          core.ShardedStorageWithIndex
	totpHandler                 handlers.TOTPHandler
	marshaller                  core.Marshaller
	getTimeHandler              func() time.Time
	delayBetweenOTPUpdatesInSec int64
	mut                         sync.RWMutex
}

// NewDBOTPHandler returns a new instance of dbOTPHandler
func NewDBOTPHandler(args ArgDBOTPHandler) (*dbOTPHandler, error) {
	err := checkArgDBOTPHandler(args)
	if err != nil {
		return nil, err
	}

	handler := &dbOTPHandler{
		db:                          args.DB,
		totpHandler:                 args.TOTPHandler,
		getTimeHandler:              time.Now,
		marshaller:                  args.Marshaller,
		delayBetweenOTPUpdatesInSec: args.DelayBetweenOTPUpdatesInSec,
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
	if check.IfNil(args.Marshaller) {
		return handlers.ErrNilMarshaller
	}
	if args.DelayBetweenOTPUpdatesInSec < minDelayBetweenOTPUpdates {
		return fmt.Errorf("%w for DelayBetweenOTPUpdatesInSec, got %d, min expected %d",
			handlers.ErrInvalidValue, args.DelayBetweenOTPUpdatesInSec, minDelayBetweenOTPUpdates)
	}

	return nil
}

// Save saves the one time password if possible, otherwise returns an error
func (handler *dbOTPHandler) Save(account, guardian []byte, otp handlers.OTP) error {
	if otp == nil {
		return handlers.ErrNilOTP
	}

	key := computeKey(account, guardian)

	// critical section, do not allow a second Put until this is done
	handler.mut.Lock()
	defer handler.mut.Unlock()

	err := handler.db.Has(key)
	if err != nil {
		return handler.saveNewOTP(key, otp)
	}

	oldOTPInfo, err := handler.getOldOTPInfo(key)
	if err != nil {
		return err
	}

	currentTimestamp := handler.getTimeHandler().Unix()
	isOTPUpdateAllowed := oldOTPInfo.LastTOTPChangeTimestamp+handler.delayBetweenOTPUpdatesInSec < currentTimestamp
	if !isOTPUpdateAllowed {
		return fmt.Errorf("%w, last update was %d seconds ago",
			handlers.ErrRegistrationFailed, currentTimestamp-oldOTPInfo.LastTOTPChangeTimestamp)
	}

	return handler.saveNewOTP(key, otp)
}

// Get returns the one time password
func (handler *dbOTPHandler) Get(account, guardian []byte) (handlers.OTP, error) {
	key := computeKey(account, guardian)
	oldOTPInfo, err := handler.getOldOTPInfo(key)
	if err != nil {
		accountAddr := data.NewAddressFromBytes(account)
		guardianAddr := data.NewAddressFromBytes(guardian)
		return nil, fmt.Errorf("%w, account=%s, guardian=%s",
			err, accountAddr.AddressAsBech32String(), guardianAddr.AddressAsBech32String())
	}

	return handler.totpHandler.TOTPFromBytes(oldOTPInfo.OTP)
}

func (handler *dbOTPHandler) getOldOTPInfo(key []byte) (*core.OTPInfo, error) {
	oldOTPInfo, err := handler.db.Get(key)
	if err != nil {
		return nil, err
	}

	otpInfo := &core.OTPInfo{}
	err = handler.marshaller.Unmarshal(otpInfo, oldOTPInfo)
	if err != nil {
		return nil, err
	}

	return otpInfo, nil
}

func (handler *dbOTPHandler) newOTPInfo(otp handlers.OTP) (*core.OTPInfo, error) {
	otpInfo := &core.OTPInfo{
		LastTOTPChangeTimestamp: handler.getTimeHandler().Unix(),
	}

	var err error
	otpInfo.OTP, err = otp.ToBytes()
	if err != nil {
		return nil, err
	}

	return otpInfo, nil
}

func (handler *dbOTPHandler) saveNewOTP(key []byte, otp handlers.OTP) error {
	otpInfo, err := handler.newOTPInfo(otp)
	if err != nil {
		return err
	}

	return handler.marshalAndSaveOTPInfo(key, otpInfo)
}

func (handler *dbOTPHandler) marshalAndSaveOTPInfo(key []byte, otpInfo *core.OTPInfo) error {
	buff, err := handler.marshaller.Marshal(otpInfo)
	if err != nil {
		return err
	}

	return handler.db.Put(key, buff)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *dbOTPHandler) IsInterfaceNil() bool {
	return handler == nil
}

func computeKey(account, guardian []byte) []byte {
	return []byte(fmt.Sprintf("%s%s%s", guardian, keySeparator, account))
}
