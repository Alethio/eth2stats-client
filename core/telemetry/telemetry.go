package telemetry

import (
	"context"
	"math"
	"time"

	proto "github.com/alethio/eth2stats-proto"
	"github.com/sirupsen/logrus"

	"github.com/alethio/eth2stats-client/beacon"
	metricsWatcher "github.com/alethio/eth2stats-client/watcher/metrics"
)

var log = logrus.WithField("module", "telemetry")

type Telemetry struct {
	service proto.TelemetryClient

	beaconClient     beacon.Client
	metricsWatcher   *metricsWatcher.Watcher
	contextWithToken func() context.Context

	data struct {
		Peers              *int64
		AttestationsInPool *int64
		Syncing            *bool
		MemoryUsage        *int64
	}
}

func New(service proto.TelemetryClient, beaconClient beacon.Client, watcher *metricsWatcher.Watcher, contextWithToken func() context.Context) *Telemetry {
	return &Telemetry{
		service:          service,
		beaconClient:     beaconClient,
		metricsWatcher:   watcher,
		contextWithToken: contextWithToken,
	}
}

func (t *Telemetry) Run() {
	for {
		log.Trace("sending telemetry")

		t.pollPeers()
		t.pollAttestations()
		t.pollSyncing()
		t.pollMemUsage()

		log.Trace("done sending telemetry")

		time.Sleep(PollingInterval)
	}
}

func (t *Telemetry) pollPeers() {
	peers, err := t.beaconClient.GetPeerCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("peers: %d", peers)

	if t.data.Peers == nil || *t.data.Peers != peers {
		t.data.Peers = &peers

		_, err := t.service.Peers(t.contextWithToken(), &proto.PeersRequest{Peers: peers})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (t *Telemetry) pollAttestations() {
	attestations, err := t.beaconClient.GetAttestationsInPoolCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("attestations: %d", attestations)

	if t.data.AttestationsInPool == nil || *t.data.AttestationsInPool != attestations {
		t.data.AttestationsInPool = &attestations

		_, err := t.service.Attestations(t.contextWithToken(), &proto.AttestationsRequest{AttestationsInPool: attestations})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (t *Telemetry) pollSyncing() {
	syncing, err := t.beaconClient.GetSyncStatus()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("node syncing: %t", syncing)

	if t.data.Syncing == nil || *t.data.Syncing != syncing {
		t.data.Syncing = &syncing

		_, err := t.service.Syncing(t.contextWithToken(), &proto.SyncingRequest{Syncing: syncing})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (t *Telemetry) pollMemUsage() {
	memUsagePointer := t.metricsWatcher.GetMemUsage()
	if memUsagePointer != nil {
		if t.data.MemoryUsage == nil || (math.Abs(float64(*t.data.MemoryUsage-*memUsagePointer)) > MemoryUsageThreshold) {
			t.data.MemoryUsage = memUsagePointer

			_, err := t.service.MemoryUsage(t.contextWithToken(), &proto.MemoryUsageRequest{MemoryUsage: *memUsagePointer})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
