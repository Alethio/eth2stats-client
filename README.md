# Ethereum 2.0 Network Stats and Monitoring - CLI Client

> This is an intial POC release of the eth2stats network monitoring suite
> 
> It supports Prysm, Lighthouse, Teku, Nimbus and v1 of standardized api (Lodestar).
> Once the standard lands the client will be refactored to support just that.

## Supported clients and protocols:

| Client        | Supported | Protocols | Supported features                                   |
|---------------|-----------|-----------|------------------------------------------------------|
| Prysm         | ✅        | GRPC      | Version, head, sync stats, memory, attestation count |
| Lighthouse (v1)   | ✅        | HTTP      | Version, head, sync stats, memory                    |
| Teku          | ✅        | HTTP      | Version, head, sync stats, memory                    |
| Lodestar (v1) | ✅        | HTTP      | Version, head, sync stats, memory, attestation count |
| Nimbus        | ✅        | HTTP      | Version, head, sync stats, memory                    |
| Trinity       |          |           |                                                      |

  
## Current live deployments:

- [eth2stats.io/](https://eth2stats.io/) - Classic testnets:
  - [Medalla](https://eth2stats.io/medalla-testnet)
  - [Onyx](https://eth2stats.io/onyx-testnet)
  - [Altona](https://eth2stats.io/altona-testnet)
  - [Witti](https://eth2stats.io/witti-testnet)
  - And previously Schelesi and Sapphire testnets.
- [atttacknet.eth2.wtf](https://atttacknet.eth2.wtf)
- [zinken.eth2.wtf](https://atttacknet.eth2.wtf)


## Getting Started

The following section uses Docker to run. If you want to build from source go [here](#building-from-source).

The most important variable to change is **`--eth2stats.node-name`** which will define what name your node has on [eth2stats](https://eth2stats.io).


### Joining a Testnet

The first thing you should do is get a beacon node running and connected to your Eth2 network of choice.

The dashboard for the given testnet has a "Add your node" button. The client information is not always accurate however.
See below for options per client.

```shell script
docker run -d --name eth2stats --restart always --network="host" \
      -v ~/eth2stats/data:/data \
      alethio/eth2stats-client:latest \
      run --v \
      --eth2stats.node-name="YourPrysmNode" \
      --data.folder="/data" \
      --eth2stats.addr="grpc.sapphire.eth2stats.io:443" --eth2stats.tls=true \
      --beacon.type="changeme" --beacon.addr="changeme" --beacon.metrics-addr="changeme" # insert client-specific options here
```

### Client options

| Client version            | `--beacon.type`        | `--beacon.addr`           | `--beacon.metrics-addr`                   |
|---------------------------|------------------------|---------------------------|-------------------------------------------|
| Lighthouse v0.3.x         | `v1`  (standard API)   | `http://localhost:5052`   | `http://127.0.0.1:5054/metrics` (changed) |
| Lighthouse v0.2.x         | `lighthouse`           | `http://localhost:5052`   | `http://127.0.0.1:5052/metrics`           |
| Lodestar                  | `v1`  (standard API)   | `http://localhost:9596`   | `http://127.0.0.1:8008/metrics`           |
| Nimbus                    | `nimbus`               | `http://localhost:9190`   | `http://127.0.0.1:8008/metrics`           |
| Prysm                     | `prysm`                | `localhost:4000` (GRPC!)  | `http://127.0.0.1:8080/metrics`           |
| Teku                      | `teku`                 | `http://localhost:5051`   | `http://127.0.0.1:8008/metrics`           |

The metrics are only required if you want to see your beacon node client's memory usage on eth2stats.


### Securing your gRPC connection to the Beacon Chain

If your Beacon node uses a TLS connection for its GRPC endpoint you need to provide a valid certificate to `eth2stats-client` via the `--beacon.tls-cert` flag:

```shell script
docker run -d --name eth2stats --restart always --network="host" \
      -v ~/eth2stats/data:/data \
      ... # omitted for brevity
      --beacon.type="prysm" --beacon.addr="localhost:4000" --beacon.tls-cert "/data/cert.pem"
```

Have a look at Prysm's documentation to learn [how to start their Beacon Chain with enabled TLS](https://docs.prylabs.network/docs/prysm-usage/secure-grpc) and how to [generate and use self-signed certificates](https://docs.prylabs.network/docs/prysm-usage/secure-grpc#generating-self-signed-tls-certificates).

### Metrics

If you want to see your beacon node client's memory usage as well, make sure you have metrics enabled and add this cli argument,
 pointing at the right host, e.g. `--beacon.metrics-addr="http://127.0.0.1:8080/metrics"`.

Default metrics endpoints of supported clients:
- Lighthouse: `127.0.0.1:5054/metrics` (using `--metrics --metrics-address=127.0.0.1 --metrics-port=5054`)
- Teku: `127.0.0.1:8008/metrics` (using `--metrics-enabled=true` in Teku options)
- Prysm: `127.0.0.1:8080/metrics`, monitoring enabled by default.
- Nimbus: `127.0.0.1:8008/metrics` (using `--metrics --metrics-port=8008`)
- Lodestar: `127.0.0.1:8008/metrics` (configure with `"metrics": { "enabled": true, "serverPort": 8008}` in config JSON)

The `process_resident_memory_bytes` gauge is extracted from the Prometheus metrics endpoint.


## Building from source

### Prerequisites

- a working Golang environment (tested with go v1.14)
    - requires go modules (>=go v1.11)

### Step-by-step

**Clone the repo**
```shell script
git clone https://github.com/Alethio/eth2stats-client.git
cd eth2stats-client
```

**Build the executable**

We are using go modules, so it will automatically download the dependencies
```shell script
make build
```

**Run**

The `eth2stats-client` can run with `run` and flags as described per client.

Example for Lighthouse:
```shell script
./eth2stats-client run \
                   --eth2stats.node-name="YourNode" \
                   --eth2stats.addr="grpc.example.eth2stats.io:443" --eth2stats.tls=true \
                   --beacon.type="lighthouse" --beacon.addr="http://localhost:5052"
```

Note that since Prysm uses GRPC, the addr flag does not start with `http://`, unlike the others.
So it would be like `--beacon.addr="localhost:4000"`.

For the other clients, it is similar as lighthouse, except you replace the name.

Client names are `prysm`, `lighthouse`, `teku`, `nimbus`, `lodestar`. And `v1` for standard API option, which clients are all planning to adopt.
