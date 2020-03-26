package prysm

import (
	"github.com/alethio/eth2stats-client/metrics"
)

type stat struct {
	key    string
	family string
	labels []metrics.LabelPair
}
