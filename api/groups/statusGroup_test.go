package groups_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/multiversx/mx-multi-factor-auth-go-service/api/groups"
	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
	mockFacade "github.com/multiversx/mx-multi-factor-auth-go-service/testscommon/facade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type statusMetricsResponse struct {
	Data struct {
		Metrics map[string]*requests.EndpointMetricsResponse `json:"metrics"`
	}
	Error string `json:"error"`
	Code  string `json:"code"`
}

func TestNewStatusGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		gg, err := groups.NewStatusGroup(nil)

		assert.Nil(t, gg)
		assert.True(t, errors.Is(err, core.ErrNilFacadeHandler))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ng, err := groups.NewStatusGroup(&mockFacade.GuardianFacadeStub{})

		assert.NotNil(t, ng)
		assert.Nil(t, err)
	})
}

func TestGetMetrics_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedMetrics := map[string]*requests.EndpointMetricsResponse{
		"/guardian/config": {
			NumRequests:       5,
			NumTotalErrors:    3,
			TotalResponseTime: 100,
		},
	}
	facade := &mockFacade.GuardianFacadeStub{
		GetMetricsCalled: func() map[string]*requests.EndpointMetricsResponse {
			return expectedMetrics
		},
	}

	statusGroup, err := groups.NewStatusGroup(facade)
	require.Nil(t, err)

	ws := startWebServer(statusGroup, "status", getStatusRoutesConfig(), providedAddr)

	req, _ := http.NewRequest("GET", "/status/metrics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	var apiResp statusMetricsResponse
	loadResponse(resp.Body, &apiResp)
	require.Equal(t, http.StatusOK, resp.Code)

	require.Equal(t, expectedMetrics, apiResp.Data.Metrics)
}

func TestGetPrometheusMetrics_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedMetrics := `num_requests{endpoint="/guardian/config"} 37`
	facade := &mockFacade.GuardianFacadeStub{
		GetMetricsForPrometheusCalled: func() string {
			return expectedMetrics
		},
	}

	statusGroup, err := groups.NewStatusGroup(facade)
	require.NoError(t, err)
	ws := startWebServer(statusGroup, "status", getStatusRoutesConfig(), providedAddr)

	req, _ := http.NewRequest("GET", "/status/prometheus-metrics", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, expectedMetrics, string(bodyBytes))
}

func TestStatusGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		sg, _ := groups.NewStatusGroup(&mockFacade.GuardianFacadeStub{})

		err := sg.UpdateFacade(nil)
		assert.Equal(t, core.ErrNilFacadeHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		sg, _ := groups.NewStatusGroup(&mockFacade.GuardianFacadeStub{})

		newFacade := &mockFacade.GuardianFacadeStub{}

		err := sg.UpdateFacade(newFacade)
		assert.Nil(t, err)
	})
}

func TestStatusGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sg, _ := groups.NewStatusGroup(nil)
	assert.True(t, sg.IsInterfaceNil())

	sg, _ = groups.NewStatusGroup(&mockFacade.GuardianFacadeStub{})
	assert.False(t, sg.IsInterfaceNil())
}

func getStatusRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"status": {
				Routes: []config.RouteConfig{
					{Name: "/metrics", Open: true},
					{Name: "/prometheus-metrics", Open: true},
				},
			},
		},
	}
}
