package beacon

import (
	"github.com/alethio/eth2stats-client/types"
)

type ChainHeadSubscription interface {
	Channel() <-chan types.ChainHead
	Close()
}

type Client interface {
	GetVersion() (string, error)
	GetGenesisTime() (int64, error)
	GetPeerCount() (int64, error)
	GetAttestationsInPoolCount() (int64, error)
	GetSyncStatus() (bool, error)
	GetChainHead() (*types.ChainHead, error)

	SubscribeChainHeads() (ChainHeadSubscription, error)
}
