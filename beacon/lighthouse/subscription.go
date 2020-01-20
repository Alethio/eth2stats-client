package lighthouse

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/common/log"

	"github.com/alethio/eth2stats-client/types"
)

type ChainHeadSubscription struct {
	data   chan types.ChainHead
	client *LighthouseClient

	stopChan chan bool
}

func NewChainHeadSubscription(client *LighthouseClient) *ChainHeadSubscription {
	return &ChainHeadSubscription{
		data:   make(chan types.ChainHead),
		client: client,
	}
}

func (s *ChainHeadSubscription) Start() {
	log.Info("polling for new heads")

	for {
		select {
		case <-s.stopChan:
			close(s.data)
			return
		default:
			log.Info("head")
			spew.Dump(s.client.GetChainHead())
			// s.data <- types.ChainHead{
			// 	HeadSlot:           data.HeadSlot,
			// 	HeadBlockRoot:      hex.EncodeToString(data.HeadBlockRoot),
			// 	FinalizedSlot:      data.FinalizedSlot,
			// 	FinalizedBlockRoot: hex.EncodeToString(data.FinalizedBlockRoot),
			// 	JustifiedSlot:      data.JustifiedSlot,
			// 	JustifiedBlockRoot: hex.EncodeToString(data.JustifiedBlockRoot),
			// }

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
