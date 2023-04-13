package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-sdk-go/authentication"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	sdkTestsCommon "github.com/multiversx/mx-sdk-go/testsCommon"
	"github.com/stretchr/testify/require"
)

var (
	expectedErr                     = errors.New("expected error")
	providedAddress                 = "provided address"
	expectedGetAccountEndpoint      = fmt.Sprintf("address/%s", providedAddress)
	expectedGetGuardianDataEndpoint = fmt.Sprintf("address/%s/guardian-data", providedAddress)
)

func TestNewHttpClientWrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil http client should error", func(t *testing.T) {
		t.Parallel()

		wrapper, err := NewHttpClientWrapper(nil)
		require.Equal(t, ErrNilHttpClient, err)
		require.Nil(t, wrapper)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wrapper, err := NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{})
		require.NoError(t, err)
		require.NotNil(t, wrapper)
	})
}

func TestHttpClientWrapper_GetAccount(t *testing.T) {
	t.Parallel()

	t.Run("GetHTTP returns error should error", testGetHTTPReturnsError(expectedGetAccountEndpoint))
	t.Run("GetHTTP returns error status code should error", testGetHTTPReturnsStatusCodeError(expectedGetAccountEndpoint))
	t.Run("GetHTTP returns nil data should error", testGetHTTPReturnsNilData(expectedGetAccountEndpoint))
	t.Run("Unmarshal fails should error", testUnmarshalFails(expectedGetAccountEndpoint))

	buff, _ := json.Marshal(&sdkData.AccountResponse{
		Data: struct {
			Account *sdkData.Account `json:"account"`
		}{
			Account: nil,
		},
	})
	t.Run("api returns nil data should error", testGetHTTPReturnsNilUnmarshalledData(expectedGetAccountEndpoint, buff))
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wrapper, _ := NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Equal(t, expectedGetAccountEndpoint, endpoint)
				buff, _ = json.Marshal(&sdkData.AccountResponse{
					Data: struct {
						Account *sdkData.Account `json:"account"`
					}{
						Account: &sdkData.Account{},
					},
				})
				return buff, 200, nil
			},
		})
		require.NotNil(t, wrapper)

		account, err := wrapper.GetAccount(context.Background(), providedAddress)
		require.NoError(t, err)
		require.Equal(t, &sdkData.Account{}, account)
	})
}

func TestHttpClientWrapper_GetGuardianData(t *testing.T) {
	t.Parallel()

	t.Run("GetHTTP returns error should error", testGetHTTPReturnsError(expectedGetGuardianDataEndpoint))
	t.Run("GetHTTP returns error status code should error", testGetHTTPReturnsStatusCodeError(expectedGetGuardianDataEndpoint))
	t.Run("GetHTTP returns nil data should error", testGetHTTPReturnsNilData(expectedGetGuardianDataEndpoint))
	t.Run("Unmarshal fails should error", testUnmarshalFails(expectedGetGuardianDataEndpoint))

	buff, _ := json.Marshal(&sdkData.GuardianDataResponse{
		Data: struct {
			GuardianData *api.GuardianData `json:"guardianData"`
		}{
			GuardianData: nil,
		},
	})
	t.Run("api returns nil data should error", testGetHTTPReturnsNilUnmarshalledData(expectedGetGuardianDataEndpoint, buff))
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedData := &api.GuardianData{
			ActiveGuardian: &api.Guardian{
				Address: "active guardian",
			},
			PendingGuardian: &api.Guardian{
				Address: "pending guardian",
			},
		}
		wrapper, _ := NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Equal(t, expectedGetGuardianDataEndpoint, endpoint)
				buff, _ = json.Marshal(&sdkData.GuardianDataResponse{
					Data: struct {
						GuardianData *api.GuardianData `json:"guardianData"`
					}{
						GuardianData: providedData,
					},
				})
				return buff, 200, nil
			},
		})
		require.NotNil(t, wrapper)

		guardianData, err := wrapper.GetGuardianData(context.Background(), providedAddress)
		require.NoError(t, err)
		require.Equal(t, providedData, guardianData)
	})
}

