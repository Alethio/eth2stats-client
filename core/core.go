package core

import (
	"context"
	"fmt"
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
	TLSCert     string
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
		beaconClient: initBeaconClient(config.BeaconNode.Type, config.BeaconNode.Addr, config.BeaconNode.TLSCert),
	}

	c.initEth2statsClient()

	if config.BeaconNode.MetricsAddr != "" {
		c.metricsWatcher = metricsWatcher.New(metricsWatcher.Config{
			MetricsURL: config.BeaconNode.MetricsAddr,
		})
	}

	err := c.searchToken()
	if err != nil {
		log.Fatalf("loading auth token", err)
	}

	return &c
}

func (c *Core) connectToServer() error {
	log.Info("getting beacon client version")
	version, err := c.beaconClient.GetVersion()
	if err != nil {
		return err
	}

	log.WithField("version", version).Info("got beacon client version")

	log.Info("getting beacon client genesis time")
	genesisTime, err := c.beaconClient.GetGenesisTime()
	if err != nil {
		return err
	}

	log.WithField("genesisTime", genesisTime).Info("beacon client genesis time")

	log.Info("awaiting connection to eth2stats server")
	resp, err := c.statsService.Connect(c.contextWithToken(), &proto.ConnectRequest{
		Name:             c.config.Eth2stats.NodeName,
		Version:          version,
		GenesisTime:      genesisTime,
		Eth2StatsVersion: c.config.Eth2stats.Version,
	}, grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("eth2stats: failed to connect: %s", err)
	}

	c.updateToken(resp.Token)

	log.Info("getting chain head for initial feed")
	head, err := c.beaconClient.GetChainHead()
	if err != nil {
		return err
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
		log.Fatalf("sending chain head: %s", err)
	}

	log.Info("successfully connected to eth2stats server")
	return nil
}

func (c *Core) watchNewHeads(ctx context.Context) {
	for {
		log.Info("setting up chain heads subscription")
		sub, err := c.beaconClient.SubscribeChainHeads()
		if err != nil {
			// TODO handle gracefully
			log.Fatal(err)
		}
		go func() {
			<-ctx.Done()
			sub.Close()
		}()

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
					log.Fatalf("sending chain head: %s", err)
				}
			} else {
				log.Debug("ChainHead request was skipped due to rate limiting")
			}
		}

		log.Warn("chain heads subscription closed")
	}
}

func (c *Core) sendHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(HeartbeatInterval)
	for {
		select {
		case <-ticker.C:
			log.Trace("sending heartbeat")

			_, err := c.statsService.Heartbeat(c.contextWithToken(), &proto.HeartbeatRequest{})
			if err != nil {
				log.Fatalf("sending heartbeat: %s", err)

				continue
			}
			log.Trace("done sending heartbeat")
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (c *Core) Run(ctx context.Context) error {
	err := c.connectToServer()
	if err != nil {
		return fmt.Errorf("setting up: %s", err)
	}

	if c.metricsWatcher != nil {
		go c.metricsWatcher.Run(ctx)
	}

	go c.watchNewHeads(ctx)

	t := telemetry.New(c.telemetryService, c.beaconClient, c.metricsWatcher, c.contextWithToken)
	go t.Run(ctx)

	// block while sending heartbeat
	c.sendHeartbeat(ctx)
	return nil
}
