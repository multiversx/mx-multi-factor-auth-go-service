package storage_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/handlers/storage"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/testsCommon"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockArgs() storage.ArgDBOTPHandler {
	return storage.ArgDBOTPHandler{
		DB:          testscommon.NewMemDbMock(),
		TOTPHandler: &testsCommon.TOTPHandlerStub{},
	}
}

func TestNewDBOTPHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil db should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DB = nil
		handler, err := storage.NewDBOTPHandler(args)
		assert.Equal(t, handlers.ErrNilDB, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("nil totp handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TOTPHandler = nil
		handler, err := storage.NewDBOTPHandler(args)
		assert.Equal(t, handlers.ErrNilTOTPHandler, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("db with invalid key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DB = &mock.PersisterStub{
			RangeKeysCalled: func(handler func(key []byte, val []byte) bool) {
				handler([]byte("invalid key"), []byte("dummy val"))
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Equal(t, handlers.ErrInvalidDBKey, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("db with more than 2 guardians for account should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DB = &mock.PersisterStub{
			RangeKeysCalled: func(handler func(key []byte, val []byte) bool) {
				handler([]byte("account_guardian1"), []byte("val1"))
				handler([]byte("account_guardian2"), []byte("val2"))
				handler([]byte("account_guardian3"), []byte("val3"))
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.True(t, errors.Is(err, handlers.ErrInvalidNumberOfGuardians))
		assert.True(t, check.IfNil(handler))
	})
	t.Run("TOTPFromBytes returns error for one of the guardians from db", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DB = &mock.PersisterStub{
			RangeKeysCalled: func(handler func(key []byte, val []byte) bool) {
				handler([]byte("account_guardian1"), []byte("val1"))
			},
		}
		args.TOTPHandler = &testsCommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return nil, expectedErr
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Equal(t, expectedErr, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work with empty db", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
	t.Run("should work with existing data", func(t *testing.T) {
		t.Parallel()

		providedOTPBytes := []byte("provided otp")
		args := createMockArgs()
		{
			handler, err := storage.NewDBOTPHandler(args)
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
			handler, err := storage.NewDBOTPHandler(args)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(handler))
			assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian"))
			assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian2"))
		}
	})
}

func TestDBOTPHandler_Save(t *testing.T) {
	t.Parallel()

	t.Run("nil otp should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		err = handler.Save("account", "guardian", nil)
		assert.Equal(t, handlers.ErrNilOTP, err)
	})
	t.Run("ToBytes returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("new account but save to db fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DB = &mock.PersisterStub{
			PutCalled: func(key, val []byte) error {
				return expectedErr
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("new account should save to db", func(t *testing.T) {
		t.Parallel()

		providedOTPBytes := []byte("provided otp")
		args := createMockArgs()
		wasCalled := false
		args.DB = &mock.PersisterStub{
			PutCalled: func(key, val []byte) error {
				assert.Equal(t, []byte("account_guardian"), key)
				assert.Equal(t, providedOTPBytes, val)
				wasCalled = true
				return nil
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("old account, same guardian, same otp should return nil", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		counter := 0
		args.DB = &mock.PersisterStub{
			PutCalled: func(key, val []byte) error {
				counter++
				return nil
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
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
		assert.Equal(t, 1, counter)
	})
	t.Run("old account, same guardian, different otp save fails and does not update", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		putCounter := 0
		args.DB = &mock.PersisterStub{
			PutCalled: func(key, val []byte) error {
				putCounter++
				if putCounter > 1 {
					return expectedErr
				}
				return nil
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedNewOTPBytes := []byte("provided new otp")
		toBytesCounter := 0
		providedOTP := &testsCommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				toBytesCounter++
				if toBytesCounter > 1 {
					return providedNewOTPBytes, nil
				}
				return providedOTPBytes, nil
			},
		}
		err = handler.Save("account", "guardian", providedOTP)
		assert.Nil(t, err)
		assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian"))
		// second call, Put fails
		err = handler.Save("account", "guardian", providedOTP)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, providedOTPBytes, handler.GetEncodedOTP("account", "guardian"))
	})
	t.Run("old account, same guardian, different otp should update and save", func(t *testing.T) {
		t.Parallel()

		providedOTPBytes := []byte("provided otp")
		providedNewOTPBytes := []byte("provided new otp")
		args := createMockArgs()
		putCounter := 0
		args.DB = &mock.PersisterStub{
			PutCalled: func(key, val []byte) error {
				putCounter++
				if putCounter > 1 {
					assert.Equal(t, providedNewOTPBytes, val)
					return nil
				}
				assert.Equal(t, providedOTPBytes, val)
				return nil
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

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
	})
}

func TestDBOTPHandler_Get(t *testing.T) {
	t.Parallel()

	t.Run("missing account should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		otp, err := handler.Get("account", "guardian")
		assert.True(t, errors.Is(err, handlers.ErrNoOtpForAddress))
		assert.Nil(t, otp)
	})
	t.Run("missing guardian for account should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		handler, err := storage.NewDBOTPHandler(args)
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
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		handler, err := storage.NewDBOTPHandler(args)
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
	})
}
