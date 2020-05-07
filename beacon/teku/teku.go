package teku

import (
	"fmt"
	"github.com/alethio/eth2stats-client/beacon/polling"
	"net/http"
	"strconv"

	"github.com/dghubble/sling"
	"github.com/sirupsen/logrus"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/types"
)

var log = logrus.WithField("module", "teku")

type TekuHTTPClient struct {
	api    *sling.Sling
	client *http.Client
}

func (s *TekuHTTPClient) GetVersion() (string, error) {
	path := fmt.Sprintf("node/version")
	version := new(string)
	_, err := s.api.New().Get(path).ReceiveSuccess(version)
	if err != nil {
		return "", err
	}
	return *version, nil
}

func (s *TekuHTTPClient) GetGenesisTime() (int64, error) {
	// node/genesis_time instead of beacon/genesis_time like lighthouse.
	path := fmt.Sprintf("node/genesis_time")
	genesis := new(string)
	_, err := s.api.New().Get(path).ReceiveSuccess(genesis)
	if err != nil {
		return 0, err
	}
	genesisTime, err := strconv.ParseInt(*genesis, 0, 64)
	if err != nil {
		return 0, err
	}
	return genesisTime, nil
}

func (s *TekuHTTPClient) GetPeerCount() (int64, error) {
	// Teku also has a `network/peers` endpoint like lighthouse, but this is more efficient.
	path := fmt.Sprintf("network/peer_count")
	peerCount := new(int64)
	_, err := s.api.New().Get(path).ReceiveSuccess(peerCount)
	if err != nil {
		return 0, err
	}
	return *peerCount, nil
}

func (s *TekuHTTPClient) GetAttestationsInPoolCount() (int64, error) {
	return 0, beacon.NotImplemented
}

func (s *TekuHTTPClient) GetSyncStatus() (bool, error) {
	path := fmt.Sprintf("node/syncing")
	type syncStatus struct {
		Syncing bool `json:"syncing"`
		// Note: ignore "sync_status" field
	}
	status := new(syncStatus)
	_, err := s.api.New().Get(path).ReceiveSuccess(status)
	if err != nil {
		return false, err
	}
	return status.Syncing, nil
}

func (s *TekuHTTPClient) GetChainHead() (*types.ChainHead, error) {
	path := fmt.Sprintf("beacon/chainhead")
	type chainHead struct {
		// Slight difference from lighthouse, to be standardized in new API proposal.
		HeadSlot           uint64 `json:"head_slot"`
		HeadBlockRoot      string `json:"head_block_root"`
		FinalizedSlot      uint64 `json:"finalized_slot"`
		FinalizedBlockRoot string `json:"finalized_block_root"`
		JustifiedSlot      uint64 `json:"justified_slot"`
		JustifiedBlockRoot string `json:"justified_block_root"`
		// Note: some fields, like epochs and previous justified epoch, are ignored.
	}
	head := new(chainHead)
	_, err := s.api.New().Get(path).ReceiveSuccess(head)
	if err != nil {
		return nil, err
	}
	typesChainHead := types.ChainHead(*head)
	return &typesChainHead, nil
}

func (c *TekuHTTPClient) SubscribeChainHeads() (beacon.ChainHeadSubscription, error) {
	sub := polling.NewChainHeadClientPoller(c)
	go sub.Start()

	return sub, nil
}

func New(httpClient *http.Client, baseURL string) *TekuHTTPClient {
	return &TekuHTTPClient{
		api:    sling.New().Client(httpClient).Base(baseURL),
		client: httpClient,
	}
}
