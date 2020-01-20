package core

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/beacon/lighthouse"
	"github.com/alethio/eth2stats-client/beacon/prysm"
)

func initBeaconClient(nodeType, nodeAddr string) beacon.Client {
	switch nodeType {
	case "prysm":
		return prysm.New(prysm.Config{GRPCAddr: nodeAddr})
	case "lighthouse":
		if !IsURL(nodeAddr) {
			log.Fatal("Invalid node URL.")
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
		return lighthouse.New(httpClient, nodeAddr)
	default:
		log.Fatal("Node type not recognized.")
		return nil
	}
}

func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
