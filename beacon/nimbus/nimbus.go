package nimbus

import (
	"fmt"
	"github.com/alethio/eth2stats-client/beacon/polling"
	"github.com/dghubble/sling"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/types"
)

var log = logrus.WithField("module", "nimbus")

type NimbusJsonHttp struct {
	api    *sling.Sling
	client *http.Client
}

type JsonReq struct {
	// example: {"method": "getNetworkPeers", "id": 123, "params": []}
	Method string        `json:"method"`
	Id     int           `json:"id"`
	Params []interface{} `json:"params"`
}

func (s *NimbusJsonHttp) JsonReq(dest interface{}, method string, params ...interface{}) error {
	paramsBase := make([]interface{}, 0)
	paramsBase = append(paramsBase, params...)
	_, err := s.api.New().Get("").Add("Content-Type", "application/json").BodyJSON(&JsonReq{
		Method: method,
		Id:     123,
		Params: paramsBase,
	}).ReceiveSuccess(dest)
	return err
}

type VersionResp struct {
	Result string      `json:"result"`
	Error  interface{} `json:"error"`
}

func (s *NimbusJsonHttp) GetVersion() (string, error) {
	var resp VersionResp
	err := s.JsonReq(&resp, "getNodeVersion")
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", fmt.Errorf("json err: %v", resp.Error)
	}
	return resp.Result, nil
}

func (s *NimbusJsonHttp) GetGenesisTime() (int64, error) {
	// TODO: harcoded goerli genesis time. Nimbus has no genesis time API
	return 1587981600, nil
}

type NetworkPeersResp struct {
	Result []string    `json:"result"`
	Error  interface{} `json:"error"`
}

func (s *NimbusJsonHttp) GetPeerCount() (int64, error) {
	var resp NetworkPeersResp
	err := s.JsonReq(&resp, "getNetworkPeers")
	if err != nil {
		return 0, err
	}
	if resp.Error != nil {
		return 0, fmt.Errorf("json err: %v", resp.Error)
	}
	return int64(len(resp.Result)), nil
}

func (s *NimbusJsonHttp) GetAttestationsInPoolCount() (int64, error) {
	return 0, beacon.NotImplemented
}

type SyncingResp struct {
	Result bool        `json:"result"`
	Error  interface{} `json:"error"`
}

func (s *NimbusJsonHttp) GetSyncStatus() (bool, error) {
	var resp SyncingResp
	err := s.JsonReq(&resp, "getSyncing")
	if err != nil {
		return false, err
	}
	if resp.Error != nil {
		return false, fmt.Errorf("json err: %v", resp.Error)
	}
	return resp.Result, nil
}

type ChainHeadResult struct {
	HeadSlot           uint64 `json:"head_slot"`
	HeadBlockRoot      string `json:"head_block_root"`
	FinalizedSlot      uint64 `json:"finalized_slot"`
	FinalizedBlockRoot string `json:"finalized_block_root"`
	JustifiedSlot      uint64 `json:"justified_slot"`
	JustifiedBlockRoot string `json:"justified_block_root"`
}

type ChainHeadResp struct {
	Result ChainHeadResult `json:"result"`
	Error  interface{}     `json:"error"`
}

func (s *NimbusJsonHttp) GetChainHead() (*types.ChainHead, error) {
	var resp ChainHeadResp
	err := s.JsonReq(&resp, "getChainHead")
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("json err: %v", resp.Error)
	}
	// no 0x in roots for nimbus, but that's ok
	typesChainHead := types.ChainHead(resp.Result)
	return &typesChainHead, nil
}

func (c *NimbusJsonHttp) SubscribeChainHeads() (beacon.ChainHeadSubscription, error) {
	sub := polling.NewChainHeadClientPoller(c)
	go sub.Start()

	return sub, nil
}

func New(httpClient *http.Client, baseURL string) *NimbusJsonHttp {
	return &NimbusJsonHttp{
		api:    sling.New().Client(httpClient).Base(baseURL),
		client: httpClient,
	}
}
