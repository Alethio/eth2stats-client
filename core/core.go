package core

import (
	"time"

	proto "github.com/alethio/eth2stats-proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/core/telemetry"
	metricsWatcher "github.com/alethio/eth2stats-client/watcher/metrics"
)

var log = logrus.WithField("module", "core")

type Eth2statsConfig struct {
	Version    string
	ServerAddr string
	TLS        bool
	NodeName   string
}

type BeaconNodeConfig struct {
	Type        string
	Addr        string
	MetricsAddr string
}

type Config struct {
	Eth2stats  Eth2statsConfig
	BeaconNode BeaconNodeConfig
	DataFolder string
}

type Core struct {
	config Config
	token  string

	statsService     proto.Eth2StatsClient
	telemetryService proto.TelemetryClient

	beaconClient   beacon.Client
	metricsWatcher *metricsWatcher.Watcher
}

func New(config Config) *Core {
	c := Core{
		config:       config,
		beaconClient: initBeaconClient(config.BeaconNode.Type, config.BeaconNode.Addr),
	}

	c.initEth2statsClient()

	if config.BeaconNode.MetricsAddr != "" {
		c.metricsWatcher = metricsWatcher.New(metricsWatcher.Config{
			MetricsURL: config.BeaconNode.MetricsAddr,
		})
		go c.metricsWatcher.Run()
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
	resp, err := c.statsService.Connect(c.contextWithToken(), &proto.ConnectRequest{
		Name:             c.config.Eth2stats.NodeName,
		Version:          version,
		GenesisTime:      genesisTime,
		Eth2StatsVersion: c.config.Eth2stats.Version,
	}, grpc.WaitForReady(true))
	if err != nil {
		log.Fatal(err)
	}

	c.updateToken(resp.Token)

	log.Info("getting chain head for initial feed")
	head, err := c.beaconClient.GetChainHead()
	if err != nil {
		log.Fatal(err)
	}
	log.WithField("headSlot", head.HeadSlot).Info("got chain head")

	_, err = c.statsService.ChainHead(c.contextWithToken(), &proto.ChainHeadRequest{
		HeadSlot:           head.HeadSlot,
		HeadBlockRoot:      head.HeadBlockRoot,
		FinalizedSlot:      head.FinalizedSlot,
		FinalizedBlockRoot: head.FinalizedBlockRoot,
		JustifiedSlot:      head.JustifiedSlot,
		JustifiedBlockRoot: head.JustifiedBlockRoot,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Info("successfully connected to eth2stats server")
}

func (c *Core) watchNewHeads() {
	for {
		log.Info("setting up chain heads subscription")
		sub, err := c.beaconClient.SubscribeChainHeads()
		if err != nil {
			log.Fatal(err)
		}

		limiter := rate.NewLimiter(1, 1)

		for msg := range sub.Channel() {
			if limiter.Allow() {
				_, err := c.statsService.ChainHead(c.contextWithToken(), &proto.ChainHeadRequest{
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
			} else {
				log.Warn("ChainHead request was skipped due to rate limiting")
			}
		}

		log.Warn("chain heads subscription closed")
	}
}

func (c *Core) sendHeartbeat() {
	for range time.Tick(HeartbeatInterval) {
		log.Trace("sending heartbeat")

		_, err := c.statsService.Heartbeat(c.contextWithToken(), &proto.HeartbeatRequest{})
		if err != nil {
			log.Fatal(err)

			continue
		}
		log.Trace("done sending heartbeat")
	}
}

func (c *Core) Run() {
	c.connectToServer()
	go c.sendHeartbeat()
	go c.watchNewHeads()

	t := telemetry.New(c.telemetryService, c.beaconClient, c.metricsWatcher, c.contextWithToken)
	go t.Run()
}

func (c *Core) Close() {
	log.Info("Got stop signal")
}
