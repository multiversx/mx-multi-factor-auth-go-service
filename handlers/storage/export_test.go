package storage

// GetEncodedOTP -
func (handler *dbOTPHandler) GetEncodedOTP(account, guardian string) []byte {
	handler.mut.RLock()
	defer handler.mut.RUnlock()

	return handler.encodedGuardiansOTPs[account][guardian]
}
