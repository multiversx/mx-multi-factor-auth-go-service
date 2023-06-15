package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/multi-factor-auth-go-service/api/shared"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	chainApiShared "github.com/multiversx/mx-chain-go/api/shared"
)

const (
	metricsPath           = "/metrics"
	prometheusMetricsPath = "/prometheus-metrics"
)

type statusGroup struct {
	*baseGroup
	facade    shared.FacadeHandler
	mutFacade sync.RWMutex
}

// NewStatusGroup returns a new instance of status group
func NewStatusGroup(facade shared.FacadeHandler) (*statusGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for status group", core.ErrNilFacadeHandler)
	}

	sg := &statusGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*chainApiShared.EndpointHandlerData{
		{
			Path:    metricsPath,
			Handler: sg.getMetrics,
			Method:  http.MethodGet,
		},
		{
			Path:    prometheusMetricsPath,
			Handler: sg.getPrometheusMetrics,
			Method:  http.MethodGet,
		},
	}
	sg.endpoints = endpoints

	return sg, nil
}

// getMetrics will expose endpoints statistics in json format
func (sg *statusGroup) getMetrics(c *gin.Context) {
	metricsResults := sg.facade.GetMetrics()

	returnStatus(c, gin.H{"metrics": metricsResults}, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

// getPrometheusMetrics will expose proxy metrics in prometheus format
func (sg *statusGroup) getPrometheusMetrics(c *gin.Context) {
	metricsResults := sg.facade.GetMetricsForPrometheus()

	c.String(http.StatusOK, metricsResults)
}

// UpdateFacade will update the facade
func (sg *statusGroup) UpdateFacade(newFacade shared.FacadeHandler) error {
	if check.IfNil(newFacade) {
		return core.ErrNilFacadeHandler
	}

	sg.mutFacade.Lock()
	sg.facade = newFacade
	sg.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sg *statusGroup) IsInterfaceNil() bool {
	return sg == nil
}
