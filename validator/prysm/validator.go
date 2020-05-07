package prysm

import (
	"context"
	"time"

	proto "github.com/alethio/eth2stats-proto"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/alethio/eth2stats-client/metrics"
)

var log = logrus.WithField("module", "validator")

// TODO reorg in a single package maybe
type Validator struct {
	contextWithToken func() context.Context
	metricsURL       string
	service          proto.ValidatorClient
}

func (v *Validator) Run() {
	ctx := v.contextWithToken()
	for {
		log.Trace("collecting validator data")

		req := &proto.ValidatorClientRequest{
			Status: proto.ValidatorClientRequest_ONLINE,
			Data:   make(map[string]float64),
		}

		err := v.extractMetrics(req)
		if err != nil {
			log.Errorf("could not query validator node: %s", err)
			req = &proto.ValidatorClientRequest{
				Status: proto.ValidatorClientRequest_UNREACHABLE,
			}
		}

		_, err = v.service.ValidatorClient(ctx, req)
		if err != nil {
			log.Errorf("setting validator client: %s", err)
		}
		spew.Dump(req)

		log.Trace("done sending validator data")

		time.Sleep(PollingInterval)
	}
}

func (v *Validator) extractMetrics(req *proto.ValidatorClientRequest) error {
	me, err := metrics.NewFromURL(v.metricsURL)
	if err != nil {
		return errors.Wrap(err, "failed to get metrics")
	}

	stats := []stat{
		{"startTime", "process_start_time_seconds", nil},
		{"memUsage", "process_resident_memory_bytes", nil},
		{"cpuSecs", "process_cpu_seconds_total", nil},
		{
			"validatorErrors",
			"log_entries_total",
			[]metrics.LabelPair{
				{"prefix", "validator"},
				{"level", "error"},
			},
		},
		{"validatorWarnings",
			"log_entries_total",
			[]metrics.LabelPair{
				{"prefix", "validator"},
				{"level", "warning"},
			},
		},
	}
	for _, s := range stats {
		v := me.First(s.family, s.labels)
		if v != nil {
			req.Data[s.key] = *v
		}

	}
	return nil
}

func NewValidator(url string, service proto.ValidatorClient, contextWithToken func() context.Context) *Validator {
	return &Validator{
		contextWithToken: contextWithToken,
		metricsURL:       url,
		service:          service,
	}
}
