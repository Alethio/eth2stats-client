package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("module", "metrics-watcher")

type Config struct {
	MetricsURL   string
	PollInterval time.Duration
}

type Watcher struct {
	config Config

	mu   sync.Mutex
	data struct {
		MemUsage *int64
	}
}

func New(config Config) *Watcher {
	return &Watcher{
		config: config,
	}
}

func (w *Watcher) Run() {
	w.poll()
	for {
		select {
		case <-time.Tick(PollingInterval):
			w.poll()
		}
	}
}

func (w *Watcher) poll() {
	metrics, err := w.query()
	if err != nil {
		log.Errorf("failed to poll metrics: %s", err)
		return
	}

	w.monitorMetrics(metrics)
}

func (w *Watcher) query() (map[string]*io_prometheus_client.MetricFamily, error) {
	request := gorequest.New()
	resp, _, errs := request.Get(w.config.MetricsURL).End()
	if len(errs) > 0 {
		log.Error(errs)
		return nil, errs[0]
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("metrics query responded with status code != 200")
		log.Error(err)
		return nil, err
	}

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		log.Error("reading text format failed:", err)
		return nil, err
	}

	return metricFamilies, nil
}

func (w *Watcher) monitorMetrics(metrics map[string]*io_prometheus_client.MetricFamily) {
	w.extractMemUsage(metrics)
}

func (w *Watcher) extractMemUsage(metrics map[string]*io_prometheus_client.MetricFamily) {
	metricFamily, ok := metrics["process_resident_memory_bytes"]
	if !ok {
		log.Warn("could not find `process_resident_memory_bytes` in metrics")
		return
	}

	metric := metricFamily.GetMetric()
	if len(metric) == 0 {
		log.Warn("could not find any metric for metric family `process_resident_memory_bytes`")
		return
	}

	memMetric := metric[0]
	gauge := memMetric.GetGauge()
	if gauge == nil {
		log.Warn("memory metric is not of type gauge")
		return
	}

	if gauge.Value == nil {
		log.Warn("memory gauge has nil value")
		return
	}

	memInt := int64(gauge.GetValue())

	w.mu.Lock()
	w.data.MemUsage = &memInt
	w.mu.Unlock()
}
