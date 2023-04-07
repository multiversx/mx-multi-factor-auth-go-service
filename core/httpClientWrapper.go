package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-sdk-go/authentication"
	"github.com/multiversx/mx-sdk-go/data"
)

type httpClientWrapper struct {
	httpClient HttpClient
}

// NewHttpClientWrapper returns a new instance of httpClientWrapper
func NewHttpClientWrapper(httpClient HttpClient) (*httpClientWrapper, error) {
	if check.IfNil(httpClient) {
		return nil, ErrNilHttpClient
	}

	return &httpClientWrapper{
		httpClient: httpClient,
	}, nil
}

// GetAccount makes a http request and returns the account for the provided address
func (hcw *httpClientWrapper) GetAccount(ctx context.Context, address string) (*data.Account, error) {
	endpoint := fmt.Sprintf(getAccountEndpointFormat, address)
	buff, err := hcw.getData(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var accountResp data.AccountResponse
	err = json.Unmarshal(buff, &accountResp)
	if err != nil {
		return nil, err
	}
	if accountResp.Data.Account == nil {
		return nil, fmt.Errorf("%w while getting account %s", ErrEmptyData, address)
	}

	return accountResp.Data.Account, nil
}

// GetGuardianData makes a http request and returns guardian data for the provided address
func (hcw *httpClientWrapper) GetGuardianData(ctx context.Context, address string) (*api.GuardianData, error) {
	endpoint := fmt.Sprintf(getGuardianDataEndpointFormat, address)
	buff, err := hcw.getData(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var guardianDataResp data.GuardianDataResponse
	err = json.Unmarshal(buff, &guardianDataResp)
	if err != nil {
		return nil, err
	}
	if guardianDataResp.Data.GuardianData == nil {
		return nil, fmt.Errorf("%w while getting guardian data for user %s", ErrEmptyData, address)
	}

	return guardianDataResp.Data.GuardianData, nil
}

func (hcw *httpClientWrapper) getData(ctx context.Context, endpoint string) ([]byte, error) {
	buff, code, err := hcw.httpClient.GetHTTP(ctx, endpoint)
	if err != nil || code != http.StatusOK {
		return nil, authentication.CreateHTTPStatusError(code, err)
	}
	if len(buff) == 0 {
		return nil, fmt.Errorf("%w while calling %s, code %d", ErrEmptyData, endpoint, code)
	}

	return buff, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hcw *httpClientWrapper) IsInterfaceNil() bool {
	return hcw == nil
}
