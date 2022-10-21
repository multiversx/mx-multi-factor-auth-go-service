package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

const (
	maxGuardiansPerAccount = 2
	minFileNameLength      = 1
)

type guardiansEncodedInfo = map[string][]byte
type guardiansOTP = map[string]OTP

// ArgFileOTPHandler is the DTO used to create a new instance of fileOTPHandler
type ArgFileOTPHandler struct {
	FileName    string
	TOTPHandler TOTPHandler
}

type fileOTPHandler struct {
	fileName             string
	totpHandler          TOTPHandler
	guardiansOTPs        map[string]guardiansOTP
	encodedGuardiansOTPs map[string]guardiansEncodedInfo
	mut                  sync.RWMutex
}

// NewFileOTPHandler returns a new instance of fileOTPHandler
func NewFileOTPHandler(args ArgFileOTPHandler) (*fileOTPHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	handler := &fileOTPHandler{
		fileName:             args.FileName,
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

func checkArgs(args ArgFileOTPHandler) error {
	if len(args.FileName) < minFileNameLength {
		return fmt.Errorf("%w for file name", ErrInvalidValue)
	}
	if check.IfNil(args.TOTPHandler) {
		return ErrNilTOTPHandler
	}
	return nil
}

// Save saves the one time password if possible, otherwise returns an error
func (handler *fileOTPHandler) Save(account, guardian string, otp OTP) error {
	if otp == nil {
		return ErrNilOTP
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
		err = handler.saveOTPs()
		if err != nil {
			handler.encodedGuardiansOTPs[account] = oldGuardiansOTPsForAccount
			return err
		}
		handler.guardiansOTPs[account] = make(guardiansOTP)
		handler.guardiansOTPs[account][guardian] = otp
	}

	if len(oldGuardiansOTPsForAccount) == maxGuardiansPerAccount {
		return fmt.Errorf("%w, account: %s", ErrGuardiansLimitReached, account)
	}

	oldEncodedOTP, exists := oldGuardiansOTPsForAccount[guardian]
	isSameOTP := string(oldEncodedOTP) == string(newEncodedOTP)
	if exists && isSameOTP {
		return nil
	}

	handler.encodedGuardiansOTPs[account][guardian] = newEncodedOTP
	err = handler.saveOTPs()
	if err != nil {
		handler.encodedGuardiansOTPs[account][guardian] = oldEncodedOTP
		return err
	}
	handler.guardiansOTPs[account][guardian] = otp

	return nil
}

// Get returns the one time password
func (handler *fileOTPHandler) Get(account, guardian string) (OTP, error) {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	guardiansOTPs, exists := handler.guardiansOTPs[account]
	if !exists {
		return nil, fmt.Errorf("%w, account %s and guardian %s", ErrNoOtpForAddress, account, guardian)
	}

	otp, exists := guardiansOTPs[guardian]
	if !exists {
		return nil, fmt.Errorf("%w, account %s and guardian %s", ErrNoOtpForGuardian, account, guardian)
	}

	return otp, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *fileOTPHandler) IsInterfaceNil() bool {
	return handler == nil
}

func (handler *fileOTPHandler) loadSavedAccounts() error {
	handler.mut.Lock()
	defer handler.mut.Unlock()

	err := handler.readOTPs()
	if err != nil {
		return err
	}
	if handler.encodedGuardiansOTPs == nil {
		return nil
	}

	for address, guardians := range handler.encodedGuardiansOTPs {
		if len(guardians) > maxGuardiansPerAccount {
			return fmt.Errorf("%w, max expected %d", ErrInvalidNumberOfGuardians, maxGuardiansPerAccount)
		}

		for guardianAddr, guardianOTP := range guardians {
			handler.guardiansOTPs[address] = make(guardiansOTP)
			handler.guardiansOTPs[address][guardianAddr], err = handler.totpHandler.TOTPFromBytes(guardianOTP)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (handler *fileOTPHandler) readOTPs() error {
	fileName := fmt.Sprintf("%s.json", handler.fileName)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	defer closeFile(file)
	if err != nil {
		log.Println(err)
		return err
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &handler.encodedGuardiansOTPs)
}

func (handler *fileOTPHandler) saveOTPs() error {
	filePath := fmt.Sprintf("%s.json", handler.fileName)
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0777)
	defer closeFile(file)
	if err != nil {
		log.Println(err)
		return err
	}

	jsonOTPs, err := json.Marshal(handler.encodedGuardiansOTPs)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = file.Write(jsonOTPs)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Println(err)
	}
}
