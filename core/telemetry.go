package core

import (
	proto "github.com/alethio/eth2stats-proto"

	"github.com/alethio/eth2stats-client/beacon"
	metricsWatcher "github.com/alethio/eth2stats-client/watcher/metrics"
)

type Telemetry struct {
	beaconClient   beacon.Client
	metricsWatcher *metricsWatcher.Watcher
}

func (t *Telemetry) BuildRequest() *proto.TelemetryRequest {
	req := &proto.TelemetryRequest{}

	t.addPeers(req)
	t.addAttestations(req)
	t.addSyncing(req)
	t.addMemUsage(req)

	return req
}

func (t *Telemetry) addPeers(req *proto.TelemetryRequest) {
	peers, err := t.beaconClient.GetPeerCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("peers: %d", peers)

	req.Peers = peers
}

func (t *Telemetry) addAttestations(req *proto.TelemetryRequest) {
	attestations, err := t.beaconClient.GetAttestationsInPoolCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("attestations: %d", attestations)

	req.AttestationsInPool = attestations
}

func (t *Telemetry) addSyncing(req *proto.TelemetryRequest) {
	syncing, err := t.beaconClient.GetSyncStatus()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("node syncing: %t", syncing)

	req.Syncing = syncing
}

func (t *Telemetry) addMemUsage(req *proto.TelemetryRequest) {
	memUsagePointer := t.metricsWatcher.GetMemUsage()
	if memUsagePointer != nil {
		req.MemoryUsage = *memUsagePointer
	}
}
