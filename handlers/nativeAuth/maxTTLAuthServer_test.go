package nativeAuth_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-storage-go/lrucache"
	"github.com/multiversx/mx-sdk-go/authentication"
	"github.com/multiversx/mx-sdk-go/authentication/native/mock"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/nativeAuth"
)

const (
	blockHash      = "hash"
	blockTimestamp = int64(1000)
	maxTTL         = int64(600)
)

var expectedErr = errors.New("expected error")

type httpClientWrapperStub struct {
	getHTTPCalled func(ctx context.Context, endpoint string) ([]byte, int, error)
}

func (stub *httpClientWrapperStub) GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error) {
	if stub.getHTTPCalled != nil {
		return stub.getHTTPCalled(ctx, endpoint)
	}

	return []byte(`{"timestamp":1000}`), http.StatusOK, nil
}

func (stub *httpClientWrapperStub) PostHTTP(_ context.Context, _ string, _ []byte) ([]byte, int, error) {
	return nil, http.StatusOK, nil
}

func (stub *httpClientWrapperStub) IsInterfaceNil() bool {
	return stub == nil
}

func createMockArgs(t *testing.T) nativeAuth.ArgsMaxTTLAuthServer {
	cacher, err := lrucache.NewCache(10)
	require.Nil(t, err)

	return nativeAuth.ArgsMaxTTLAuthServer{
		AuthServer:        &mock.AuthServerStub{},
		HttpClientWrapper: &httpClientWrapperStub{},
		TimestampsCacher:  cacher,
		MaxTTLInSec:       maxTTL,
	}
}

func createToken(ttl int64) *mock.AuthTokenStub {
	return &mock.AuthTokenStub{
		GetTtlCalled: func() int64 {
			return ttl
		},
		GetBlockHashCalled: func() string {
			return blockHash
		},
	}
}

func createServerAt(t *testing.T, args nativeAuth.ArgsMaxTTLAuthServer, now int64) authentication.AuthServer {
	server, err := nativeAuth.NewMaxTTLAuthServer(args)
	require.Nil(t, err)
	server.SetTimeHandler(func() time.Time {
		return time.Unix(now, 0)
	})

	return server
}

func TestNewMaxTTLAuthServer(t *testing.T) {
	t.Parallel()

	t.Run("nil auth server should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.AuthServer = nil
		server, err := nativeAuth.NewMaxTTLAuthServer(args)
		require.Equal(t, nativeAuth.ErrNilAuthServer, err)
		require.True(t, check.IfNil(server))
	})
	t.Run("nil http client wrapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.HttpClientWrapper = nil
		server, err := nativeAuth.NewMaxTTLAuthServer(args)
		require.Equal(t, authentication.ErrNilHttpClientWrapper, err)
		require.True(t, check.IfNil(server))
	})
	t.Run("nil cacher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.TimestampsCacher = nil
		server, err := nativeAuth.NewMaxTTLAuthServer(args)
		require.Equal(t, authentication.ErrNilCacher, err)
		require.True(t, check.IfNil(server))
	})
	t.Run("invalid max ttl should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.MaxTTLInSec = 0
		server, err := nativeAuth.NewMaxTTLAuthServer(args)
		require.True(t, errors.Is(err, nativeAuth.ErrInvalidValue))
		require.True(t, check.IfNil(server))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		server, err := nativeAuth.NewMaxTTLAuthServer(createMockArgs(t))
		require.Nil(t, err)
		require.False(t, check.IfNil(server))
	})
}

func TestMaxTTLAuthServer_Validate(t *testing.T) {
	t.Parallel()

	t.Run("wrapped auth server error should be returned", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.AuthServer = &mock.AuthServerStub{
			ValidateCalled: func(accessToken authentication.AuthToken) error {
				return expectedErr
			},
		}
		server := createServerAt(t, args, blockTimestamp)
		require.Equal(t, expectedErr, server.Validate(createToken(maxTTL)))
	})
	t.Run("token ttl within max ttl should rely on the wrapped auth server", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.HttpClientWrapper = &httpClientWrapperStub{
			getHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Fail(t, "should not have been called")
				return nil, 0, nil
			},
		}
		server := createServerAt(t, args, blockTimestamp+maxTTL*10)
		require.Nil(t, server.Validate(createToken(maxTTL)))
	})
	t.Run("token ttl higher than max ttl should expire after max ttl", func(t *testing.T) {
		t.Parallel()

		token := createToken(365 * 24 * 3600)
		args := createMockArgs(t)

		server := createServerAt(t, args, blockTimestamp+maxTTL)
		require.Nil(t, server.Validate(token))

		server = createServerAt(t, args, blockTimestamp+maxTTL+1)
		err := server.Validate(token)
		require.True(t, errors.Is(err, authentication.ErrTokenExpired))
	})
	t.Run("cached block timestamp should be used", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.TimestampsCacher.Put([]byte(blockHash), blockTimestamp, 8)
		args.HttpClientWrapper = &httpClientWrapperStub{
			getHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Fail(t, "should not have been called")
				return nil, 0, nil
			},
		}
		server := createServerAt(t, args, blockTimestamp+maxTTL+1)
		err := server.Validate(createToken(maxTTL + 1))
		require.True(t, errors.Is(err, authentication.ErrTokenExpired))
	})
	t.Run("block fetched should be cached", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		numCalls := 0
		args.HttpClientWrapper = &httpClientWrapperStub{
			getHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				numCalls++
				require.Equal(t, "blocks/"+blockHash, endpoint)
				return []byte(`{"timestamp":1000}`), http.StatusOK, nil
			},
		}
		server := createServerAt(t, args, blockTimestamp)
		require.Nil(t, server.Validate(createToken(maxTTL+1)))
		require.Nil(t, server.Validate(createToken(maxTTL+1)))
		require.Equal(t, 1, numCalls)
	})
	t.Run("invalid cached value should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.TimestampsCacher.Put([]byte(blockHash), "invalid", 8)
		server := createServerAt(t, args, blockTimestamp)
		err := server.Validate(createToken(maxTTL + 1))
		require.True(t, errors.Is(err, authentication.ErrInvalidValue))
	})
	t.Run("block request failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.HttpClientWrapper = &httpClientWrapperStub{
			getHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return nil, http.StatusNotFound, nil
			},
		}
		server := createServerAt(t, args, blockTimestamp)
		require.NotNil(t, server.Validate(createToken(maxTTL+1)))
	})
	t.Run("invalid block response should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.HttpClientWrapper = &httpClientWrapperStub{
			getHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return []byte("invalid"), http.StatusOK, nil
			},
		}
		server := createServerAt(t, args, blockTimestamp)
		require.NotNil(t, server.Validate(createToken(maxTTL+1)))
	})
}
