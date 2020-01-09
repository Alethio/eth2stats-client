package core

import (
	"crypto/tls"

	proto "github.com/alethio/eth2stats-proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func initEth2statsClient(config Eth2statsConfig) proto.Eth2StatsClient {
	log.Info("setting up eth2stats server connection")

	var conn *grpc.ClientConn
	var err error

	if config.TLS {
		tlsConfig := &tls.Config{}
		conn, err = grpc.Dial(config.ServerAddr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		conn, err = grpc.Dial(config.ServerAddr, grpc.WithInsecure())
	}
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	return proto.NewEth2StatsClient(conn)
}
