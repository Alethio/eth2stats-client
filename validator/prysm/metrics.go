package prysm

import (
	"github.com/davecgh/go-spew/spew"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

type MetricsExtractor map[string]*io_prometheus_client.MetricFamily

type FamilyToKey struct {
	Family string
	Key    string
	Labels []LabelPair
}

type LabelPair struct {
	Name  string
	Value string
}

func (me MetricsExtractor) SetNotNil(f2k FamilyToKey, gauges *map[string]float64) {
	var v *float64
	if len(f2k.Labels) == 0 {
		v = me.extractFloat(f2k.Family)
	} else {
		v = me.extractFloatWithLabels(f2k.Family, f2k.Labels)
	}

	if v != nil {
		(*gauges)[f2k.Key] = *v
	}
}

func (me MetricsExtractor) GetGaugeInt64(family string) *int64 {
	v := me.extractFloat(family)
	if v != nil {
		i := int64(*v)
		return &i
	}
	return nil
}

func (me MetricsExtractor) GetFloat(family string) *float64 {
	return me.extractFloat(family)
}

func (me MetricsExtractor) extractFloatWithLabels(family string, labels []LabelPair) *float64 {
	metricFamily, ok := me[family]
	if !ok {
		log.Trace("could not find metric family %s", family)
		return nil
	}

	metrics := metricFamily.GetMetric()
	if len(metrics) == 0 {
		log.Trace("could not find any metric for metric family %s", family)
		return nil
	}

	spew.Dump("----", metrics[0])
	// for _, l := range labels {
	// 	//found := false
	// 	for _, ml := range metrics[0].Label {
	// 		spew.Dump(l, ml)
	// 	}
	// }
	return nil
}

func (me MetricsExtractor) extractFloat(family string) *float64 {
	metricFamily, ok := me[family]
	if !ok {
		log.Trace("could not find metric family %s", family)
		return nil
	}

	metric := metricFamily.GetMetric()
	if len(metric) == 0 {
		log.Trace("could not find any metric for metric family %s", family)
		return nil
	}

	mm := metric[0]
	gauge := mm.GetGauge()
	if gauge != nil {
		return gauge.Value
	}
	counter := mm.GetCounter()
	if counter != nil {
		return counter.Value
	}

	log.Tracef("%s metric is not of type gauge or counter", family)
	return nil
}
