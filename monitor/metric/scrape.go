package metric

import (
	"fmt"
	"github.com/go-logr/logr"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"net/http"
)

type TrtisMetrics struct {
	log        logr.Logger
	url        string
	GpuMetrics map[string]*float64
}

const (
	Nv_gpu_utilization        = "nv_gpu_utilization"
	Nv_gpu_memory_total_bytes = "nv_gpu_memory_total_bytes"
	Nv_gpu_memory_used_bytes  = "nv_gpu_memory_used_bytes"
)

func NewTrtisMetrics(host string, port int, log logr.Logger) *TrtisMetrics {
	url := fmt.Sprintf("http://%s:%d/metrics", host, port)

	return &TrtisMetrics{
		log:        log,
		url:        url,
		GpuMetrics: map[string]*float64{Nv_gpu_utilization: nil, Nv_gpu_memory_total_bytes: nil, Nv_gpu_memory_used_bytes: nil},
	}
}

func (t *TrtisMetrics) getMetricsFromServer() (map[string]*dto.MetricFamily, error) {
	request, err := http.NewRequest("GET", t.url, nil)
	if err != nil {
		t.log.Error(err, "Failed to create request")
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.log.Error(err, "Metrics call failed")
		return nil, err
	}

	tp := expfmt.TextParser{}
	metrics, err := tp.TextToMetricFamilies(response.Body)
	if err != nil {
		t.log.Error(err, "Failed to parse metrics")
		return nil, err
	}

	return metrics, nil
}

func (t *TrtisMetrics) updateGpuMetrics(metrics map[string]*dto.MetricFamily) {
	//reset metrics
	for k, _ := range t.GpuMetrics {
		t.GpuMetrics[k] = nil
	}
	for name, val := range metrics {
		t.log.Info(name)
		for _, metric := range val.GetMetric() {
			if metric.Gauge != nil {
				t.log.Info("Value", "name", name, "value", metric.Gauge.Value)
				switch name {
				case Nv_gpu_memory_total_bytes, Nv_gpu_memory_used_bytes, Nv_gpu_utilization:
					t.GpuMetrics[name] = metric.Gauge.Value
				}
			}
		}
	}
}

func (t *TrtisMetrics) ShowMetrics() {
	for k, v := range t.GpuMetrics {
		t.log.Info("GPU Metrics", k, v)
	}
}

func (t *TrtisMetrics) UpdateMetrics() error {
	metrics, err := t.getMetricsFromServer()
	if err != nil {
		return err
	}
	t.updateGpuMetrics(metrics)
	return nil
}
