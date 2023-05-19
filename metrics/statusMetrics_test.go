package metrics_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/metrics"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusMetrics(t *testing.T) {
	t.Parallel()

	sm := metrics.NewStatusMetrics()
	require.False(t, check.IfNil(sm))
}

func TestStatusMetrics_AddRequestData(t *testing.T) {
	t.Parallel()

	t.Run("one metric exists for an endpoint", func(t *testing.T) {
		t.Parallel()

		sm := metrics.NewStatusMetrics()

		testEndpoint, testDuration := "/guardian/config", 1*time.Second
		sm.AddRequestData(testEndpoint, testDuration, metrics.NonErrorCode)

		res := sm.GetAll()
		require.Equal(t, res[testEndpoint], &requests.EndpointMetricsResponse{
			NumRequests:       1,
			NumTotalErrors:    0,
			ErrorsCount:       make(map[int]uint64),
			TotalResponseTime: testDuration,
		})
	})

	t.Run("multiple entries exist for an endpoint", func(t *testing.T) {
		t.Parallel()

		sm := metrics.NewStatusMetrics()

		testEndpoint := "/guardian/config"
		testDuration0, testDuration1, testDuration2 := 4*time.Millisecond, 20*time.Millisecond, 2*time.Millisecond
		sm.AddRequestData(testEndpoint, testDuration0, metrics.NonErrorCode)
		sm.AddRequestData(testEndpoint, testDuration1, 400)
		sm.AddRequestData(testEndpoint, testDuration2, metrics.NonErrorCode)

		res := sm.GetAll()
		require.Equal(t, res[testEndpoint], &requests.EndpointMetricsResponse{
			NumRequests:    3,
			NumTotalErrors: 1,
			ErrorsCount: map[int]uint64{
				400: 1,
			},
			TotalResponseTime: testDuration0 + testDuration1 + testDuration2,
		})
	})

	t.Run("multiple entries for multiple endpoints", func(t *testing.T) {
		t.Parallel()

		sm := metrics.NewStatusMetrics()

		testEndpoint0, testEndpoint1 := "/guardian/config", "/guardian/config2"

		testDuration0End0, testDuration1End0 := time.Second, 5*time.Second
		testDuration0End1, testDuration1End1 := time.Hour, 4*time.Hour

		sm.AddRequestData(testEndpoint0, testDuration0End0, 400)
		sm.AddRequestData(testEndpoint0, testDuration1End0, metrics.NonErrorCode)

		sm.AddRequestData(testEndpoint1, testDuration0End1, 400)
		sm.AddRequestData(testEndpoint1, testDuration1End1, 400)

		res := sm.GetAll()

		require.Len(t, res, 2)
		require.Equal(t, res[testEndpoint0], &requests.EndpointMetricsResponse{
			NumRequests:    2,
			NumTotalErrors: 1,
			ErrorsCount: map[int]uint64{
				400: 1,
			},
			TotalResponseTime: testDuration0End0 + testDuration1End0,
		})
		require.Equal(t, res[testEndpoint1], &requests.EndpointMetricsResponse{
			NumRequests:    2,
			NumTotalErrors: 2,
			ErrorsCount: map[int]uint64{
				400: 2,
			},
			TotalResponseTime: testDuration0End1 + testDuration1End1,
		})
	})
}

func TestStatusMetrics_GetMetricsForPrometheus(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		sm := metrics.NewStatusMetrics()

		testEndpoint := "/guardian/config"
		testDuration0, testDuration1, testDuration2 := 4*time.Millisecond, 20*time.Millisecond, 2*time.Millisecond
		sm.AddRequestData(testEndpoint, testDuration0, metrics.NonErrorCode)
		sm.AddRequestData(testEndpoint, testDuration1, 400)
		sm.AddRequestData(testEndpoint, testDuration2, metrics.NonErrorCode)

		res := sm.GetMetricsForPrometheus()

		expectedString := `# TYPE num_requests counter
num_requests{operation="/guardian/config"} 3

# TYPE num_total_errors counter
num_total_errors{operation="/guardian/config"} 1

# TYPE total_response_time_ns counter
total_response_time_ns{operation="/guardian/config"} 2.6e+07

# TYPE requests_errors gauge
requests_errors{operation="/guardian/config",errorCode="400"} 1

`

		require.Equal(t, expectedString, res)

	})
}

func TestStatusMetrics_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.Nil(t, r)
	}()

	sm := metrics.NewStatusMetrics()

	numIterations := 500
	wg := sync.WaitGroup{}
	wg.Add(numIterations)

	for i := 0; i < numIterations; i++ {
		go func(index int) {
			switch index % 3 {
			case 0:
				sm.AddRequestData(fmt.Sprintf("endpoint_%d", index%5), time.Hour*time.Duration(index), metrics.NonErrorCode)
			case 1:
				res := sm.GetAll()
				delete(res, "endpoint_0")
			case 2:
				_ = sm.GetMetricsForPrometheus()
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestStatusMetrics_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sm := metrics.NewStatusMetrics()
	assert.False(t, sm.IsInterfaceNil())
}
