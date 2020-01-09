package core

import (
	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/beacon/prysm"
)

func initBeaconClient(nodeType, nodeAddr string) beacon.Client {
	switch nodeType {
	case "prysm":
		return prysm.New(prysm.Config{GRPCAddr: nodeAddr})
	default:
		log.Fatal("Node type not recognized.")
		return nil
	}
}
