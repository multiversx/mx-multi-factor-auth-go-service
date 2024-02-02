package metrics

import (
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

// NonErrorCode defines the non error value
const NonErrorCode = 0

const (
	numRequestsPromMetric       = "num_requests"
	numTotalErrorsPromMetric    = "num_total_errors"
	totalResponseTimePromMetric = "total_response_time"
	requestsErrorsPromMetric    = "requests_errors"
)

type statusMetrics struct {
	endpointMetrics     map[string]*requests.EndpointMetricsResponse
	mutEndpointsMetrics sync.RWMutex
}

// NewStatusMetrics will return an instance of the statusMetrics
func NewStatusMetrics() *statusMetrics {
	return &statusMetrics{
		endpointMetrics: make(map[string]*requests.EndpointMetricsResponse),
	}
}

// AddRequestData will add the received data to the metrics map
func (sm *statusMetrics) AddRequestData(path string, duration time.Duration, status int) {
	sm.mutEndpointsMetrics.Lock()
	defer sm.mutEndpointsMetrics.Unlock()

	errorIncrementalStep := uint64(0)
	if status != NonErrorCode {
		errorIncrementalStep = 1
	}

	currentData := sm.endpointMetrics[path]
	if currentData == nil {
		newMetricsData := &requests.EndpointMetricsResponse{
			NumRequests:       1,
			NumTotalErrors:    errorIncrementalStep,
			ErrorsCount:       make(map[int]uint64),
			TotalResponseTime: duration,
		}

		if errorIncrementalStep == 1 {
			newMetricsData.ErrorsCount[status]++
		}

		sm.endpointMetrics[path] = newMetricsData

		return
	}

	currentData.NumRequests++
	if errorIncrementalStep == 1 {
		currentData.NumTotalErrors++
		currentData.ErrorsCount[status]++
	}
	currentData.TotalResponseTime += duration
}

// GetAll returns the metrics map
func (sm *statusMetrics) GetAll() map[string]*requests.EndpointMetricsResponse {
	sm.mutEndpointsMetrics.RLock()
	defer sm.mutEndpointsMetrics.RUnlock()

	return sm.getAll()
}

func (sm *statusMetrics) getAll() map[string]*requests.EndpointMetricsResponse {
	newMap := make(map[string]*requests.EndpointMetricsResponse)
	for key, value := range sm.endpointMetrics {
		newMap[key] = value
	}

	return newMap
}

// GetMetricsForPrometheus returns the metrics in a prometheus format
func (sm *statusMetrics) GetMetricsForPrometheus() string {
	sm.mutEndpointsMetrics.RLock()
	defer sm.mutEndpointsMetrics.RUnlock()

	metricsMap := sm.getAll()

	stringBuilder := strings.Builder{}

	for endpointPath, endpointData := range metricsMap {
		stringBuilder.WriteString(requestsCounterMetric(numRequestsPromMetric, endpointPath, endpointData.NumRequests))
		stringBuilder.WriteString(requestsCounterMetric(numTotalErrorsPromMetric, endpointPath, endpointData.NumTotalErrors))
		stringBuilder.WriteString(requestsCounterMetric(totalResponseTimePromMetric, endpointPath, uint64(endpointData.TotalResponseTime.Milliseconds())))
		stringBuilder.WriteString(requestsErrorsMetrics(requestsErrorsPromMetric, endpointPath, metricsMap[endpointPath].ErrorsCount))
	}

	return stringBuilder.String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *statusMetrics) IsInterfaceNil() bool {
	return sm == nil
}
