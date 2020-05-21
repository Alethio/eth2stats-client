package prysm

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	prysmAPI "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/types"
)

var log = logrus.WithField("module", "prysm")
var ClientMaxReceiveMessageSize = 67108864

type Config struct {
	GRPCAddr string
	TLSCert  string
}

type PrysmGRPCClient struct {
	config Config

	beacon prysmAPI.BeaconChainClient
	node   prysmAPI.NodeClient
}

func New(config Config) *PrysmGRPCClient {
	log.Info("setting up beacon client connection")

	var dialOpt grpc.DialOption
	if config.TLSCert != "" {
		creds, err := credentials.NewClientTLSFromFile(config.TLSCert, "")
		if err != nil {
			log.Fatalf("failed to create tls credentials: %v", err)
		}
		dialOpt = grpc.WithTransportCredentials(creds)
	} else {
		dialOpt = grpc.WithInsecure()
		log.Warn("no tls certificate provided; will use insecure connection to beacon chain")
	}

	conn, err := grpc.Dial(
		config.GRPCAddr,
		dialOpt,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(ClientMaxReceiveMessageSize)),
	)
	if err != nil {
		log.Fatalf("failed to connect to prysm: %v", err)
	}

	beaconAPI := prysmAPI.NewBeaconChainClient(conn)
	nodeAPI := prysmAPI.NewNodeClient(conn)

	return &PrysmGRPCClient{
		config: config,
		beacon: beaconAPI,
		node:   nodeAPI,
	}
}

func (c *PrysmGRPCClient) GetVersion() (string, error) {
	version, err := c.node.GetVersion(context.Background(), &empty.Empty{})
	if err != nil {
		return "", fmt.Errorf("prysm: getting version: %s", err)
	}

	return version.GetVersion(), nil
}

func (c *PrysmGRPCClient) GetGenesisTime() (int64, error) {
	genesis, err := c.node.GetGenesis(context.Background(), &empty.Empty{})
	if err != nil {
		return 0, fmt.Errorf("prysm: getting genesis time: %s", err)
	}

	return genesis.GetGenesisTime().GetSeconds(), nil
}

func (c *PrysmGRPCClient) GetPeerCount() (int64, error) {
	peers, err := c.node.ListPeers(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return int64(len(peers.Peers)), nil
}

func (c *PrysmGRPCClient) GetAttestationsInPoolCount() (int64, error) {
	req := &prysmAPI.AttestationPoolRequest{
		PageSize: 1,
	}
	resp, err := c.beacon.AttestationPool(context.Background(), req)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	if resp == nil {
		return 0, beacon.NotImplemented
	}
	// fallback for not updated nodes
	if resp.TotalSize == 0 && len(resp.Attestations) > 0 {
		return int64(len(resp.Attestations)), nil
	}
	return int64(resp.TotalSize), nil
}

func (c *PrysmGRPCClient) GetSyncStatus() (bool, error) {
	sync, err := c.node.GetSyncStatus(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)
		return false, err
	}

	return sync.GetSyncing(), nil
}

func (c *PrysmGRPCClient) GetChainHead() (*types.ChainHead, error) {
	head, err := c.beacon.GetChainHead(context.Background(), &empty.Empty{})
	if err != nil {
		return nil, fmt.Errorf("prysm: getting chain head: %s", err)
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

func (c *PrysmGRPCClient) SubscribeChainHeads() (beacon.ChainHeadSubscription, error) {
	stream, err := c.beacon.StreamChainHead(context.Background(), &empty.Empty{})
	if err != nil {
		log.Error(err)

		return nil, err
	}

	sub := NewChainHeadSubscription()
	go sub.FeedFromStream(stream)

	return sub, nil
}
