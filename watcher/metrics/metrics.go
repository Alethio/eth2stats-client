package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/avast/retry-go"
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
	client *http.Client // re-use for metrics requests.
}

func New(config Config) *Watcher {
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: PollDialTimeout,
		}).DialContext,
		// make keep-alives longer than the interval to make client re-use effective.
		IdleConnTimeout:     2 * PollingInterval,
		TLSHandshakeTimeout: PollTLSTimeout,
	}
	var httpClient = &http.Client{
		Timeout:   PollTimeout,
		Transport: netTransport,
	}
	return &Watcher{
		config: config,
		client: httpClient,
	}
}

func (w *Watcher) Run(ctx context.Context) {
	log.Info("Started polling metrics")
	w.poll()
	ticker := time.NewTicker(PollingInterval)
	for {
		select {
		case <-ticker.C:
			w.poll()
			break
		case <-ctx.Done():
			ticker.Stop()
			log.Info("Stopped polling metrics")
			return
		}
	}
}

func (w *Watcher) poll() {
	_ = retry.Do(
		func() error {
			log.Info("querying metrics")
			metrics, err := w.query()
			if err != nil {
				log.Warnf("failed to poll metrics: %s", err)
				return err
			}
			w.monitorMetrics(metrics)
			return nil
		},
		retry.Attempts(PollRetryAttempts),
	)
}

func (w *Watcher) query() (map[string]*io_prometheus_client.MetricFamily, error) {
	// Don't keep a request open for longer than the interval time.
	req, err := http.NewRequest("GET", w.config.MetricsURL, nil)
	if err != nil {
		return nil, err
	}
	// disable caching for up-to-date metrics (if running behind a proxy or something else)s
	req.Header.Set("Cache-control", "no-cache")
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	log.Trace("done querying metrics")

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

	var value float64

	if gauge := memMetric.GetGauge(); gauge != nil && gauge.Value != nil {
		value = gauge.GetValue()
	} else if untyped := memMetric.GetUntyped(); untyped != nil {
		value = untyped.GetValue()
	} else {
		log.Info("could not extract mem value")
		return
	}

	memInt := int64(value)

	log.Tracef("mem usage: %d", memInt)
	w.mu.Lock()
	w.data.MemUsage = &memInt
	w.mu.Unlock()
	log.Tracef("mem usage written: %d", memInt)
}
