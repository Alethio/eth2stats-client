package lighthouse

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/dghubble/sling"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/types"
)

type LighthouseClient struct {
	api    *sling.Sling
	client *http.Client
}

func (s *LighthouseClient) Get(path string) ([]byte, error) {
	req, err := s.api.New().Get(path).Request()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *LighthouseClient) GetVersion() (string, error) {
	path := fmt.Sprintf("node/version")
	body, err := s.Get(path)

	return string(body), err
}

func (s *LighthouseClient) GetGenesisTime() (int64, error) {
	path := fmt.Sprintf("beacon/genesis_time")
	body, err := s.Get(path)
	if err != nil {
		return 0, err
	}
	timestamp, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, err
	}
	return timestamp, nil
}

func (s *LighthouseClient) GetPeerCount() (int64, error) {
	path := fmt.Sprintf("network/peers")
	peers := new([]string)
	_, err := s.api.New().Get(path).ReceiveSuccess(peers)
	if err != nil {
		return 0, err
	}
	return int64(len(*peers)), nil
}

func (s *LighthouseClient) GetAttestationsInPoolCount() (int64, error) {
	return 0, beacon.NotAvailable
}

func (s *LighthouseClient) GetSyncStatus() (bool, error) {
	return false, beacon.NotAvailable
}

func (s *LighthouseClient) GetChainHead() (*types.ChainHead, error) {
	path := fmt.Sprintf("beacon/head")
	type chainHead struct {
		HeadSlot           uint64 `json:"slot"`
		HeadBlockRoot      string `json:"block_root"`
		FinalizedSlot      uint64 `json:"finalized_slot"`
		FinalizedBlockRoot string `json:"finalized_block_root"`
		JustifiedSlot      uint64 `json:"justified_slot"`
		JustifiedBlockRoot string `json:"justified_block_root"`
	}

	head := new(chainHead)
	_, err := s.api.New().Get(path).ReceiveSuccess(head)
	if err != nil {
		return nil, err
	}
	// TODO this returns roots with 0x while prysm doesn't ... which one is the correct form?
	typesChainHead := types.ChainHead(*head)
	return &typesChainHead, nil
}

func (c *LighthouseClient) SubscribeChainHeads() (beacon.ChainHeadSubscription, error) {
	sub := NewChainHeadSubscription(c)
	go sub.Start()

	return sub, nil
}

func New(httpClient *http.Client, baseURL string) *LighthouseClient {
	return &LighthouseClient{
		api:    sling.New().Client(httpClient).Base(baseURL),
		client: httpClient,
	}
}
