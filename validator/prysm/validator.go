package prysm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	proto "github.com/alethio/eth2stats-proto"
	"github.com/davecgh/go-spew/spew"
	"github.com/parnurzeal/gorequest"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("module", "validator")

// TODO reorg in a single package maybe
type Validator struct {
	metricsURL string
	service    proto.ValidatorClient
}

func (v *Validator) Run() {
	ctx := context.Background()
	for {
		log.Trace("collecting validator data")
		metrics, err := v.queryMetrics()
		if err != nil {
			// TODO report to validator service as unconnectable
			log.Errorf("could not query validator node: %s", err)
			req := &proto.ValidatorClientRequest{
				Status: proto.ValidatorClientRequest_UNREACHABLE,
			}
			_, err := v.service.ValidatorClient(ctx, req)
			if err != nil {
				log.Errorf("setting unreachable validator status: %s", err)
			}
			time.Sleep(PollingInterval)
			continue
		}

		req := &proto.ValidatorClientRequest{
			Gauges: make(map[string]float64),
		}

		v.extractMetrics(metrics, req)
		_, err = v.service.ValidatorClient(ctx, req)
		if err != nil {
			log.Errorf("setting validator client: %s", err)
		}
		spew.Dump(req)

		log.Trace("done sending validator data")

		time.Sleep(PollingInterval)
	}
}

func (v *Validator) queryMetrics() (map[string]*io_prometheus_client.MetricFamily, error) {
	request := gorequest.New()
	resp, _, errs := request.Get(v.metricsURL).End()
	if len(errs) > 0 {
		log.Error(errs)
		return nil, fmt.Errorf("%+q", errs)
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("validator metrics query responded with status code != 200: %d", resp.StatusCode)
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

func (v *Validator) extractMetrics(metrics map[string]*io_prometheus_client.MetricFamily, req *proto.ValidatorClientRequest) {
	spew.Dump(metrics["log_entries_total"].Metric[0].Counter)
	me := MetricsExtractor(metrics)
	families := []FamilyToKey{
		{"process_start_time_seconds", "start_time", nil},
		{"process_resident_memory_bytes", "mem_usage", nil},
		{"process_cpu_seconds_total", "cpu_secs", nil},
		{"log_entries_total", "validator-errors", []LabelPair{
			{"prefix", "validator"},
			{"level", "error"},
		}},
		{"log_entries_total", "validator-warnings", []LabelPair{
			{"prefix", "validator"},
			{"level", "warning"},
		}},
	}
	for _, g := range families {
		me.SetNotNil(g, &req.Gauges)
	}
}

func NewValidator(url string, service proto.ValidatorClient) *Validator {
	return &Validator{
		metricsURL: url,
		service:    service,
	}
}
