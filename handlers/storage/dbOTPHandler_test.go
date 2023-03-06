package storage_test

import (
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockArgs() storage.ArgDBOTPHandler {
	return storage.ArgDBOTPHandler{
		RegisteredUsersDB:           testscommon.NewShardedStorageWithIndexMock(),
		TOTPHandler:                 &testscommon.TOTPHandlerStub{},
		Marshaller:                  &testscommon.MarshallerStub{},
		DelayBetweenOTPUpdatesInSec: 5,
	}
}

func TestNewDBOTPHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil db should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = nil
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
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, err := storage.NewDBOTPHandler(createMockArgs())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
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

		err = handler.Save([]byte("account"), []byte("guardian"), nil)
		assert.Equal(t, handlers.ErrNilOTP, err)
	})
	t.Run("ToBytes returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTP := &testscommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("new account but save to db fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedOTP := &testscommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("new account should save to db", func(t *testing.T) {
		t.Parallel()

		providedOTPBytes := []byte("provided otp")
		args := createMockArgs()
		wasCalled := false
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
			PutCalled: func(key, val []byte) error {
				assert.Equal(t, []byte("guardian_account"), key)
				assert.Equal(t, providedOTPBytes, val)
				wasCalled = true
				return nil
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTP := &testscommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("old account, same guardian, different otp save fails and does not update", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		providedOTPBytes := []byte("provided otp")
		providedNewOTPBytes := []byte("provided new otp")
		toBytesCounter := 0
		providedOTP := &testscommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				toBytesCounter++
				if toBytesCounter > 1 {
					return providedNewOTPBytes, nil
				}
				return providedOTPBytes, nil
			},
		}
		putCounter := 0
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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

		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Nil(t, err)
		// second call, Put fails
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("old account, same guardian, different otp should update and save", func(t *testing.T) {
		t.Parallel()

		providedOTPBytes := []byte("provided otp")
		providedNewOTPBytes := []byte("provided new otp")
		args := createMockArgs()
		putCounter := 0
		args.RegisteredUsersDB = &testscommon.ShardedStorageWithIndexStub{
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
		providedOTP := &testscommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				counter++
				if counter > 1 {
					return providedNewOTPBytes, nil
				}
				return providedOTPBytes, nil
			},
		}
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Nil(t, err)
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Nil(t, err)
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

		otp, err := handler.Get([]byte("account2"), []byte("guardian"))
		assert.NotNil(t, err)
		assert.Nil(t, otp)
	})
	t.Run("missing guardian for account should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		providedOTPBytes := []byte("provided otp")
		providedOTP := &testscommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Nil(t, err)
		otp, err := handler.Get([]byte("account"), []byte("guardian2"))
		assert.NotNil(t, err)
		assert.Nil(t, otp)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		providedOTPBytes := []byte("provided otp")
		providedOTP := &testscommon.TotpStub{
			ToBytesCalled: func() ([]byte, error) {
				return providedOTPBytes, nil
			},
		}
		args.TOTPHandler = &testscommon.TOTPHandlerStub{
			TOTPFromBytesCalled: func(encryptedMessage []byte) (handlers.OTP, error) {
				return providedOTP, nil
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Nil(t, err)
		otp, err := handler.Get([]byte("account"), []byte("guardian"))
		assert.Nil(t, err)
		assert.Equal(t, providedOTP, otp)
	})
}
