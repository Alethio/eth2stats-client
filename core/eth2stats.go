package core

import (
	"crypto/tls"

	proto "github.com/alethio/eth2stats-proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (c *Core) initEth2statsClient() {
	log.Info("setting up eth2stats server connection")

	var conn *grpc.ClientConn
	var err error

	if c.config.Eth2stats.TLS {
		tlsConfig := &tls.Config{}
		conn, err = grpc.Dial(
			c.config.Eth2stats.ServerAddr,
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		)
	} else {
		conn, err = grpc.Dial(c.config.Eth2stats.ServerAddr,
			grpc.WithInsecure(),
		)
	}
	if err != nil {
		log.Fatalf("failed to connect to eth2stats: %v", err)
	}

	c.statsService = proto.NewEth2StatsClient(conn)
	c.telemetryService = proto.NewTelemetryClient(conn)
	c.validatorService = proto.NewValidatorClient(conn)
}
