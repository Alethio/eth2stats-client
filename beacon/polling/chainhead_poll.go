package polling

import (
	"github.com/alethio/eth2stats-client/beacon"
	"github.com/sirupsen/logrus"

	"time"

	"github.com/alethio/eth2stats-client/types"
)

const (
	PollingInterval = time.Second
)

var log = logrus.WithField("module", "polling")

type ChainHeadClientPoller struct {
	data   chan types.ChainHead
	client beacon.Client

	stopChan chan bool
}

// Check interface
var _ = beacon.ChainHeadSubscription((*ChainHeadClientPoller)(nil))

func NewChainHeadClientPoller(client beacon.Client) *ChainHeadClientPoller {
	return &ChainHeadClientPoller{
		data:   make(chan types.ChainHead),
		client: client,
	}
}

func (s *ChainHeadClientPoller) Start() {
	log.Info("polling for new heads")
	var lastHead *types.ChainHead

	for {
		select {
		case <-s.stopChan:
			close(s.data)
			return
		default:
			head, err := s.client.GetChainHead()
			if err != nil {
				log.Errorf("failed to poll for chain head")
				time.Sleep(PollingInterval)
				continue
			}
			if lastHead == nil || *lastHead != *head {
				s.data <- types.ChainHead{
					HeadSlot:           head.HeadSlot,
					HeadBlockRoot:      head.HeadBlockRoot,
					FinalizedSlot:      head.FinalizedSlot,
					FinalizedBlockRoot: head.FinalizedBlockRoot,
					JustifiedSlot:      head.JustifiedSlot,
					JustifiedBlockRoot: head.JustifiedBlockRoot,
				}
				lastHead = head
			}

			time.Sleep(PollingInterval)
		}
	}
}

func (s *ChainHeadClientPoller) Channel() <-chan types.ChainHead {
	return s.data
}

func (s *ChainHeadClientPoller) Close() {
	s.stopChan <- true
}
