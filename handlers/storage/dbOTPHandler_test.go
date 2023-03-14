package storage_test

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockArgs() storage.ArgDBOTPHandler {
	return storage.ArgDBOTPHandler{
		DB:                          testscommon.NewShardedStorageWithIndexMock(),
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
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = nil
		handler, err := storage.NewDBOTPHandler(args)
		assert.Equal(t, handlers.ErrNilMarshaller, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("invalid delay should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DelayBetweenOTPUpdatesInSec = 0
		handler, err := storage.NewDBOTPHandler(args)
		assert.True(t, errors.Is(err, handlers.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "DelayBetweenOTPUpdatesInSec"))
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
	t.Run("new account but marshal fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args.DB = &testscommon.ShardedStorageWithIndexStub{
			PutCalled: func(key, data []byte) error {
				assert.Fail(t, "should have not been called")
				return nil
			},
			HasCalled: func(key []byte) error {
				return errors.New("new account")
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
	t.Run("new account but save to db fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DB = &testscommon.ShardedStorageWithIndexStub{
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
			HasCalled: func(key []byte) error {
				return errors.New("new account")
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
		args.Marshaller = &mock.MarshalizerMock{}
		args.DB = &testscommon.ShardedStorageWithIndexStub{
			PutCalled: func(key, val []byte) error {
				assert.Equal(t, []byte("guardian_account"), key)
				otpInfo := &core.OTPInfo{}
				assert.Nil(t, args.Marshaller.Unmarshal(otpInfo, val))
				assert.Equal(t, providedOTPBytes, otpInfo.OTP)
				wasCalled = true
				return nil
			},
			HasCalled: func(key []byte) error {
				return errors.New("new account")
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
	t.Run("old account, get old otp fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DB = &testscommon.ShardedStorageWithIndexStub{
			UpdateWithCheckCalled: func(key []byte, fn func(data interface{}) (interface{}, error)) error {
				_, err := args.DB.Get(key)
				return err
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args.Marshaller = &mock.MarshalizerMock{}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		err = handler.Save([]byte("account"), []byte("guardian"), &testscommon.TotpStub{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("old account, get old otp unmarshal fails", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DB = &testscommon.ShardedStorageWithIndexStub{
			UpdateWithCheckCalled: func(key []byte, fn func(data interface{}) (interface{}, error)) error {
				_, err := fn([]byte("badEncodedData"))
				return err
			},
		}
		args.Marshaller = &testscommon.MarshallerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		err = handler.Save([]byte("account"), []byte("guardian"), &testscommon.TotpStub{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("old account, same guardian, different otp fails - too early", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DelayBetweenOTPUpdatesInSec = 10
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
		args.DB = testscommon.NewShardedStorageWithIndexMock()
		args.Marshaller = &mock.MarshalizerMock{}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))

		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Nil(t, err)
		otpInfoBuff, err := args.DB.Get([]byte("guardian_account"))
		assert.Nil(t, err)
		otpInfo := &core.OTPInfo{}
		err = args.Marshaller.Unmarshal(otpInfo, otpInfoBuff)
		assert.Nil(t, err)
		assert.Equal(t, providedOTPBytes, otpInfo.OTP)
		currentTime := time.Now().Unix()
		timeDiff := currentTime - otpInfo.LastTOTPChangeTimestamp
		assert.LessOrEqual(t, timeDiff, int64(1))

		time.Sleep(time.Second)
		// second call too early fails
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.True(t, errors.Is(err, handlers.ErrRegistrationFailed))
		otpInfoBuff, err = args.DB.Get([]byte("guardian_account"))
		assert.Nil(t, err)
		err = args.Marshaller.Unmarshal(otpInfo, otpInfoBuff)
		assert.Nil(t, err)
		assert.Equal(t, providedOTPBytes, otpInfo.OTP)
		currentTime = time.Now().Unix()
		timeDiff = currentTime - otpInfo.LastTOTPChangeTimestamp
		assert.GreaterOrEqual(t, timeDiff, int64(1))
	})
	t.Run("old account, same guardian, different otp should update and save", func(t *testing.T) {
		t.Parallel()

		providedOTPBytes := []byte("provided otp")
		providedNewOTPBytes := []byte("provided new otp")
		args := createMockArgs()
		args.DelayBetweenOTPUpdatesInSec = 1
		args.DB = testscommon.NewShardedStorageWithIndexMock()
		args.Marshaller = &mock.MarshalizerMock{}
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
		otpInfoBuff, err := args.DB.Get([]byte("guardian_account"))
		assert.Nil(t, err)
		otpInfo := &core.OTPInfo{}
		err = args.Marshaller.Unmarshal(otpInfo, otpInfoBuff)
		assert.Nil(t, err)
		assert.Equal(t, providedOTPBytes, otpInfo.OTP)
		currentTime := time.Now().Unix()
		timeDiff := currentTime - otpInfo.LastTOTPChangeTimestamp
		assert.LessOrEqual(t, timeDiff, int64(1))

		time.Sleep(time.Duration(args.DelayBetweenOTPUpdatesInSec+1) * time.Second)
		err = handler.Save([]byte("account"), []byte("guardian"), providedOTP)
		assert.Nil(t, err)
		otpInfoBuff, err = args.DB.Get([]byte("guardian_account"))
		assert.Nil(t, err)
		err = args.Marshaller.Unmarshal(otpInfo, otpInfoBuff)
		assert.Nil(t, err)
		assert.Equal(t, providedNewOTPBytes, otpInfo.OTP)
		currentTime = time.Now().Unix()
		timeDiff = currentTime - otpInfo.LastTOTPChangeTimestamp
		assert.LessOrEqual(t, timeDiff, int64(1))
	})
	t.Run("multiple concurrent calls should work", func(t *testing.T) {
		if testing.Short() {
			t.Skip("this is not a short test")
		}

		t.Parallel()

		args := createMockArgs()
		args.DelayBetweenOTPUpdatesInSec = 5
		mockDB := testscommon.NewShardedStorageWithIndexMock()
		putCounter := uint32(0)
		args.DB = &testscommon.ShardedStorageWithIndexStub{
			HasCalled: func(key []byte) error {
				return mockDB.Has(key)
			},
			PutCalled: func(key, data []byte) error {
				atomic.AddUint32(&putCounter, 1)
				return mockDB.Put(key, data)
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return mockDB.Get(key)
			},
			UpdateWithCheckCalled: func(key []byte, fn func(data interface{}) (interface{}, error)) error {
				data, err := mockDB.Get(key)
				if err != nil {
					return err
				}

				newData, err := fn(data)
				if err != nil {
					return err
				}

				atomic.AddUint32(&putCounter, 1)
				return mockDB.Put(key, newData.([]byte))
			},
		}
		args.Marshaller = &mock.MarshalizerMock{}
		handler, err := storage.NewDBOTPHandler(args)
		assert.Nil(t, err)

		numCalls := 120
		var wg sync.WaitGroup
		wg.Add(numCalls)
		for i := 0; i < numCalls; i++ {
			go func() {
				defer wg.Done()
				_ = handler.Save([]byte("account"), []byte("guardian"), &testscommon.TotpStub{})
			}()
			// 50 calls/5 sec => 3 times Put called
			time.Sleep(time.Millisecond * 100)
		}

		wg.Wait()
		assert.Equal(t, uint32(3), atomic.LoadUint32(&putCounter))
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
