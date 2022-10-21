package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

const maxGuardiansPerAccount = 2

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

	return nil
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
			handler.guardiansOTPs[address][guardianAddr], err = handler.totpHandler.TOTPFromBytes(guardianOTP)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Save saves the one time password if possible, otherwise returns an error
func (handler *fileOTPHandler) Save(account, guardian string, otp OTP) error {
	newEncodedOTP, err := otp.ToBytes()
	if err != nil {
		return err
	}

	handler.mut.Lock()
	defer handler.mut.Unlock()

	oldGuardiansOTPsForAccount, exists := handler.encodedGuardiansOTPs[account]
	if !exists {
		if len(oldGuardiansOTPsForAccount) == maxGuardiansPerAccount {
			return fmt.Errorf("%w, account: %s", ErrGuardiansLimitReached, account)
		}

		handler.encodedGuardiansOTPs[account][guardian] = newEncodedOTP
		err = handler.saveOTPs()
		if err != nil {
			handler.encodedGuardiansOTPs[account] = oldGuardiansOTPsForAccount
			return err
		}
		handler.guardiansOTPs[account][guardian] = otp
	}

	oldEncodedOTP, exists := oldGuardiansOTPsForAccount[guardian]
	isSameOTP := string(oldEncodedOTP) == string(newEncodedOTP)
	if exists && isSameOTP {
		return nil
	}

	handler.encodedGuardiansOTPs[account][guardian] = newEncodedOTP
	err = handler.saveOTPs()
	if err != nil {
		handler.encodedGuardiansOTPs[account][guardian] = oldGuardiansOTPsForAccount[guardian]
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

func (handler *fileOTPHandler) readOTPs() error {
	data, err := os.ReadFile(fmt.Sprintf("%s.json", handler.fileName))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &handler.encodedGuardiansOTPs)
}

func (handler *fileOTPHandler) saveOTPs() error {
	filePath := fmt.Sprintf("%s.json", handler.fileName)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0777)
	defer closeFile(file)
	if err != nil {
		log.Println(err)
		return err
	}

	jsonOtps, err := json.Marshal(handler.encodedGuardiansOTPs)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = file.Write(jsonOtps)
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
