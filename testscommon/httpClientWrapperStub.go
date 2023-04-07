package testscommon

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-sdk-go/data"
)

// HttpClientWrapperStub -
type HttpClientWrapperStub struct {
	GetAccountCalled      func(ctx context.Context, address string) (*data.Account, error)
	GetGuardianDataCalled func(ctx context.Context, address string) (*api.GuardianData, error)
}

// GetAccount -
func (stub *HttpClientWrapperStub) GetAccount(ctx context.Context, address string) (*data.Account, error) {
	if stub.GetAccountCalled != nil {
		return stub.GetAccountCalled(ctx, address)
	}
	return &data.Account{}, nil
}

// GetGuardianData -
func (stub *HttpClientWrapperStub) GetGuardianData(ctx context.Context, address string) (*api.GuardianData, error) {
	if stub.GetGuardianDataCalled != nil {
		return stub.GetGuardianDataCalled(ctx, address)
	}
	return &api.GuardianData{}, nil
}

// IsInterfaceNil -
func (stub *HttpClientWrapperStub) IsInterfaceNil() bool {
	return stub == nil
}
