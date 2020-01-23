# Ethereum 2.0 Network Stats and Monitoring - CLI Client

> This is an intial POC release of the eth2stats network monitoring suite
> 
> Currently it supports Prysm and Lighthouse.
> More to come soon.

## Supported clients and protocols:
- [ ] Prysm
  - [x] GRPC
  - [ ] HTTP
- [ ] Lighthouse
  - [X] HTTP
  - [ ] Websockets
- [ ] Artemis
- [ ] ...
  
## Current live deployments:

- [https://eth2stats.io/](https://eth2stats.io/) - Our main deployment. Will add new testnets to this as they arrive.

## Getting Started (Prysm Sapphire Testnet)

The first thing you should do is get a beacon chain client node running and connected to said beacon chain by joining the [Prysm Sapphire Testnet](https://prylabs.net/participate).

Then you can run the following Docker command to start sending stats to [eth2stats](https://sapphire.eth2stats.net).  
**Please update your `--eth2stats.node-name` arg before starting the cli tool.**
**Please update your `--data-folder` arg before starting the cli tool.**

```shell script
docker run -d --name eth2stats --restart always --network="host" \
      -v ~/eth2stats/data:/data \
      alethio/eth2stats-client:latest \
      run --v \
      --eth2stats.node-name="YourNode" \
      --data.folder="/data" \
      --eth2stats.addr="grpc.sapphire.eth2stats.io:443" --eth2stats.tls=true \
      --beacon.type="prysm" --beacon.addr="localhost:4000"
```

You should now be able to see your node and it's stats on [eth2stats](https://sapphire.eth2stats.net).

If you want to see your beacon node client's memory usage as well, make sure you have metrics enabled in Prysm and add this cli argument, pointing at the right host `--beacon.metrics-addr="http://localhost:8080/metrics"`.

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
```shell script
./eth2stats-client run \
                   --eth2stats.node-name="YourNode" \
                   --eth2stats.addr="grpc.sapphire.eth2stats.io:443" --eth2stats.tls=true \
                   --beacon.type="prysm" --beacon.addr="localhost:4000"
```

If you want to see your beacon node client's memory usage as well, make sure you have metrics enabled in Prysm and add this cli argument, pointing at the right host `--beacon.metrics-addr="http://localhost:8080/metrics"`.
