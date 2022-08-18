package facade

import "github.com/ElrondNetwork/multi-factor-auth-go-service/providers"

// FacadeStub -
type FacadeStub struct {
	GetRarityCalled           func(record providers.DataRecord) (float64, error)
	GetCollectionRarityCalled func(records []*providers.DataRecord) ([]float64, error)
	RestApiInterfaceCalled    func() string
	PprofEnabledCalled        func() bool
}

// GetRarity -
func (stub *FacadeStub) GetRarity(record providers.DataRecord) (float64, error) {
	if stub.GetRarityCalled != nil {
		return stub.GetRarityCalled(record)
	}

	return 0, nil
}

// GetCollectionRarity -
func (stub *FacadeStub) GetCollectionRarity(records []*providers.DataRecord) ([]float64, error) {
	if stub.GetCollectionRarityCalled != nil {
		return stub.GetCollectionRarityCalled(records)
	}

	return make([]float64, 0), nil
}

// RestApiInterface -
func (stub *FacadeStub) RestApiInterface() string {
	if stub.RestApiInterfaceCalled != nil {
		return stub.RestApiInterfaceCalled()
	}
	return "localhost:8080"
}

// PprofEnabled -
func (stub *FacadeStub) PprofEnabled() bool {
	if stub.PprofEnabledCalled != nil {
		return stub.PprofEnabledCalled()
	}
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *FacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
