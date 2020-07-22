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
)

var log = logrus.WithField("module", "v1")

type V1HTTPClient struct {
	api    *sling.Sling
	client *http.Client
}

func (s *V1HTTPClient) GetVersion() (string, error) {
	path := "v1/node/version"
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
	path := "v1/beacon/genesis"
	type genesisResponse struct {
		Data struct {
			GenesisTime string `json:"genesis_time,omitempty"`
		} `json:"data,omitempty"`
	}
	response := new(genesisResponse)
	_, err := s.api.New().Get(path).ReceiveSuccess(response)
	if err != nil {
		return 0, err
	}
	genesisTime, err := strconv.ParseInt(response.Data.GenesisTime, 10, 64)
	if err != nil {
		return 0, err
	}
	return genesisTime, nil
}

func (s *V1HTTPClient) GetPeerCount() (int64, error) {
	path := "v1/node/peers"
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
	path := "v1/beacon/pool/attestations"
	type attestationsResponse struct {
		Data []struct {
		} `json:"data,omitempty"`
	}
	response := new(attestationsResponse)
	_, err := s.api.New().Get(path).ReceiveSuccess(response)
	if err != nil {
		return 0, err
	}
	return int64(len(response.Data)), nil
}

func (s *V1HTTPClient) GetSyncStatus() (bool, error) {
	path := "v1/node/syncing"
	type syncingResponse struct {
		Data struct {
			SyncDistance string `json:"sync_distance,omitempty"`
		} `json:"data,omitempty"`
	}
	response := new(syncingResponse)
	_, err := s.api.New().Get(path).ReceiveSuccess(response)
	if err != nil {
		return false, err
	}
	syncDistance, err := strconv.ParseUint(response.Data.SyncDistance, 10, 64)
	if err != nil {
		return false, err
	}
	return syncDistance != 0, nil
}

func (s *V1HTTPClient) GetChainHead() (*types.ChainHead, error) {

	typesChainHead := new(types.ChainHead)

	headRootPath := "v1/beacon/blocks/head/root"
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

	finalityCheckpointsPath := "v1/beacon/state/head/finality_checkpoints"
	type finalityCheckpointsType struct {
		Data struct {
			Finalized struct {
				Root  string `json:"root,omitempty"`
				Epoch string `json:"epoch,omitempty"`
			} `json:"finalized,omitempty"`
			Justified struct {
				Root  string `json:"root,omitempty"`
				Epoch string `json:"epoch,omitempty"`
			} `json:"current_justified,omitempty"`
		} `json:"data,omitempty"`
	}
	finalityCheckpointsResponse := new(finalityCheckpointsType)
	_, err = s.api.New().Get(finalityCheckpointsPath).ReceiveSuccess(finalityCheckpointsResponse)
	if err != nil {
		return nil, err
	}
	typesChainHead.JustifiedBlockRoot = finalityCheckpointsResponse.Data.Justified.Root
	typesChainHead.JustifiedSlot, _ = s.startSlotOfEpoch(finalityCheckpointsResponse.Data.Justified.Epoch)
	typesChainHead.FinalizedBlockRoot = finalityCheckpointsResponse.Data.Finalized.Root
	typesChainHead.FinalizedSlot, _ = s.startSlotOfEpoch(finalityCheckpointsResponse.Data.Finalized.Epoch)
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
	blockHeaderPath := fmt.Sprintf("v1/beacon/headers/%s", blockId)
	type blockHeaderTypeResponse struct {
		Data struct {
			Header struct {
				Message struct {
					Slot string `json:"slot,omitempty"`
				} `json:"message,omitempty"`
			} `json:"header,omitempty"`
		} `json:"data,omitempty"`
	}
	blockHeaderResponse := new(blockHeaderTypeResponse)
	_, err := s.api.New().Get(blockHeaderPath).ReceiveSuccess(blockHeaderResponse)
	if err != nil {
		return 0, err
	}
	slot, err := strconv.ParseUint(blockHeaderResponse.Data.Header.Message.Slot, 10, 64)
	if err != nil {
		return 0, err
	}
	return slot, nil
}

func (s *V1HTTPClient) startSlotOfEpoch(epochStr string) (uint64, error) {
	epoch, err := strconv.ParseUint(epochStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return epoch * 32, nil
}
