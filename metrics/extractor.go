package metrics

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/parnurzeal/gorequest"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("module", "metrics")

type Extractor map[string]*io_prometheus_client.MetricFamily

type LabelPair struct {
	Name  string
	Value string
}

func (me Extractor) GetInt64(family string) *int64 {
	v := me.First(family, nil)
	if v != nil {
		i := int64(*v)
		return &i
	}
	return nil
}

func (me Extractor) First(family string, labels []LabelPair) *float64 {
	metricFamily, ok := me[family]
	if !ok {
		log.Tracef("could not find metric family %s", family)
		return nil
	}

	metric := metricFamily.GetMetric()

	var value *float64
	for _, m := range metric {
		var v *float64
		if *metricFamily.Type == io_prometheus_client.MetricType_COUNTER {
			c := m.GetCounter()
			if c != nil {
				v = c.Value
			}
		} else if *metricFamily.Type == io_prometheus_client.MetricType_GAUGE {
			g := m.GetGauge()
			if g != nil {
				v = g.Value
			}
		}

		// check labels are a match
		// implicit true if no labels are given
		foundAll := true
		for _, l := range labels {
			found := false
			for _, ml := range m.Label {
				if l.Name == *ml.Name && l.Value == *ml.Value {
					found = true
					break
				}
			}
			// if a label is not found, consider it failed
			if !found {
				foundAll = false
				break
			}
		}

		// not all labels found
		if !foundAll {
			continue
		}

		value = v
		break
	}

	return value
}

func NewFromURL(url string) (*Extractor, error) {
	request := gorequest.New()
	resp, _, errs := request.Get(url).End()
	if len(errs) > 0 {
		log.Error(errs)
		return nil, fmt.Errorf("%+q", errs)
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("validator metrics query responded with status code != 200: %d", resp.StatusCode)
		log.Error(err)
		return nil, err
	}

	metrics, err := loadMetrics(resp.Body)
	if err != nil {
		return nil, err
	}

	return metrics, resp.Body.Close()
}

func NewFromFile(path string) (*Extractor, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	metrics, err := loadMetrics(file)
	if err != nil {
		return nil, err
	}

	return metrics, file.Close()
}

func loadMetrics(reader io.Reader) (*Extractor, error) {
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		log.Error("reading text format failed:", err)
		return nil, err
	}

	me := Extractor(metricFamilies)
	return &me, nil
}
