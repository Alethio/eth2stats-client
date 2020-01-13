package core

import (
	"time"

	proto "github.com/alethio/eth2stats-proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/alethio/eth2stats-client/beacon"
)

var log = logrus.WithField("module", "core")

type Eth2statsConfig struct {
	ServerAddr string
	TLS        bool
	NodeName   string
}

type Config struct {
	Eth2stats      Eth2statsConfig
	BeaconNodeType string
	BeaconNodeAddr string
	DataFolder     string
}

type Core struct {
	config Config

	stats        proto.Eth2StatsClient
	beaconClient beacon.Client

	token string

	heartbeatActive bool
	heartbeatStop   chan bool

	newHeadsWatchActive bool
	newHeadsWatchStop   chan bool

	restartChan chan bool
	stopChan    chan bool
}

func New(config Config) *Core {
	c := Core{
		config:       config,
		stats:        initEth2statsClient(config.Eth2stats),
		beaconClient: initBeaconClient(config.BeaconNodeType, config.BeaconNodeAddr),
		restartChan:  make(chan bool),
		stopChan:     make(chan bool),
	}

	err := c.searchToken()
	if err != nil {
		log.Fatal(err)
	}

	return &c
}

func (c *Core) connectToServer() {
	log.Info("getting beacon client version")
	version, err := c.beaconClient.GetVersion()
	if err != nil {
		log.Fatal(err)
	}

	log.WithField("version", version).Info("got beacon client version")

	log.Info("getting beacon client genesis time")
	genesisTime, err := c.beaconClient.GetGenesisTime()
	if err != nil {
		log.Fatal(err)
	}

	log.WithField("genesisTime", genesisTime).Info("got beacon client genesis time")

	log.Info("awaiting connection to eth2stats server")
	resp, err := c.stats.Connect(c.contextWithToken(), &proto.ConnectRequest{
		Name:        c.config.Eth2stats.NodeName,
		Version:     version,
		GenesisTime: genesisTime,
	}, grpc.WaitForReady(true))
	if err != nil {
		log.Fatal(err)
	}

	c.updateToken(resp.Token)

	log.Info("successfully connected to eth2stats server")
}

func (c *Core) watchNewHeads() {
	for {
		log.Info("setting up chain heads subscription")
		sub, err := c.beaconClient.SubscribeChainHeads()
		if err != nil {
			log.Fatal(err)
		}

		for msg := range sub.Channel() {
			_, err := c.stats.ChainHead(c.contextWithToken(), &proto.ChainHeadRequest{
				HeadSlot:           msg.HeadSlot,
				HeadBlockRoot:      msg.HeadBlockRoot,
				FinalizedSlot:      msg.FinalizedSlot,
				FinalizedBlockRoot: msg.FinalizedBlockRoot,
				JustifiedSlot:      msg.JustifiedSlot,
				JustifiedBlockRoot: msg.JustifiedBlockRoot,
			})
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Warn("chain heads subscription closed")
	}
}

func (c *Core) sendHeartbeat() {
	for range time.Tick(HeartbeatInterval) {
		log.Trace("sending heartbeat")

		_, err := c.stats.Heartbeat(c.contextWithToken(), &proto.HeartbeatRequest{})
		if err != nil {
			log.Fatal(err)

			continue
		}
		log.Trace("done sending heartbeat")
	}
}

func (c *Core) sendTelemetry() {
	for {
		log.Trace("sending telemetry")

		peers, err := c.beaconClient.GetPeerCount()
		if err != nil {
			log.Fatal(err)
		}
		log.Tracef("peers: %d", peers)

		syncing, err := c.beaconClient.GetSyncStatus()
		if err != nil {
			log.Fatal(err)
		}
		log.Tracef("node syncing: %s", syncing)

		_, err = c.stats.Telemetry(c.contextWithToken(), &proto.TelemetryRequest{
			Peers:   peers,
			Syncing: syncing,
		})
		if err != nil {
			log.Fatal(err)

			continue
		}
		log.Trace("done sending telemetry")
		time.Sleep(TelemetryInterval)
	}
}

func (c *Core) Run() {
	c.connectToServer()
	go c.sendHeartbeat()
	go c.watchNewHeads()
	go c.sendTelemetry()
}

func (c *Core) Close() {
	log.Info("Got stop signal")
}
