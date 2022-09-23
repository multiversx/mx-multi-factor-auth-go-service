package providers

import (
	"crypto"
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func TestTimebasedOnetimePassword_NewTimebasedOnetimePassword(t *testing.T) {
	t.Parallel()

	totp := NewTimebasedOnetimePassword("issuer", 6)
	require.False(t, totp.IsInterfaceNil())
}
func TestTimebasedOnetimePassword_Validate(t *testing.T) {
	t.Parallel()

	t.Run("account does not exists", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		isValid, err := totp.Validate("addr1", "1234")
		require.False(t, isValid)
		assert.True(t, errors.Is(err, ErrNoOtpForAddress))
		assert.True(t, strings.Contains(err.Error(), "addr1"))
	})
	t.Run("code not valid for otp", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.saveOtpHandle = createSaveOtpHandle(nil)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		totp.otps["addr1"] = &testsCommon.TotpStub{
			ValidateCalled: func(userCode string) error {
				return expectedErr
			},
		}
		isValid, err := totp.Validate("addr1", "1234")
		require.False(t, isValid)
		assert.True(t, errors.Is(err, ErrInvalidCode))
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
	t.Run("cannot save information", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.saveOtpHandle = createSaveOtpHandle(expectedErr)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		totp.otps["addr1"] = &testsCommon.TotpStub{
			ValidateCalled: func(userCode string) error {
				return expectedErr
			},
		}
		isValid, err := totp.Validate("addr1", "1234")
		require.False(t, isValid)
		assert.True(t, errors.Is(err, ErrCannotUpdateInformation))
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.saveOtpHandle = createSaveOtpHandle(nil)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		totp.otps["addr1"] = &testsCommon.TotpStub{
			ValidateCalled: func(userCode string) error {
				return nil
			},
		}
		isValid, err := totp.Validate("addr1", "1234")
		require.True(t, isValid)
		assert.Nil(t, err)
	})
}

func createMockArgs() ArgsTimebasedOnetimePasswordWithHandler {
	return ArgsTimebasedOnetimePasswordWithHandler{
		issuer:              "issuer",
		digits:              6,
		createNewOtpHandle:  nil,
		totpFromBytesHandle: nil,
		readOtpsHandle:      nil,
		saveOtpHandle:       nil,
	}

}

func TestTimebasedOnetimePassword_RegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("createNewOtpHandle returns error shall error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.createNewOtpHandle = createCreateNewOtpHandle(nil, expectedErr)
		args.saveOtpHandle = createSaveOtpHandle(nil)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		qr, err := totp.RegisterUser("addr1")
		require.Nil(t, qr)
		require.Equal(t, expectedErr, err)
	})
	t.Run("otp.QR returns error shall error", func(t *testing.T) {
		t.Parallel()

		createNewOtpHandle := createCreateNewOtpHandle(&testsCommon.TotpStub{
			QRCalled: func() ([]byte, error) {
				return make([]byte, 0), expectedErr
			},
		}, nil)

		args := createMockArgs()
		args.createNewOtpHandle = createNewOtpHandle
		args.saveOtpHandle = createSaveOtpHandle(nil)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		qr, err := totp.RegisterUser("addr1")
		require.Nil(t, qr)
		require.Equal(t, expectedErr, err)
	})
	t.Run("cannot save otp", func(t *testing.T) {
		t.Parallel()

		createNewOtpHandle := createCreateNewOtpHandle(&testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return make([]byte, 0), expectedErr
			},
		}, nil)
		args := createMockArgs()
		args.createNewOtpHandle = createNewOtpHandle
		args.saveOtpHandle = createSaveOtpHandle(nil)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		qr, err := totp.RegisterUser("addr1")
		require.Equal(t, 0, len(qr))
		assert.True(t, errors.Is(err, ErrCannotUpdateInformation))
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedQrByte := []byte("qrCode")
		createNewOtpHandle := createCreateNewOtpHandle(&testsCommon.TotpStub{
			QRCalled: func() ([]byte, error) {
				return expectedQrByte, nil
			},
		}, nil)
		args := createMockArgs()
		args.createNewOtpHandle = createNewOtpHandle
		args.saveOtpHandle = createSaveOtpHandle(nil)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		qr, err := totp.RegisterUser("addr1")
		assert.Equal(t, qr, expectedQrByte)
		assert.Nil(t, err)
	})
}

func TestTimebasedOnetimePassword_update(t *testing.T) {
	t.Parallel()

	t.Run("revert to old otp", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		args.saveOtpHandle = createSaveOtpHandle(expectedErr)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		addr := "addr"
		expectedOtp := []byte("otp")
		totp.otpsEncoded[addr] = expectedOtp

		err := totp.update(addr, &testsCommon.TotpStub{})
		require.Equal(t, expectedErr, err)
		require.Equal(t, expectedOtp, totp.otpsEncoded[addr])

	})
}

func TestTimebasedOnetimePassword_LoadSavedAccounts(t *testing.T) {
	t.Parallel()

	t.Run("readOtpsHandle returns error shall error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.readOtpsHandle = createReadOtpsHandle(make(map[string][]byte, 0), expectedErr)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		err := totp.LoadSavedAccounts()
		require.Equal(t, expectedErr, err)
	})
	t.Run("readOtpsHandle returns nil shall create new map", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.readOtpsHandle = createReadOtpsHandle(nil, nil)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		err := totp.LoadSavedAccounts()
		require.Nil(t, err)
		require.Equal(t, 0, len(totp.otps))
		require.Equal(t, 0, len(totp.otpsEncoded))
	})
	t.Run("totpFromBytesHandle returns error shall error", func(t *testing.T) {
		t.Parallel()

		m := make(map[string][]byte)
		m["addr"] = []byte("bytes")

		args := createMockArgs()
		args.totpFromBytesHandle = createTotpFromBytesHandle(nil, expectedErr)
		args.readOtpsHandle = createReadOtpsHandle(m, nil)
		totp := NewTimebasedOnetimePasswordWithHandler(args)

		err := totp.LoadSavedAccounts()
		require.Equal(t, expectedErr, err)
	})
}

func createCreateNewOtpHandle(totp Totp, err error) func(account, issuer string, hash crypto.Hash, digits int) (Totp, error) {
	return func(account, issuer string, hash crypto.Hash, digits int) (Totp, error) {
		return totp, err
	}
}

func createTotpFromBytesHandle(totp Totp, err error) func(encryptedMessage []byte, issuer string) (Totp, error) {
	return func(encryptedMessage []byte, issuer string) (Totp, error) {
		return totp, err
	}
}

func createReadOtpsHandle(m map[string][]byte, err error) func(filename string) (map[string][]byte, error) {
	return func(filename string) (map[string][]byte, error) {
		return m, err
	}
}

func createSaveOtpHandle(err error) func(filename string, otps map[string][]byte) error {
	return func(filename string, otps map[string][]byte) error {
		return err
	}
}