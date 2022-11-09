package handlers_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockArgs(t *testing.T) handlers.ArgFileOTPHandler {
	return handlers.ArgFileOTPHandler{
		FileName:    "file_name_" + strings.Replace(t.Name(), "/", "", -1),
		TOTPHandler: &testsCommon.TOTPHandlerStub{},
	}
}

func TestNewFileOTPHandler(t *testing.T) {
	t.Parallel()

	t.Run("empty file name should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.FileName = ""
		handler, err := handlers.NewFileOTPHandler(args)
		assert.True(t, errors.Is(err, handlers.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "file name"))
		assert.True(t, check.IfNil(handler))
	})
	t.Run("nil totp handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.TOTPHandler = nil
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Equal(t, handlers.ErrNilTOTPHandler, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work with no file", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
	t.Run("should work with existing empty file", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		fileName := fmt.Sprintf("%s.json", args.FileName)
		_, err := os.Create(fileName)
		assert.Nil(t, err)

		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		assert.Nil(t, os.Remove(fileName))
	})
	t.Run("should work with existing data", func(t *testing.T) {
		t.Parallel()

		providedOTPBytes := []byte("provided otp")
		args := createMockArgs(t)
		{
			handler, err := handlers.NewFileOTPHandler(args)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(handler))

			providedOTP := &testsCommon.TotpStub{
				ToBytesCalled: func() ([]byte, error) {
					return providedOTPBytes, nil
				},
			}
			err = handler.Save("account", "guardian", providedOTP)
			assert.Nil(t, err)
			err = handler.Save("account", "guardian2", providedOTP)
			assert.Nil(t, err)
		}
		{
			handler, err := handlers.NewFileOTPHandler(args)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(handler))
			assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian"))
			assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian2"))
		}
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
}

func TestFileOTPHandler_Save(t *testing.T) {
	t.Parallel()

	t.Run("nil otp should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		err = handler.Save("account", "guardian", nil)
		assert.Equal(t, handlers.ErrNilOTP, err)
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
	t.Run("ToBytes returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
	t.Run("new account but save to file fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		// remove the file so Save fails
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
		err = handler.Save("account", "guardian", providedOTP)
		assert.NotNil(t, err)
	})
	t.Run("new account should save to file", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
	t.Run("old account, same guardian, same otp should return nil", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
	t.Run("old account, same guardian, different otp save fails and does not update", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedNewOTPBytes := []byte("provided new otp")
		counter := 0
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				counter++
				if counter > 1 {
					return providedNewOTPBytes, nil
				}
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		// remove the file so Save fails
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
		assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian"))
		err = handler.Save("account", "guardian", providedOTP)
		assert.NotNil(t, err)
		assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian"))
	})
	t.Run("old account, same guardian, different otp should update and save", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedNewOTPBytes := []byte("provided new otp")
		counter := 0
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				counter++
				if counter > 1 {
					return providedNewOTPBytes, nil
				}
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian"))
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		assert.Equal(t, providedNewOTPBytes, handler.GetEncodedOTP("account", "guardian"))
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
}

func TestFileOTPHandler_Get(t *testing.T) {
	t.Parallel()

	t.Run("missing account should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		otp, err := handler.Get("account", "guardian")
		assert.True(t, errors.Is(err, handlers.ErrNoOtpForAddress))
		assert.Nil(t, otp)
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
	t.Run("missing guardian for account should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		otp, err := handler.Get("account", "guardian2")
		assert.True(t, errors.Is(err, handlers.ErrNoOtpForGuardian))
		assert.Nil(t, otp)
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		handler, err := handlers.NewFileOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		otp, err := handler.Get("account", "guardian")
		assert.Nil(t, err)
		assert.NotNil(t, otp)
		assert.Nil(t, os.Remove(fmt.Sprintf("%s.json", args.FileName)))
	})
}
