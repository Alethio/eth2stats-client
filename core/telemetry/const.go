package telemetry

import (
	"time"
)

const (
	PollingInterval      = 12 * time.Second * 4 // TODO increase poll interval until we get actual counters
	MemoryUsageThreshold = 10 * 1024 * 1024
)
