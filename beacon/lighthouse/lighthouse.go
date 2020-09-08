package lighthouse

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

var log = logrus.WithField("module", "lighthouse")

type LighthouseHTTPClient struct {
	api    *sling.Sling
	client *http.Client
}

func (s *LighthouseHTTPClient) GetVersion() (string, error) {
	path := fmt.Sprintf("node/version")
	version := new(string)
	_, err := s.api.New().Get(path).ReceiveSuccess(version)
	if err != nil {
		return "", err
	}
	return *version, nil
}

func (s *LighthouseHTTPClient) GetGenesisTime() (int64, error) {
	path := fmt.Sprintf("beacon/genesis_time")
	genesis := new(int64)
	_, err := s.api.New().Get(path).ReceiveSuccess(genesis)
	if err != nil {
		return 0, err
	}
	return *genesis, nil
}

func (s *LighthouseHTTPClient) GetPeerCount() (int64, error) {
	path := fmt.Sprintf("network/peers")
	peers := new([]string)
	_, err := s.api.New().Get(path).ReceiveSuccess(peers)
	if err != nil {
		return 0, err
	}
	return int64(len(*peers)), nil
}

func (s *LighthouseHTTPClient) GetAttestationsInPoolCount() (int64, error) {
	return 0, beacon.NotImplemented
}

func (s *LighthouseHTTPClient) GetSyncStatus() (bool, error) {
	return false, beacon.NotImplemented
}

func (s *LighthouseHTTPClient) GetChainHead() (*types.ChainHead, error) {
	path := fmt.Sprintf("beacon/head")
	type chainHead struct {
		HeadSlot           string `json:"slot"`
		HeadBlockRoot      string `json:"block_root"`
		FinalizedSlot      string `json:"finalized_slot"`
		FinalizedBlockRoot string `json:"finalized_block_root"`
		JustifiedSlot      string `json:"justified_slot"`
		JustifiedBlockRoot string `json:"justified_block_root"`
	}

	head := new(chainHead)
	_, err := s.api.New().Get(path).ReceiveSuccess(head)
	if err != nil {
		return nil, err
	}
	headSlot, err := strconv.ParseUint(head.HeadSlot, 0, 64)
	if err != nil {
		return nil, err
	}
	finalizedSlot, err := strconv.ParseUint(head.FinalizedSlot, 0, 64)
	if err != nil {
		return nil, err
	}
	justifiedSlot, err := strconv.ParseUint(head.JustifiedSlot, 0, 64)
	if err != nil {
		return nil, err
	}
	// TODO this returns roots with 0x while prysm doesn't ... which one is the correct form?
	typesChainHead := types.ChainHead{
		HeadSlot:           headSlot,
		HeadBlockRoot:      head.HeadBlockRoot,
		FinalizedSlot:      finalizedSlot,
		FinalizedBlockRoot: head.FinalizedBlockRoot,
		JustifiedSlot:      justifiedSlot,
		JustifiedBlockRoot: head.JustifiedBlockRoot,
	}
	return &typesChainHead, nil
}

func (c *LighthouseHTTPClient) SubscribeChainHeads() (beacon.ChainHeadSubscription, error) {
	sub := polling.NewChainHeadClientPoller(c)
	go sub.Start()

	return sub, nil
}

func New(httpClient *http.Client, baseURL string) *LighthouseHTTPClient {
	return &LighthouseHTTPClient{
		api:    sling.New().Client(httpClient).Base(baseURL),
		client: httpClient,
	}
}
