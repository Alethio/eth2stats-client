package telemetry

import (
	"time"
)

const (
	PollingInterval      = 12 * time.Second
	MemoryUsageThreshold = 10 * 1024 * 1024
)
