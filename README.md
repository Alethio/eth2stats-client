# Ethereum 2.0 Network Stats and Monitoring - CLI Client

> This is an intial POC release of the eth2stats network monitoring suite
> 
> Currently it supports Prysm and Lighthouse.
> More to come soon.

## Supported clients and protocols:

| Client     | Supported | Protocols | Supported features                 |
|------------|-----------|-----------|------------------------------------|
| Prysm      | ✅         | GRPC      | Version, head, sync stats, memory |
| Lighthouse | ✅         | HTTP      | Version, head                      |
| Teku       | ✅         | HTTP      | Version, head, memory             |
| Lodestar   |           |           |                                    |
| Nimbus     |           |           |                                    |
| Trinity    |           |           |                                    |

  
## Current live deployments:

- [https://eth2stats.io/](https://eth2stats.io/) - Our main deployment. Will add new testnets to this as they arrive.

## Getting Started

The following section uses Docker to run. If you want to build from source go [here](#building-from-source).

The most important variable to change is **`--eth2stats.node-name`** which will define what name your node has on [eth2stats](https://eth2stats.io).


###  Prysm Sapphire Testnet

The first thing you should do is get a beacon chain client node running and connected to said beacon chain by joining the [Prysm Sapphire Testnet](https://prylabs.net/participate).

You can then get eth2stats sending data by running the following command:

```shell script
docker run -d --name eth2stats --restart always --network="host" \
      -v ~/eth2stats/data:/data \
      alethio/eth2stats-client:latest \
      run --v \
      --eth2stats.node-name="YourPrysmNode" \
      --data.folder="/data" \
      --eth2stats.addr="grpc.sapphire.eth2stats.io:443" --eth2stats.tls=true \
      --beacon.type="prysm" --beacon.addr="localhost:4000"
```

If you want to see your beacon node client's memory usage as well, make sure you have metrics enabled in Prysm and add this cli argument, pointing at the right host `--beacon.metrics-addr="http://localhost:8080/metrics"`.

### Lighthouse Testnet
The first thing you should do is get a beacon chain client node running and connected to said beacon chain by joining the [Lighthouse Testnet](https://lighthouse-book.sigmaprime.io/become-a-validator.html).

You can then get eth2stats sending data by running the following command: 

```shell script
docker run -d --name eth2stats --restart always --network="host" \
      -v ~/eth2stats/data:/data \
      alethio/eth2stats-client:latest \
      run --v \
      --eth2stats.node-name="YourLighthouseNode" \
      --data.folder="/data" \
      --eth2stats.addr="grpc.summer.eth2stats.io:443" --eth2stats.tls=true \
      --beacon.type="lighthouse" --beacon.addr="http://localhost:5052"
```

You should now be able to see your node and it's stats on [eth2stats](https://eth2stats.io).


## Building from source
### Prerequisites
- a working Golang environment (tested with go v1.13.5)
    - requires go modules (>=go v1.11)

### Step-by-step
**Clone the repo**
```shell script
git clone git@github.com:Alethio/eth2stats-client.git
cd eth2stats-client
```

**Build the executable**

We are using go modules, so it will automatically download the dependencies
```shell script
make build
```

**Run**

Example for Prysm:
```shell script
./eth2stats-client run \
                   --eth2stats.node-name="YourNode" \
                   --eth2stats.addr="grpc.sapphire.eth2stats.io:443" --eth2stats.tls=true \
                   --beacon.type="prysm" --beacon.addr="localhost:4000"
```

Example for Lighthouse:
```shell script
./eth2stats-client run \
                   --eth2stats.node-name="YourNode" \
                   --eth2stats.addr="grpc.summer.eth2stats.io:443" --eth2stats.tls=true \
                   --beacon.type="lighthouse" --beacon.addr="localhost:5052"
```

Example for Teku:
```shell script
./eth2stats-client run \
                   --eth2stats.node-name="YourNode" \
                   --eth2stats.addr="grpc.summer.eth2stats.io:443" --eth2stats.tls=true \
                   --beacon.type="teku" --beacon.addr="localhost:5051"
```

#### Memory usage metrics

If you want to see your beacon node client's memory usage as well, make sure you have metrics enabled and add this cli argument, pointing at the right host `--beacon.metrics-addr="http://127.0.0.1:8080/metrics"`.

Default metrics endpoints of supported clients:
- Lighthouse: `127.0.0.1:5052/metrics` (under regular http API address and port), currently not supporting the memory metric.
- Teku: `127.0.0.1:8008/metrics` (using `--metrics-enabled=true` in Teku options)
- Prysm: `127.0.0.1:8080/metrics`, monitoring enabled by default.

The `process_resident_memory_bytes` gauge is extracted from the Prometheus metrics endpoint.
