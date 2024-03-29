package testscommon

import (
	"time"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

// StatusMetricsStub -
type StatusMetricsStub struct {
	AddRequestDataCalled          func(path string, duration time.Duration, status int)
	GetAllCalled                  func() map[string]*requests.EndpointMetricsResponse
	GetMetricsForPrometheusCalled func() string
}

// AddRequestData -
func (s *StatusMetricsStub) AddRequestData(path string, duration time.Duration, status int) {
	if s.AddRequestDataCalled != nil {
		s.AddRequestDataCalled(path, duration, status)
	}
}

// GetAll -
func (s *StatusMetricsStub) GetAll() map[string]*requests.EndpointMetricsResponse {
	if s.GetAllCalled != nil {
		return s.GetAllCalled()
	}

	return nil
}

// GetMetricsForPrometheus -
func (s *StatusMetricsStub) GetMetricsForPrometheus() string {
	if s.GetMetricsForPrometheusCalled != nil {
		return s.GetMetricsForPrometheusCalled()
	}

	return ""
}

// IsInterfaceNil -
func (s *StatusMetricsStub) IsInterfaceNil() bool {
	return s == nil
}
