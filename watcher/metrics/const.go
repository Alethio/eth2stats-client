package metrics

import (
	"time"
)

const PollingInterval = 30 * time.Second

const PollRetryAttempts = 4

const PollTimeout = 5 * time.Second

const PollDialTimeout = 10 * time.Second
const PollTLSTimeout = 10 * time.Second
