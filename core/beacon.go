package core

import (
	"github.com/alethio/eth2stats-client/beacon/lighthouse"
	"github.com/alethio/eth2stats-client/beacon/nimbus"
	"github.com/alethio/eth2stats-client/beacon/teku"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/beacon/prysm"
)

func initBeaconClient(nodeType, nodeAddr string) beacon.Client {
	// check GRPC clients
	switch nodeType {
	case "prysm":
		return prysm.New(prysm.Config{GRPCAddr: nodeAddr})
	default:
		break
	}

	// If not GRPC, then default to HTTP
	// FIXME: For clients with multiple supported types, enable the user to select the type.
	if !IsURL(nodeAddr) {
		log.Fatalf("invalid node URL: %s", nodeAddr)
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
	default:
		log.Fatalf("node type not recognized: %s", nodeType)
		return nil
	}
}

func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
