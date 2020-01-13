# Ethereum 2.0 Network Stats and Monitoring - CLI Client

> This is an intial POC release of the eth2stats network monitoring suite
> 
> Currently it supports just Prysm and their public testnet.
> More to come soon.

## Supported clients and protocols:
- [ ] Prysm
  - [x] GRPC
  - [ ] HTTP
- [ ] ...
  
## Current live deployments:

- [https://sapphire.eth2stats.net/](https://sapphire.eth2stats.net/) - [Prysm Sapphire Testnet](https://prylabs.net/participate)

## Getting Started (Prysm Sapphire Testnet)

The first thing you should do is get a beacon chain client node running and connected to said beacon chain by joining the [Prysm Sapphire Testnet](https://prylabs.net/participate).

Then you can run the following Docker command to start sending stats to [eth2stats](https://sapphire.eth2stats.net).  
**Please update your `--eth2stats.node-name` arg before starting the cli tool.**
**Please update your `--data-folder` arg before starting the cli tool.**

```shell script
docker run -d --name eth2stats --restart always --network="host" \
    alethio/eth2stats-client:latest \
    run --v \
    -v ~/eth2stats/data:/data \
    --eth2stats.node-name="YourNode" \ 
    --data.folder="/data" \
    --eth2stats.addr="grpc.sapphire.eth2stats.net:443" --eth2stats.tls=true \
    --beacon.type="prysm" --beacon.addr="localhost:4000"
```

You should now be able to see your node and it's stats on [eth2stats](https://sapphire.eth2stats.net).

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
                   --data.folder="/data" \
                   --eth2stats.node-name="YourNode" \
                   --eth2stats.addr="grpc.sapphire.eth2stats.net:443" --eth2stats.tls=true \
                   --beacon.type="prysm" --beacon.addr="localhost:4000"
```
