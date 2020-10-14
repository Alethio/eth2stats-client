package v1

import (
	"fmt"
	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/beacon/polling"
	"github.com/alethio/eth2stats-client/types"
	"github.com/dghubble/sling"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

var log = logrus.WithField("module", "v1")

// JsonUint64 is a helper to make both string and integer formatted uint64 numbers unmarshal correctly.
type JsonUint64 uint64

func (v *JsonUint64) UnmarshalJSON(b []byte) error {
	x := string(b)
	x = strings.TrimSpace(x)
	if len(x) >= 2 && x[0] == '"' && x[len(x)-1] == '"' {
		x = x[1 : len(x)-1]
	}
	d, err := strconv.ParseUint(x, 0, 64)
	if err != nil {
		return err
	}
	*v = JsonUint64(d)
	return nil
}

type V1HTTPClient struct {
	api    *sling.Sling
	client *http.Client
}

func (s *V1HTTPClient) GetVersion() (string, error) {
	path := "eth/v1/node/version"
	type versionResponse struct {
		Data struct {
			Version string `json:"version,omitempty"`
		} `json:"data,omitempty"`
	}
	response := new(versionResponse)
	_, err := s.api.New().Get(path).ReceiveSuccess(response)
	if err != nil {
		return "", err
	}
	return response.Data.Version, nil
}

func (s *V1HTTPClient) GetGenesisTime() (int64, error) {
	path := "eth/v1/beacon/genesis"
	type genesisResponse struct {
		Data struct {
			GenesisTime JsonUint64 `json:"genesis_time,omitempty"`
		} `json:"data,omitempty"`
	}
	response := new(genesisResponse)
	_, err := s.api.New().Get(path).ReceiveSuccess(response)
	if err != nil {
		return 0, err
	}
	return int64(response.Data.GenesisTime), nil
}

func (s *V1HTTPClient) GetPeerCount() (int64, error) {
	path := "eth/v1/node/peers"
	type peersResponse struct {
		Data []struct {
		} `json:"data,omitempty"`
	}
	response := new(peersResponse)
	_, err := s.api.New().Get(path).ReceiveSuccess(response)
	if err != nil {
		return 0, err
	}
	return int64(len(response.Data)), nil
}

func (s *V1HTTPClient) GetAttestationsInPoolCount() (int64, error) {
	// TODO: There's an attestations pool endpoint, but it lists way too much.
	//       So much, that querying it a lot is similar to a self-induced DoS attack.
	return 0, beacon.NotImplemented
}

func (s *V1HTTPClient) GetSyncStatus() (bool, error) {
	path := "eth/v1/node/syncing"
	type syncingResponse struct {
		Data struct {
			SyncDistance JsonUint64 `json:"sync_distance,omitempty"`
		} `json:"data,omitempty"`
	}
	response := new(syncingResponse)
	_, err := s.api.New().Get(path).ReceiveSuccess(response)
	if err != nil {
		return false, err
	}
	return response.Data.SyncDistance != 0, nil
}

func (s *V1HTTPClient) GetChainHead() (*types.ChainHead, error) {

	typesChainHead := new(types.ChainHead)

	headRootPath := "eth/v1/beacon/blocks/head/root"
	type headRootType struct {
		Data struct {
			HeadBlockRoot string `json:"root,omitempty"`
		} `json:"data,omitempty"`
	}
	headRootResponse := new(headRootType)
	_, err := s.api.New().Get(headRootPath).ReceiveSuccess(headRootResponse)
	if err != nil {
		return nil, err
	}
	typesChainHead.HeadBlockRoot = headRootResponse.Data.HeadBlockRoot

	slot, err := s.getBlockSlot(typesChainHead.HeadBlockRoot)
	if err != nil {
		return nil, err
	}
	typesChainHead.HeadSlot = slot

	finalityCheckpointsPath := "eth/v1/beacon/states/head/finality_checkpoints"
	type finalityCheckpointsType struct {
		Data struct {
			Finalized struct {
				Root  string     `json:"root,omitempty"`
				Epoch JsonUint64 `json:"epoch,omitempty"`
			} `json:"finalized,omitempty"`
			Justified struct {
				Root  string     `json:"root,omitempty"`
				Epoch JsonUint64 `json:"epoch,omitempty"`
			} `json:"current_justified,omitempty"`
		} `json:"data,omitempty"`
	}
	finalityCheckpointsResponse := new(finalityCheckpointsType)
	_, err = s.api.New().Get(finalityCheckpointsPath).ReceiveSuccess(finalityCheckpointsResponse)
	if err != nil {
		return nil, err
	}
	typesChainHead.JustifiedBlockRoot = finalityCheckpointsResponse.Data.Justified.Root
	typesChainHead.JustifiedSlot, _ = s.startSlotOfEpoch(uint64(finalityCheckpointsResponse.Data.Justified.Epoch))
	typesChainHead.FinalizedBlockRoot = finalityCheckpointsResponse.Data.Finalized.Root
	typesChainHead.FinalizedSlot, _ = s.startSlotOfEpoch(uint64(finalityCheckpointsResponse.Data.Finalized.Epoch))
	return typesChainHead, nil
}

func (s *V1HTTPClient) SubscribeChainHeads() (beacon.ChainHeadSubscription, error) {
	sub := polling.NewChainHeadClientPoller(s)
	go sub.Start()

	return sub, nil
}

func New(httpClient *http.Client, baseURL string) *V1HTTPClient {
	return &V1HTTPClient{
		api:    sling.New().Client(httpClient).Base(baseURL),
		client: httpClient,
	}
}

func (s *V1HTTPClient) getBlockSlot(blockId string) (uint64, error) {
	blockHeaderPath := fmt.Sprintf("eth/v1/beacon/headers/%s", blockId)
	type blockHeaderTypeResponse struct {
		Data struct {
			Header struct {
				Message struct {
					Slot JsonUint64 `json:"slot,omitempty"`
				} `json:"message,omitempty"`
			} `json:"header,omitempty"`
		} `json:"data,omitempty"`
	}
	blockHeaderResponse := new(blockHeaderTypeResponse)
	_, err := s.api.New().Get(blockHeaderPath).ReceiveSuccess(blockHeaderResponse)
	if err != nil {
		return 0, err
	}
	return uint64(blockHeaderResponse.Data.Header.Message.Slot), nil
}

func (s *V1HTTPClient) startSlotOfEpoch(epoch uint64) (uint64, error) {
	return epoch * 32, nil
}
