package metrics

import (
	"bytes"
	"strconv"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"google.golang.org/protobuf/proto"
)

func requestsCounterMetric(metricName, endpoint string, count uint64) string {
	metricFamily := &dto.MetricFamily{
		Name: proto.String(metricName),
		Type: dto.MetricType_COUNTER.Enum(),
		Metric: []*dto.Metric{
			{
				Label: []*dto.LabelPair{
					{
						Name:  proto.String("operation"),
						Value: proto.String(endpoint),
					},
				},
				Counter: &dto.Counter{
					Value: proto.Float64(float64(count)),
				},
			},
		},
	}

	return promMetricAsString(metricFamily)
}

func requestsErrorsMetrics(metricName, endpoint string, errorsCount map[int]uint64) string {
	metrics := make([]*dto.Metric, 0, len(errorsCount))

	for code, count := range errorsCount {
		m := &dto.Metric{
			Label: []*dto.LabelPair{
				{
					Name:  proto.String("operation"),
					Value: proto.String(endpoint),
				},
				{
					Name:  proto.String("errorCode"),
					Value: proto.String(strconv.Itoa(code)),
				},
			},
			Gauge: &dto.Gauge{
				Value: proto.Float64(float64(count)),
			},
		}
		metrics = append(metrics, m)
	}

	metricFamily := &dto.MetricFamily{
		Name:   proto.String(metricName),
		Type:   dto.MetricType_GAUGE.Enum(),
		Metric: metrics,
	}

	return promMetricAsString(metricFamily)
}

func promMetricAsString(metric *dto.MetricFamily) string {
	out := bytes.NewBuffer(make([]byte, 0))
	_, err := expfmt.MetricFamilyToText(out, metric)
	if err != nil {
		return ""
	}

	return out.String()
}
