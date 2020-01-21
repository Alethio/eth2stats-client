package lighthouse

import (
	"time"

	"github.com/alethio/eth2stats-client/types"
)

type ChainHeadSubscription struct {
	data   chan types.ChainHead
	client *LighthouseHTTPClient

	stopChan chan bool
}

func NewChainHeadSubscription(client *LighthouseHTTPClient) *ChainHeadSubscription {
	return &ChainHeadSubscription{
		data:   make(chan types.ChainHead),
		client: client,
	}
}

func (s *ChainHeadSubscription) Start() {
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
				close(s.data)
				return
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

			time.Sleep(time.Second)
		}
	}
}

func (s *ChainHeadSubscription) Channel() <-chan types.ChainHead {
	return s.data
}

func (s *ChainHeadSubscription) Close() {
	s.stopChan <- true
}
