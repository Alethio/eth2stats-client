package core

import (
	"github.com/alethio/eth2stats-client/beacon/lodestar"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/beacon/lighthouse"
	"github.com/alethio/eth2stats-client/beacon/nimbus"
	"github.com/alethio/eth2stats-client/beacon/prysm"
	"github.com/alethio/eth2stats-client/beacon/teku"
)

func initBeaconClient(nodeType, nodeAddr, nodeCert string) beacon.Client {
	// check GRPC clients
	switch nodeType {
	case "prysm":
		return prysm.New(prysm.Config{GRPCAddr: nodeAddr, TLSCert: nodeCert})
	default:
		break
	}

	// If not GRPC, then default to HTTP
	// FIXME: For clients with multiple supported types, enable the user to select the type.
	if !IsURL(nodeAddr) {
		log.Fatalf("invalid node URL: %s", nodeAddr)
	}
	// Custom TLS certificates have only been tested with GRPC connections.
	// Fail early here to avoid mistakes. This can be implemented later if needed.
	if nodeCert != "" {
		log.Fatal("custom TLS certificates are currently only supported for GRPC connections")
	}
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 15 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 15 * time.Second,
	}
	var httpClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	switch nodeType {
	case "lighthouse":
		return lighthouse.New(httpClient, nodeAddr)
	case "teku":
		return teku.New(httpClient, nodeAddr)
	case "nimbus":
		return nimbus.New(httpClient, nodeAddr)
	case "lodestar":
		return lodestar.New(httpClient, nodeAddr)
	default:
		log.Fatalf("node type not recognized: %s", nodeType)
		return nil
	}
}

func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
