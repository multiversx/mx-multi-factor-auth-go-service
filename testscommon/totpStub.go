package testscommon

// TotpStub -
type TotpStub struct {
	ValidateCalled func(userCode string) error
	OTPCalled      func() (string, error)
	QRCalled       func() ([]byte, error)
	ToBytesCalled  func() ([]byte, error)
	UrlCalled      func() (string, error)
}

// Validate -
func (stub *TotpStub) Validate(userCode string) error {
	if stub.ValidateCalled != nil {
		return stub.ValidateCalled(userCode)
	}
	return nil
}

// OTP -
func (stub *TotpStub) OTP() (string, error) {
	if stub.OTPCalled != nil {
		return stub.OTPCalled()
	}
	return "", nil
}

// QR -
func (stub *TotpStub) QR() ([]byte, error) {
	if stub.QRCalled != nil {
		return stub.QRCalled()
	}
	return make([]byte, 0), nil
}

// ToBytes -
func (stub *TotpStub) ToBytes() ([]byte, error) {
	if stub.ToBytesCalled != nil {
		return stub.ToBytesCalled()
	}
	return make([]byte, 0), nil
}

// Url -
func (stub *TotpStub) Url() (string, error) {
	if stub.UrlCalled != nil {
		return stub.UrlCalled()
	}
	return "", nil
}
