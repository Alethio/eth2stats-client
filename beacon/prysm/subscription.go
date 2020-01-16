package prysm

import (
	"encoding/hex"

	prysmAPI "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"

	"github.com/alethio/eth2stats-client/types"
)

type ChainHeadSubscription struct {
	data chan types.ChainHead

	stopChan chan bool
}

func NewChainHeadSubscription() *ChainHeadSubscription {
	return &ChainHeadSubscription{
		data: make(chan types.ChainHead),
	}
}

func (s *ChainHeadSubscription) FeedFromStream(stream prysmAPI.BeaconChain_StreamChainHeadClient) {
	log.Info("listening on stream")

	for {
		select {
		case <-s.stopChan:
			close(s.data)
			return
		default:
			data, err := stream.Recv()
			if err != nil {
				close(s.data)
				return
			}

			log.WithField("headSlot", data.GetHeadSlot()).Info("got chain head")

			s.data <- types.ChainHead{
				HeadSlot:           data.HeadSlot,
				HeadBlockRoot:      hex.EncodeToString(data.HeadBlockRoot),
				FinalizedSlot:      data.FinalizedSlot,
				FinalizedBlockRoot: hex.EncodeToString(data.FinalizedBlockRoot),
				JustifiedSlot:      data.JustifiedSlot,
				JustifiedBlockRoot: hex.EncodeToString(data.JustifiedBlockRoot),
			}
		}
	}
}

func (s *ChainHeadSubscription) Channel() <-chan types.ChainHead {
	return s.data
}

func (s *ChainHeadSubscription) Close() {
	s.stopChan <- true
}
