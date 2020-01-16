package prysm

import (
	"context"
	"encoding/hex"

	"github.com/golang/protobuf/ptypes/empty"
	prysmAPI "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/types"
)

var log = logrus.WithField("module", "prysm")

type Config struct {
	GRPCAddr string
}

type BeaconClient struct {
	config Config

	beacon prysmAPI.BeaconChainClient
	node   prysmAPI.NodeClient
}

func New(config Config) *BeaconClient {
	log.Info("setting up beacon client connection")

	conn, err := grpc.Dial(config.GRPCAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	beaconAPI := prysmAPI.NewBeaconChainClient(conn)
	nodeAPI := prysmAPI.NewNodeClient(conn)

	return &BeaconClient{
		config: config,
		beacon: beaconAPI,
		node:   nodeAPI,
	}
}

func (c *BeaconClient) GetVersion() (string, error) {
	version, err := c.node.GetVersion(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)
		return "", err
	}

	return version.GetVersion(), nil
}

func (c *BeaconClient) GetGenesisTime() (int64, error) {
	genesis, err := c.node.GetGenesis(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return genesis.GetGenesisTime().GetSeconds(), nil
}

func (c *BeaconClient) GetPeerCount() (int64, error) {
	peers, err := c.node.ListPeers(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return int64(len(peers.Peers)), nil
}

func (c *BeaconClient) GetAttestationsInPoolCount() (int64, error) {
	attestations, err := c.beacon.AttestationPool(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return int64(len(attestations.Attestations)), nil
}

func (c *BeaconClient) GetSyncStatus() (bool, error) {
	sync, err := c.node.GetSyncStatus(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)
		return false, err
	}

	return sync.GetSyncing(), nil
}

func (c *BeaconClient) GetChainHead() (*types.ChainHead, error) {
	head, err := c.beacon.GetChainHead(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &types.ChainHead{
		HeadSlot:           head.HeadSlot,
		HeadBlockRoot:      hex.EncodeToString(head.HeadBlockRoot),
		FinalizedSlot:      head.FinalizedSlot,
		FinalizedBlockRoot: hex.EncodeToString(head.FinalizedBlockRoot),
		JustifiedSlot:      head.JustifiedSlot,
		JustifiedBlockRoot: hex.EncodeToString(head.JustifiedBlockRoot),
	}, nil
}

func (c *BeaconClient) SubscribeChainHeads() (beacon.ChainHeadSubscription, error) {
	stream, err := c.beacon.StreamChainHead(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)

		return nil, err
	}

	sub := NewChainHeadSubscription()
	go sub.FeedFromStream(stream)

	return sub, nil
}