func TestHttpClientWrapper_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var wrapper *httpClientWrapper
	require.True(t, wrapper.IsInterfaceNil())

	wrapper, _ = NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{})
	require.False(t, wrapper.IsInterfaceNil())
}

func testGetHTTPReturnsError(expectedEndpoint string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		wrapper, _ := NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Equal(t, expectedEndpoint, endpoint)
				return nil, 0, expectedErr
			},
		})
		require.NotNil(t, wrapper)

		if strings.Contains(expectedEndpoint, "guardian-data") {
			guardianData, err := wrapper.GetGuardianData(context.Background(), providedAddress)
			require.True(t, errors.Is(err, expectedErr))
			require.Nil(t, guardianData)
			return
		}

		account, err := wrapper.GetAccount(context.Background(), providedAddress)
		require.True(t, errors.Is(err, expectedErr))
		require.Nil(t, account)
	}
}

func testGetHTTPReturnsStatusCodeError(expectedEndpoint string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		wrapper, _ := NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Equal(t, expectedEndpoint, endpoint)
				return nil, 500, nil
			},
		})
		require.NotNil(t, wrapper)

		if strings.Contains(expectedEndpoint, "guardian-data") {
			guardianData, err := wrapper.GetGuardianData(context.Background(), providedAddress)
			require.True(t, errors.Is(err, authentication.ErrHTTPStatusCodeIsNotOK))
			require.Nil(t, guardianData)
			return
		}

		account, err := wrapper.GetAccount(context.Background(), providedAddress)
		require.True(t, errors.Is(err, authentication.ErrHTTPStatusCodeIsNotOK))
		require.Nil(t, account)
	}
}

func testGetHTTPReturnsNilData(expectedEndpoint string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		wrapper, _ := NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Equal(t, expectedEndpoint, endpoint)
				return nil, 200, nil
			},
		})
		require.NotNil(t, wrapper)

		if strings.Contains(expectedEndpoint, "guardian-data") {
			guardianData, err := wrapper.GetGuardianData(context.Background(), providedAddress)
			require.True(t, errors.Is(err, ErrEmptyData))
			require.Nil(t, guardianData)
			return
		}

		account, err := wrapper.GetAccount(context.Background(), providedAddress)
		require.True(t, errors.Is(err, ErrEmptyData))
		require.Nil(t, account)
	}
}

func testUnmarshalFails(expectedEndpoint string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		wrapper, _ := NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Equal(t, expectedEndpoint, endpoint)
				return []byte("invalid json"), 200, nil
			},
		})
		require.NotNil(t, wrapper)

		if strings.Contains(expectedEndpoint, "guardian-data") {
			guardianData, err := wrapper.GetGuardianData(context.Background(), providedAddress)
			require.Error(t, err)
			require.Nil(t, guardianData)
			return
		}

		account, err := wrapper.GetAccount(context.Background(), providedAddress)
		require.Error(t, err)
		require.Nil(t, account)
	}
}

func testGetHTTPReturnsNilUnmarshalledData(expectedEndpoint string, buffToReturn []byte) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		wrapper, _ := NewHttpClientWrapper(&sdkTestsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				require.Equal(t, expectedEndpoint, endpoint)
				return buffToReturn, 200, nil
			},
		})
		require.NotNil(t, wrapper)

		if strings.Contains(expectedEndpoint, "guardian-data") {
			guardianData, err := wrapper.GetGuardianData(context.Background(), providedAddress)
			require.True(t, errors.Is(err, ErrEmptyData))
			require.Nil(t, guardianData)
			return
		}

		account, err := wrapper.GetAccount(context.Background(), providedAddress)
		require.True(t, errors.Is(err, ErrEmptyData))
		require.Nil(t, account)
	}
}
